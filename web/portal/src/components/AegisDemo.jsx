import React, { useRef, useMemo, useEffect, useState } from 'react'
import { useFrame, useThree } from '@react-three/fiber'
import { Float, Text, Stars, Html } from '@react-three/drei'

import * as THREE from 'three'
import { LazyCanvas } from './CanvasWrapper'

// Laptop model with screen showing the intent text
function Laptop({ screenOpacity }) {
  const screenRef = useRef()
  const laptopRef = useRef()
  return (
    <group ref={laptopRef} position={[0, -0.5, 0]}>
      {/* Base of the laptop */}
      <mesh>
        <boxGeometry args={[4, 0.15, 2.5]} />
        <meshPhysicalMaterial color="#111" metalness={0.8} roughness={0.2} />
      </mesh>
      {/* Screen – a thin plane with emissive text */}
      <mesh ref={screenRef} position={[0, 0.8, -1.2]} rotation={[Math.PI / 2, 0, 0]}> {/* Rotate to face camera */}
        <planeGeometry args={[3.6, 2.2]} />
        <meshPhysicalMaterial
          color="#111"
          transmission={0.9}
          thickness={0.2}
          roughness={0.1}
          metalness={0.5}
          emissive="#0a0a0a"
          emissiveIntensity={0.6}
          opacity={screenOpacity}
          transparent
        />
        {/* Intent text */}
        <Text
          fontSize={0.25}
          position={[0, 0, 0.01]}
          color="#fff"
          anchorX="center"
          anchorY="middle"
        >
          Buy 1 red shoe for me.
        </Text>
      </mesh>
    </group>
  )
}

// The single blue filament representing the purchase intent
function Filament({ progress }) {
  // Using a tapered tube from start to end based on progress
  const start = new THREE.Vector3(0, 0.6, -0.5) // just in front of screen
  const end = new THREE.Vector3(0, 0, -6) // gate position
  const curPos = start.clone().lerp(end, progress)
  const points = []
  const segments = 20
  for (let i = 0; i <= segments; i++) {
    const p = start.clone().lerp(end, i / segments)
    points.push(p)
  }
  const curve = new THREE.CatmullRomCurve3(points)
  const tubeGeometry = new THREE.TubeGeometry(curve, 30, 0.04, 8, false)
  return (
    <mesh geometry={tubeGeometry} position={[0, 0, 0]}>
      <meshBasicMaterial color="#3b82f6" transparent opacity={0.9} />
    </mesh>
  )
}

// Red torrent – many particles emitted from the filament tip
function RedTorrent({ start, active }) {
  const pointsRef = useRef()
  const particleCount = 800
  const positions = useMemo(() => new Float32Array(particleCount * 3), [])
  const velocities = useMemo(() => new Float32Array(particleCount * 3), [])

  // Initialise particle positions & velocities when active becomes true
  useEffect(() => {
    if (!active) return
    for (let i = 0; i < particleCount; i++) {
      const idx = i * 3
      positions[idx] = start.x + (Math.random() - 0.5) * 0.2
      positions[idx + 1] = start.y + (Math.random() - 0.5) * 0.2
      positions[idx + 2] = start.z
      // Randomised forward velocity (towards negative Z)
      const speed = 0.4 + Math.random() * 0.2
      velocities[idx] = (Math.random() - 0.5) * 0.02
      velocities[idx + 1] = (Math.random() - 0.5) * 0.02
      velocities[idx + 2] = -speed
    }
  }, [active, start, positions, velocities])

  useFrame(() => {
    if (!active || !pointsRef.current) return
    const posArr = pointsRef.current.geometry.attributes.position.array
    for (let i = 0; i < particleCount; i++) {
      const idx = i * 3
      posArr[idx] += velocities[idx]
      posArr[idx + 1] += velocities[idx + 1]
      posArr[idx + 2] += velocities[idx + 2]
    }
    pointsRef.current.geometry.attributes.position.needsUpdate = true
  })

  return (
    <points ref={pointsRef}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={particleCount}
          array={positions}
          itemSize={3}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.08}
        color="#ef4444"
        transparent
        opacity={0.8}
        sizeAttenuation
        blending={THREE.AdditiveBlending}
      />
    </points>
  )
}

