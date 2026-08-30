import React, { useState, useRef, useEffect, useMemo } from 'react'
import { useToast } from '../App'
import { Canvas, useFrame } from '@react-three/fiber'
import { Float, Stars, Line } from '@react-three/drei'
import * as THREE from 'three'
import gsap from 'gsap'
import { LazyCanvas } from './CanvasWrapper'

// 3D Purchase Flow Visualization
function PurchaseFlowVisualization({ storeId, result, loading }) {
  const groupRef = useRef()
  const nodesRef = useRef({})
  const particlesRef = useRef()
  const particleCount = 100
  const particlePositionsRef = useMemo(() => new Float32Array(particleCount * 3), [])

  // Node positions for the flow
  const nodePositions = {
    buyer: [-6, 0, 0],
    mcp: [-2, 0, 0],
    gateway: [2, 0, 0],
    razorpay: [6, 0, 0],
  }

  const nodeColors = {
    buyer: '#a855f7',
    mcp: '#06b6d4',
    gateway: '#f59e0b',
    razorpay: '#ef4444',
  }

  const nodeLabels = {
    buyer: 'AI Buyer',
    mcp: 'Merchant MCP',
    gateway: 'Aegis Gateway',
    razorpay: 'Razorpay',
  }

  const flowPaths = [
    { from: 'buyer', to: 'mcp' },
    { from: 'mcp', to: 'gateway' },
    { from: 'gateway', to: 'razorpay' },
  ]

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    if (groupRef.current) {
      groupRef.current.rotation.y = t * 0.05
    }

    // Animate flow particles
    if (particlesRef.current) {
      const positions = particlePositionsRef.current
      const pathCount = flowPaths.length

      for (let i = 0; i < particleCount; i++) {
        const pathIndex = i % pathCount
        const path = flowPaths[pathIndex]
        const from = nodePositions[path.from]
        const to = nodePositions[path.to]
        const progress = ((t * 0.3 + i / particleCount) % 1)

        // Easing
        const eased = progress < 0.5
          ? 2 * progress * progress
          : 1 - Math.pow(-2 * progress + 2, 2) / 2

        const x = from[0] + (to[0] - from[0]) * eased
        const y = from[1] + (to[1] - from[1]) * eased + Math.sin(t * 2 + i) * 0.15
        const z = from[2] + (to[2] - from[2]) * eased

        positions[i * 3] = x
        positions[i * 3 + 1] = y
        positions[i * 3 + 2] = z
      }
      particlesRef.current.geometry.attributes.position.needsUpdate = true
    }

    // Pulse nodes based on state
    if (nodesRef.current.buyer) {
      const scale = 1 + Math.sin(t * 3) * 0.1
      nodesRef.current.buyer.scale.setScalar(scale)
    }
  })

  return (
    <group ref={groupRef}>
      {/* Flow particles */}
      <points ref={particlesRef}>
        <bufferGeometry>
          <bufferAttribute
            attach="attributes-position"
            count={particleCount}
            array={particlePositionsRef.current}
            itemSize={3}
          />
        </bufferGeometry>
        <pointsMaterial
          size={0.06}
          color={result?.result?.Allowed ? '#10b981' : '#ef4444'}
          transparent
          opacity={loading ? 0.3 : 0.8}
          sizeAttenuation
          blending={THREE.AdditiveBlending}
        />
      </points>

      {/* Nodes */}
      {Object.entries(nodePositions).map(([key, pos]) => {
        return (
          <group
            key={key}
            position={pos}
            ref={(el) => { nodesRef.current[key] = el }}
          >
            <Float speed={1.5} rotationIntensity={0.1} floatIntensity={0.3}>
              <mesh>
                <icosahedronGeometry args={[1.2, 2]} />
                <meshPhysicalMaterial
                  color={nodeColors[key]}
                  roughness={0.1}
                  metalness={0.8}
                  transmission={0.3}
                  thickness={0.5}
                  clearcoat={1}
                  transparent
                  opacity={0.9}
                />
              </mesh>
              <mesh>
                <icosahedronGeometry args={[1.22, 1]} />
                <meshBasicMaterial
                  color={nodeColors[key]}
                  wireframe
                  transparent
                  opacity={0.2}
                />
              </mesh>
            </Float>
          </group>
        )
      })}

      {/* Connection lines */}
      {flowPaths.map((path) => {
        const from = nodePositions[path.from]
        const to = nodePositions[path.to]

        const curve = new THREE.QuadraticBezierCurve3(
          new THREE.Vector3(...from),
          new THREE.Vector3(
            (from[0] + to[0]) / 2,
            1.5,
            (from[2] + to[2]) / 2
          ),
          new THREE.Vector3(...to)
        )

        const points = curve.getPoints(20)

        return (
          <Line
            key={path.from + '-' + path.to}
            points={points}
            color={nodeColors[path.to]}
            lineWidth={2}
            transparent
            opacity={0.3}
          />
        )
      })}

      {/* Stars */}
      <Stars radius={60} depth={30} count={800} factor={4} saturation={0} fade speed={0.2} />
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
    targetRef.current.x += (mouseRef.current.x * 1.5 - targetRef.current.x) * 0.02
    targetRef.current.y += (mouseRef.current.y * 1 - targetRef.current.y) * 0.02

    camera.position.x += (targetRef.current.x - camera.position.x) * 0.03
    camera.position.y += (targetRef.current.y - camera.position.y) * 0.03
    camera.lookAt(0, 0, 0)
  })

  return null
}

