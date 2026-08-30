import React, { useState, useRef, useEffect } from 'react'
import { useAuth, useToast } from '../App'
import Navbar from './Navbar'
import { Canvas, useFrame } from '@react-three/fiber'
import { Float, Stars } from '@react-three/drei'
import * as THREE from 'three'
import gsap from 'gsap'
import { LazyCanvas } from './CanvasWrapper'

// 3D Attack Visualization
function AttackVisualization({ isRunning, results }) {
  const groupRef = useRef()
  const particlesRef = useRef()
  const count = 200
  const positionsRef = useRef()

  // Initialize refs
  useEffect(() => {
    const pos = new Float32Array(count * 3)
    for (let i = 0; i < count; i++) {
      const radius = 2 + Math.random() * 6
      const angle = Math.random() * Math.PI * 2
      const height = (Math.random() - 0.5) * 8
      pos[i * 3] = Math.cos(angle) * radius
      pos[i * 3 + 1] = height
      pos[i * 3 + 2] = Math.sin(angle) * radius
    }
    positionsRef.current = pos
  }, [])

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    if (groupRef.current) {
      groupRef.current.rotation.y = t * 0.1
    }
    if (particlesRef.current && isRunning) {
      particlesRef.current.rotation.y = t * 0.5
      const positions = positionsRef.current
      for (let i = 0; i < count; i++) {
        const angle = Math.atan2(positions[i * 3 + 2], positions[i * 3]) + 0.01
        const radius = Math.sqrt(positions[i * 3] ** 2 + positions[i * 3 + 2] ** 2)
        positions[i * 3] = Math.cos(angle) * radius
        positions[i * 3 + 2] = Math.sin(angle) * radius
        positions[i * 3 + 1] += Math.sin(t * 3 + i) * 0.01
      }
      particlesRef.current.geometry.attributes.position.needsUpdate = true
    }
  })

  const blockedCount = results?.summary?.blocked || 0
  const vulnerableCount = results?.summary?.vulnerable || 0

  return (
    <group ref={groupRef}>
      {/* Central core */}
      <Float speed={2} rotationIntensity={0.2} floatIntensity={0.5}>
        <mesh>
          <icosahedronGeometry args={[1.5, 3]} />
          <meshPhysicalMaterial
            color={isRunning ? '#ef4444' : (vulnerableCount > 0 ? '#ef4444' : '#10b981')}
            roughness={0.1}
            metalness={0.8}
            transparent
            opacity={0.3}
            transmission={0.5}
          />
        </mesh>

        {/* Inner core */}
        <mesh scale={0.6}>
          <sphereGeometry args={[1, 32, 32]} />
          <meshBasicMaterial
            color={isRunning ? '#ef4444' : (vulnerableCount > 0 ? '#ef4444' : '#10b981')}
            transparent
            opacity={0.8}
          />
        </mesh>

        {/* Wireframe overlay */}
        <mesh>
          <icosahedronGeometry args={[1.52, 2]} />
          <meshBasicMaterial
            color={isRunning ? '#ef4444' : '#10b981'}
            wireframe
            transparent
            opacity={0.2}
          />
        </mesh>
      </Float>

      {/* Protective shield if no vulnerabilities */}
      {!isRunning && vulnerableCount === 0 && blockedCount > 0 && (
        <mesh rotation={[Math.PI / 2, 0, 0]}>
          <torusGeometry args={[3, 0.05, 16, 100]} />
          <meshBasicMaterial color="#10b981" transparent opacity={0.5} />
        </mesh>
      )}

      {/* Particles */}
      <points ref={particlesRef}>
        <bufferGeometry>
          <bufferAttribute
            attach="attributes-position"
            count={count}
            array={positionsRef.current}
            itemSize={3}
          />
        </bufferGeometry>
        <pointsMaterial
          size={0.06}
          color={isRunning ? '#ef4444' : (vulnerableCount > 0 ? '#ef4444' : '#10b981')}
          transparent
          opacity={0.6}
          sizeAttenuation
          blending={THREE.AdditiveBlending}
        />
      </points>

      {/* Stars */}
      <Stars radius={60} depth={30} count={1000} factor={4} saturation={0} fade speed={0.3} />
    </group>
  )
}