// Aegis Policy Gateway – a glass‑like torus that shatters the torrent
function Gateway({ visible }) {
  const gateRef = useRef()
  return (
    <group visible={visible} ref={gateRef} position={[0, 0, -6]}>
      {/* Main torus body */}
      <mesh>
        <torusGeometry args={[2.2, 0.25, 64, 100]} />
        <meshPhysicalMaterial
          color="#9ca3af"
          transmission={0.9}
          thickness={0.5}
          roughness={0.05}
          metalness={0.9}
          clearcoat={1}
          clearcoatRoughness={0.02}
          reflectivity={0.5}
        />
      </mesh>
      {/* Overlay holographic text */}
      <Float speed={1.2} rotationIntensity={0.2} floatIntensity={0.3}>
        <Text
          fontSize={0.25}
          position={[0, 0.8, 0]}
          color="#ef4444"
          anchorX="center"
          anchorY="middle"
        >
          QUANTITY CAP EXCEEDED
          {'\n'}REQUEST BLOCKED
          {'\n'}FORWARDED TO APPROVAL QUEUE
        </Text>
      </Float>
    </group>
  )
}

// Dust particles emitted after the collision
function Dust({ active }) {
  const pointsRef = useRef()
  const particleCount = 500
  const positions = useMemo(() => new Float32Array(particleCount * 3), [])
  const lifetimes = useMemo(() => new Float32Array(particleCount), [])

  // Spawn dust when active toggles true
  useEffect(() => {
    if (!active) return
    for (let i = 0; i < particleCount; i++) {
      const idx = i * 3
      // Start near the gate centre
      positions[idx] = (Math.random() - 0.5) * 0.2
      positions[idx + 1] = (Math.random() - 0.5) * 0.2
      positions[idx + 2] = -6
      lifetimes[i] = Math.random() * 2 + 1 // seconds
    }
  }, [active, positions, lifetimes])

  useFrame((state, delta) => {
    if (!active || !pointsRef.current) return
    const posArr = pointsRef.current.geometry.attributes.position.array
    for (let i = 0; i < particleCount; i++) {
      const idx = i * 3
      // Simple upward drift and fade
      posArr[idx] += (Math.random() - 0.5) * 0.01
      posArr[idx + 1] += 0.02
      // Decrease lifetime (used for opacity via material)
      lifetimes[i] -= delta
    }
    pointsRef.current.geometry.attributes.position.needsUpdate = true
    // Update opacity via material uniform – here we keep it static for simplicity
  })

  return (
    <points ref={pointsRef}>
      <bufferGeometry>
        <bufferAttribute attach="attributes-position" count={particleCount} array={positions} itemSize={3} />
      </bufferGeometry>
      <pointsMaterial size={0.04} color="#ef4444" transparent opacity={0.5} sizeAttenuation />
    </points>
  )
}

/**
 * Main demo scene – orchestrates the timeline described in the storyboard.
 */