// AI Purchase Page Component
export default function AiPurchasePage() {
  const { showToast } = useToast()
  const [storeId, setStoreId] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [prompt, setPrompt] = useState('')
  const [suggestion, setSuggestion] = useState('')
  const [loading, setLoading] = useState(false)
  const [suggestionLoading, setSuggestionLoading] = useState(false)
  const [result, setResult] = useState(null)
  const [showVisualization, setShowVisualization] = useState(false)

  const getSuggestion = async () => {
    if (!prompt.trim()) {
      showToast('Prompt required for suggestion', 'error')
      return
    }
    setSuggestionLoading(true)
    try {
      const resp = await fetch('/api/ai/complete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt }),
      })
      if (!resp.ok) throw new Error('Groq error')
      const data = await resp.json()
      setSuggestion(data.completion)
      showToast('Got AI suggestion', 'success')
    } catch (e) {
      console.error(e)
      showToast('Failed to get suggestion', 'error')
    } finally {
      setSuggestionLoading(false)
    }
  }

  const submit = async () => {
    if (!storeId.trim() || !prompt.trim() || !apiKey.trim()) {
      showToast('Store ID, prompt, and API Key required', 'error')
      return
    }
    setLoading(true)
    setShowVisualization(true)
    try {
      const request = {
        jsonrpc: '2.0',
        id: Date.now(),
        method: 'tools/call',
        params: {
          name: 'purchase',
          arguments: {
            buyer_id: 'demo-buyer',
            session_id: 'demo-session',
            sku: 'SKU-DUMMY',
            product_id: 'PROD-DUMMY',
            quantity: 1,
            idempotency_key: crypto.randomUUID(),
          },
        },
      }

      const resp = await fetch(`/mcp/${storeId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${apiKey}`,
        },
        body: JSON.stringify(request),
      })

      const text = await resp.text()
      let data
      try {
        data = JSON.parse(text)
      } catch {
        data = { error: { message: text || `HTTP ${resp.status}` } }
      }
      setResult(data)
      if (data.error) {
        showToast(`MCP error: ${data.error.message}`, 'error')
      } else {
        showToast('MCP response received', 'success')
      }
    } catch (e) {
      console.error(e)
      showToast('Network error', 'error')
    } finally {
      setLoading(false)
    }
  }

  const reset = () => {
    setStoreId('')
    setApiKey('')
    setPrompt('')
    setSuggestion('')
    setResult(null)
    setShowVisualization(false)
  }

  return (
    <div style={{ maxWidth: '1400px', margin: '0 auto', padding: '40px 24px' }}>
      {/* Header */}
      <div style={{ textAlign: 'center', marginBottom: '60px' }}>
        <div className="section-label" style={{ color: 'var(--accent-purple)' }}>AI checkout demo</div>
        <h1 style={{ fontSize: 'clamp(32px, 5vw, 48px)', fontWeight: '900', marginBottom: '16px', letterSpacing: '-0.03em' }}>
          Test AI-Driven <span style={{ background: 'var(--gradient-brand)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', backgroundClip: 'text' }}>Purchases</span>
        </h1>
        <p style={{ fontSize: '18px', color: 'var(--text-secondary)', maxWidth: '600px', margin: '0 auto' }}>
          Enter a store ID and prompt to simulate an AI agent making a purchase through the Aegis policy gateway.
        </p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '40px', alignItems: 'start' }}>
        {/* Left Panel - Form */}
        <div className="onboarding-card" style={{ padding: '40px', background: 'rgba(12, 12, 29, 0.6)', border: '1px solid var(--border)' }}>
          <div className="form-group">
            <label className="form-label">Store ID</label>
            <input
              type="text"
              className="form-input"
              placeholder="e.g. demo-store"
              value={storeId}
              onChange={(e) => setStoreId(e.target.value)}
              disabled={loading || result}
            />
          </div>

          <div className="form-group">
            <label className="form-label">Merchant API Key</label>
            <input
              type="password"
              className="form-input"
              placeholder="store_... (shown once at provisioning)"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              disabled={loading || result}
            />
            <p className="text-xs text-tertiary mt-1">Get this from the Merchants page after provisioning</p>
          </div>

          <div className="form-group">
            <label className="form-label">Prompt</label>
            <textarea
              rows={4}
              className="form-input"
              placeholder="I want to buy a smartwatch for ₹90,000"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              disabled={loading || result}
            />
          </div>

          <div style={{ display: 'flex', gap: '12px', marginTop: '24px' }}>
            <button
              className="btn btn-secondary"
              style={{ flex: 1 }}
              onClick={getSuggestion}
              disabled={suggestionLoading || loading || result}
            >
              {suggestionLoading ? 'Thinking…' : 'Get AI Suggestion'}
            </button>
            <button
              className="btn btn-primary"
              style={{ flex: 1 }}
              onClick={submit}
              disabled={loading || result}
            >
              {loading ? 'Processing…' : 'Execute Purchase'}
            </button>
          </div>

          {suggestion && (
            <div style={{ marginTop: '24px', padding: '20px', background: 'rgba(5,5,16,0.4)', borderRadius: '12px', border: '1px solid var(--border)' }}>
              <strong style={{ color: 'var(--accent-purple)' }}>AI Suggestion:</strong>
              <p style={{ marginTop: '12px', whiteSpace: 'pre-wrap', color: 'var(--text-secondary)' }}>{suggestion}</p>
            </div>
          )}

          {result && (
            <div style={{ marginTop: '24px', padding: '24px', borderRadius: '16px', background: result.result?.Allowed
              ? 'rgba(16,185,129,0.1)'
              : 'rgba(239,68,68,0.1)',
              border: `1px solid ${result.result?.Allowed ? 'rgba(16,185,129,0.3)' : 'rgba(239,68,68,0.3)'}`,
              color: result.result?.Allowed ? 'var(--accent-green)' : 'var(--accent-red)',
            }}>
              <h3 style={{ fontSize: '18px', marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                {result.result?.Allowed ? (
                  <>
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M22 11.08V12a10 10 0 11-5.93-9.14"/><path d="M22 4L12 14.01l-3-3"/></svg>
                    Purchase Allowed
                  </>
                ) : (
                  <>
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="10"/><path d="M15 9l-6 6M9 9l6 6"/></svg>
                    Purchase Blocked
                  </>
                )}
              </h3>
              <pre style={{ fontSize: '12px', maxHeight: '200px', overflowY: 'auto', whiteSpace: 'pre-wrap', color: 'var(--text-secondary)' }}>
{JSON.stringify(result, null, 2)}
              </pre>
              <button className="btn btn-secondary" style={{ marginTop: '16px', width: '100%' }} onClick={reset}>Try Another</button>
            </div>
          )}
        </div>

        {/* Right Panel - 3D Visualization */}
        <div style={{ position: 'relative' }}>
          <div style={{ width: '100%', height: '600px', borderRadius: '20px', overflow: 'hidden', background: 'var(--surface)', border: '1px solid var(--border)' }}>
            <LazyCanvas
              threshold={0.1}
              rootMargin='100px'
              canvasProps={{
                camera: { position: [0, 0, 14], fov: 50, near: 0.1, far: 100 }
              }}
            >
              <ambientLight intensity={0.2} />
              <pointLight position={[5, 5, 5]} intensity={1} color="#a855f7" />
              <pointLight position={[-5, -3, 3]} intensity={0.8} color="#06b6d4" />
              <pointLight position={[0, -5, 5]} intensity={0.5} color="#f59e0b" />

              <PurchaseFlowVisualization
                storeId={storeId}
                result={result}
                loading={loading}
              />
              <CameraController />
            </LazyCanvas>
          </div>

          {/* Flow Legend */}
          <div style={{ display: 'flex', justifyContent: 'center', gap: '24px', marginTop: '20px', flexWrap: 'wrap' }}>
            {[
              { label: 'AI Buyer', color: '#a855f7' },
              { label: 'Merchant MCP', color: '#06b6d4' },
              { label: 'Aegis Gateway', color: '#f59e0b' },
              { label: 'Razorpay', color: '#ef4444' },
            ].map((node) => (
              <div key={node.label} style={{ display: 'flex', alignItems: 'center', gap: '8px', color: 'var(--text-secondary)', fontSize: '13px' }}>
                <div style={{ width: '12px', height: '12px', borderRadius: '50%', background: node.color, boxShadow: `0 0 8px ${node.color}` }} />
                {node.label}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}