// Camera controller
function CameraController() {
  const { camera } = useThree()
  const mouseRef = useRef({ x: 0, y: 0 })
  const targetRef = useRef({ x: 0, y: 0 })

  useEffect(() => {
    const handleMouseMove = (event) => {
      mouseRef.current.x = (event.clientX / window.innerWidth) * 2 - 1
      mouseRef.current.y = -(event.clientY / window.innerHeight) * 2 + 1
    }
    window.addEventListener('mousemove', handleMouseMove)
    return () => window.removeEventListener('mousemove', handleMouseMove)
  }, [])

  useFrame(() => {
    targetRef.current.x += (mouseRef.current.x * 2 - targetRef.current.x) * 0.02
    targetRef.current.y += (mouseRef.current.y * 1 - targetRef.current.y) * 0.02

    camera.position.x += (targetRef.current.x - camera.position.x) * 0.03
    camera.position.y += (targetRef.current.y - camera.position.y) * 0.03
    camera.lookAt(0, 0, 0)
  })

  return null
}

import { useThree } from '@react-three/fiber'

export default function RedTeamPage() {
  const [running, setRunning] = useState(false)
  const [results, setResults] = useState(null)
  const { authFetch, isAuthenticated } = useAuth()
  const { showToast } = useToast()

  const runAttacks = async () => {
    setRunning(true)
    setResults(null)
    try {
      const res = await authFetch('/api/redteam/run', { method: 'POST' })
      const data = await res.json()

      if (!res.ok) {
        if (data.attacks) {
          setResults(data)
          showToast('Attacks completed with vulnerabilities', 'error')
        } else {
          throw new Error(data.error || 'Failed to run attacks')
        }
      } else {
        setResults(data)
        showToast('All attacks blocked successfully!', 'success')
      }
    } catch (err) {
      showToast(err.message, 'error')
    } finally {
      setRunning(false)
    }
  }

  if (!isAuthenticated) {
    return (
      <div className="merchants-page">
        <Navbar />
        <div className="container" style={{ display: 'flex', justifyContent: 'center', paddingTop: '80px' }}>
          <div className="onboarding-card" style={{ maxWidth: '420px', width: '100%', textAlign: 'center' }}>
            <h2 className="onboarding-title">Admin Access Required</h2>
            <p className="onboarding-subtitle">Please login from the Merchants page first.</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="merchants-page">
      <Navbar />
      <div className="container">
        <div className="page-header">
          <div className="page-header-row">
            <div>
              <h1 className="page-title">Purchase Safety Test Lab</h1>
              <p className="page-subtitle">
                Validate how your Merchant MCP handles unusual, adversarial, and injected purchase requests
              </p>
            </div>
            <div>
              <button
                className="btn btn-primary"
                onClick={runAttacks}
                disabled={running}
                style={{
                  background: 'linear-gradient(135deg, #ef4444, #b91c1c)',
                  borderColor: 'rgba(239, 68, 68, 0.4)',
                  boxShadow: '0 4px 16px rgba(239, 68, 68, 0.3)'
                }}
              >
                {running ? (
                  <>

                    <span className="spinner" style={{ width: 16, height: 16, borderTopColor: '#fff' }} />
                    Running Simulation...
                  </>
                ) : (
                  <>
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>
                    Launch Attack Suite
                  </>
                )}
              </button>
            </div>
          </div>
        </div>

        {/* 3D Visualization */}
        <div style={{ width: '100%', height: '400px', marginBottom: '40px', borderRadius: '20px', overflow: 'hidden' }}>
          <LazyCanvas
            threshold={0.1}
            rootMargin='100px'
            canvasProps={{
              camera: { position: [0, 0, 12], fov: 50, near: 0.1, far: 100 }
            }}
          >
            <ambientLight intensity={0.3} />
            <pointLight position={[5, 5, 5]} intensity={1} color="#ef4444" />
            <pointLight position={[-5, -3, 3]} intensity={0.6} color="#10b981" />

            <AttackVisualization isRunning={running} results={results} />
            <CameraController />
          </LazyCanvas>
        </div>

        {!results && !running && (
          <div className="card" style={{ padding: '60px 24px', textAlign: 'center' }}>
            <div style={{ opacity: 0.5, marginBottom: '20px', color: 'var(--accent-red)' }}>
              <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5 12 2"></polygon><line x1="12" y1="22" x2="12" y2="15.5"></line><polyline points="22 8.5 12 15.5 2 8.5"></polyline><polyline points="2 15.5 12 8.5 22 15.5"></polyline><line x1="12" y1="2" x2="12" y2="8.5"></line></svg>
            </div>
            <h2 style={{ fontSize: '24px', marginBottom: '12px' }}>System Ready for Simulation</h2>
            <p style={{ color: 'var(--text-secondary)', maxWidth: '500px', margin: '0 auto' }}>
              Launch the automated test suite to simulate adversarial prompt injections, velocity abuse, and idempotency replays against the policy engine in real-time.
            </p>
          </div>
        )}

        {running && (
          <div className="card" style={{ padding: '60px 24px', textAlign: 'center' }}>
            <div className="spinner" style={{ width: 48, height: 48, margin: '0 auto 24px', borderTopColor: 'var(--accent-red)' }}></div>
            <h2 style={{ fontSize: '20px', marginBottom: '8px', color: 'var(--accent-red)' }}>Executing Attack Vectors...</h2>
            <p style={{ color: 'var(--text-secondary)' }}>Fuzzing MCP endpoints and triggering limit boundaries</p>
          </div>
        )}

        {results && !running && (
          <div>
            <div className="platform-tiles" style={{ marginBottom: '24px' }}>
              <div className="platform-tile" style={{ borderLeft: '4px solid #38bdf8' }}>
                <span style={{ fontSize: '24px', color: 'var(--text-primary)', fontWeight: 700 }}>{results.summary?.total || 0}</span>
                <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>Total Attacks</span>
              </div>
              <div className="platform-tile" style={{ borderLeft: '4px solid #22c55e' }}>
                <span style={{ fontSize: '24px', color: 'var(--text-primary)', fontWeight: 700 }}>{results.summary?.blocked || 0}</span>
                <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>Blocked Safely</span>
              </div>
              <div className="platform-tile" style={{ borderLeft: '4px solid #ef4444' }}>
                <span style={{ fontSize: '24px', color: 'var(--text-primary)', fontWeight: 700 }}>{results.summary?.vulnerable || 0}</span>
                <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>Vulnerabilities</span>
              </div>
            </div>

            <div className="card">
              <div style={{ overflowX: 'auto' }}>
                <table className="merchants-table">
                  <thead>
                    <tr>
                      <th style={{ width: '40px' }}>Result</th>
                      <th style={{ width: '250px' }}>Attack Vector</th>
                      <th>Description & Details</th>
                    </tr>
                  </thead>
                  <tbody>
                    {results.attacks?.map((attack, i) => (
                      <tr key={i} style={{ background: attack.passed ? 'rgba(239, 68, 68, 0.05)' : 'transparent' }}>
                        <td style={{ textAlign: 'center' }}>
                          {attack.passed ? (
                            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" color="var(--accent-red)"><circle cx="12" cy="12" r="10"/><path d="M15 9l-6 6M9 9l6 6"/></svg>
                          ) : (
                            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" color="var(--accent-green)"><path d="M22 11.08V12a10 10 0 11-5.93-9.14"/><path d="M22 4L12 14.01l-3-3"/></svg>
                          )}
                        </td>
                        <td style={{ fontWeight: 600, color: attack.passed ? 'var(--accent-red)' : 'var(--text-primary)' }}>
                          {attack.name}
                        </td>
                        <td>
                          <div style={{ marginBottom: '8px' }}>{attack.description}</div>
                          {attack.details && (
                            <code style={{
                              display: 'block',
                              fontSize: '11px',
                              color: 'var(--text-tertiary)',
                              background: 'rgba(0,0,0,0.2)',
                              padding: '8px',
                              borderRadius: '4px'
                            }}>
                              {attack.details}
                            </code>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}