function AegisDemoScene() {
  const { camera } = useThree()
  const [showResultOverlay, setShowResultOverlay] = useState(false)
  const startTimeRef = useRef(0)
  const [screenOpacity, setScreenOpacity] = useState(1)
  const [filamentProgress, setFilamentProgress] = useState(0)
  const [showFilament, setShowFilament] = useState(false)
  const [showTorrent, setShowTorrent] = useState(false)
  const [gateVisible, setGateVisible] = useState(false)
  const [showDust, setShowDust] = useState(false)

  // Camera initial position
  useEffect(() => {
    camera.position.set(0, 0, 12)
    camera.lookAt(0, 0, 0)
  }, [camera])

  useFrame((state) => {
    const elapsed = state.clock.getElapsedTime() - startTimeRef.current
    // Phase 0‑3: idle, laptop visible
    if (elapsed < 3) {
      // nothing
    }
    // Phase 3‑5: camera pushes through screen, fade screen
    if (elapsed >= 3 && elapsed < 5) {
      const t = (elapsed - 3) / 2 // 0‑1
      // Interpolate camera position from z=12 to z=3
      const startPos = new THREE.Vector3(0, 0, 12)
      const endPos = new THREE.Vector3(0, 0, 3)
      camera.position.lerpVectors(startPos, endPos, t)
      setScreenOpacity(1 - t)
    }
    // Phase 5‑8: filament appears and red torrent emitted
    if (elapsed >= 5 && elapsed < 8) {
      if (!showFilament) setShowFilament(true)
      // progress of filament from screen to gate (0‑1)
      const t = (elapsed - 5) / 3
      setFilamentProgress(Math.min(t, 1))
      if (!showTorrent) setShowTorrent(true)
    }
    // Phase 8‑12: gateway appears, torrent collides and turns into dust
    if (elapsed >= 8 && elapsed < 12) {
      if (!gateVisible) setGateVisible(true)
      // simple collision: after 9 seconds we stop torrent and spawn dust
      if (elapsed >= 9 && !showDust) setShowDust(true)
    }
    // Phase 12+: pull camera back and reveal UI overlay
    if (elapsed >= 12) {
      if (!showResultOverlay) setShowResultOverlay(true)
      const t = Math.min(1, (elapsed - 12) / 3)
      const backStart = new THREE.Vector3(0, 0, 3)
      const backEnd = new THREE.Vector3(0, 0, 12)
      camera.position.lerpVectors(backStart, backEnd, t)
    }
  })

  // Placement for the filament – derived from current progress
  const filamentTip = useMemo(() => {
    const start = new THREE.Vector3(0, 0.6, -0.5)
    const end = new THREE.Vector3(0, 0, -6)
    return start.clone().lerp(end, filamentProgress)
  }, [filamentProgress])

  return (
    <>
      {/* Ambient lighting */}
      <ambientLight intensity={0.3} />
      <pointLight position={[5, 5, 5]} intensity={1} color="#a855f7" />
      <pointLight position={[-5, -3, 3]} intensity={0.8} color="#06b6d4" />
      <pointLight position={[0, -5, 5]} intensity={0.5} color="#f59e0b" />

      {/* Stars background for depth */}
      <Stars radius={100} depth={50} count={2000} factor={4} saturation={0} fade speed={0.6} />

      {/* Scene objects */}
      <Laptop screenOpacity={screenOpacity} />
      {showFilament && <Filament progress={filamentProgress} />}
      {showTorrent && <RedTorrent start={filamentTip} active={showTorrent} />}
      <Gateway visible={gateVisible} />
      {showDust && <Dust active={showDust} />}

      {/* Bloom post‑processing */}
      {/* Bloom effect omitted */}

      {/* UI overlay after the animation completes */}
      {showResultOverlay && (
        <Html position={[0, 0, 0]} center>
          <div style={{
            background: 'rgba(10,10,15,0.8)',
            padding: '20px 32px',
            borderRadius: '12px',
            color: '#fff',
            textAlign: 'center',
            fontFamily: 'system-ui, sans-serif',
            lineHeight: 1.4
          }}>
            <strong>Malicious AI Purchase Blocked (10,000× Quantity)</strong>
          </div>
        </Html>
      )}
    </>
  )
}

/**
 * Page component – wraps the demo scene in a lazy‑loaded Canvas and adds a brief description.
 */
export default function AegisDemo() {
  return (
    <div style={{ width: '100%', height: '100vh', background: 'var(--surface)', overflow: 'hidden' }}>
      <LazyCanvas
        threshold={0.1}
        rootMargin='100px'
        canvasProps={{
          camera: { position: [0, 0, 12], fov: 55, near: 0.1, far: 100 },
          gl: { antialias: true, alpha: true },
        }}
      >
        <AegisDemoScene />
      </LazyCanvas>
    </div>
  )
}
