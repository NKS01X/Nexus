import React, { useRef, useMemo, useEffect, useState } from 'react'
import { Canvas, useFrame, useThree } from '@react-three/fiber'
import { Float, Html, EffectComposer, RenderPass, UnrealBloomPass, Stars } from '@react-three/drei'
import * as THREE from 'three'
import gsap from 'gsap'

const NODES = [
  {
    id: 'ai-agent',
    label: 'AI Agent',
    sub: 'MCP Client',
    color: '#a855f7',
    position: [-7, 0, 0],
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 2a4 4 0 014 4v2a4 4 0 01-8 0V6a4 4 0 014-4z" />
        <path d="M6 21v-2a4 4 0 014-4h4a4 4 0 014 4v2" />
        <circle cx="12" cy="6" r="1" fill="currentColor" stroke="none" />
      </svg>
    ),
  },
  {
    id: 'merchant-mcp',
    label: 'Merchant MCP',
    sub: 'Tools API',
    color: '#06b6d4',
    position: [-2, 0, 0],
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <rect x="2" y="3" width="20" height="14" rx="2" />
        <path d="M8 21h8M12 17v4" />
      </svg>
    ),
  },
  {
    id: 'policy-gateway',
    label: 'Policy Gateway',
    sub: 'Aegis Engine',
    color: '#f59e0b',
    position: [3, 0, 0],
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    ),
  },
  {
    id: 'razorpay',
    label: 'Razorpay',
    sub: 'Payment Rails',
    color: '#ef4444',
    position: [8, 0, 0],
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <rect x="1" y="4" width="22" height="16" rx="2" ry="2" />
        <path d="M1 10h22" />
      </svg>
    ),
  },
  {
    id: 'audit-log',
    label: 'Audit Log',
    sub: 'Hash-Chained',
    color: '#10b981',
    position: [3, -3.5, 0],
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71" />
        <path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71" />
      </svg>
    ),
  },
]

const CONNECTIONS = [
  { from: 'ai-agent', to: 'merchant-mcp' },
  { from: 'merchant-mcp', to: 'policy-gateway' },
  { from: 'policy-gateway', to: 'razorpay' },
  { from: 'policy-gateway', to: 'audit-log' },
  { from: 'merchant-mcp', to: 'audit-log' },
]

// Animated node with hover effects
function FlowNode({ node, index, isActive, onClick }) {
  const groupRef = useRef()
  const [hovered, setHovered] = useState(false)
  const [scale, setScale] = useState(1)

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    if (groupRef.current) {
      // Gentle floating animation
      groupRef.current.position.y = node.position[1] + Math.sin(t * 1.5 + index) * 0.1
      groupRef.current.position.x = node.position[0] + Math.cos(t * 0.8 + index * 2) * 0.05
    }
  })

  useEffect(() => {
    if (isActive) {
      gsap.to(groupRef.current?.scale, { x: 1.2, y: 1.2, z: 1.2, duration: 0.5, ease: 'elastic.out(1, 0.5)' })
    } else {
      gsap.to(groupRef.current?.scale, { x: 1, y: 1, z: 1, duration: 0.5, ease: 'power2.out' })
    }
  }, [isActive])

  return (
    <group ref={groupRef} position={node.position} onClick={onClick} onPointerOver={() => setHovered(true)} onPointerOut={() => setHovered(false)}>
      <Float speed={1.5} rotationIntensity={0.2} floatIntensity={0.3}>
        {/* Outer glow */}
        <mesh scale={hovered || isActive ? 1.3 : 1}>
          <sphereGeometry args={[1.8, 32, 32]} />
          <meshBasicMaterial
            color={node.color}
            transparent
            opacity={hovered ? 0.25 : 0.1}
            side={THREE.DoubleSide}
          />
        </mesh>

        {/* Main node body */}
        <mesh>
          <icosahedronGeometry args={[1.2, 2]} />
          <meshPhysicalMaterial
            color={node.color}
            roughness={0.1}
            metalness={0.8}
            transmission={0.3}
            thickness={0.5}
            clearcoat={1}
            clearcoatRoughness={0.1}
            transparent
            opacity={0.9}
          />
        </mesh>

        {/* Wireframe overlay */}
        <mesh>
          <icosahedronGeometry args={[1.22, 1]} />
          <meshBasicMaterial
            color={node.color}
            wireframe
            transparent
            opacity={hovered ? 0.3 : 0.15}
          />
        </mesh>

        {/* Core glow */}
        <mesh scale={0.6}>
          <sphereGeometry args={[1, 32, 32]} />
          <meshBasicMaterial
            color={node.color}
            transparent
            opacity={0.6}
          />
        </mesh>
      </Float>

      {/* HTML Label */}
      <Html
        position={[0, -2.5, 0]}
        style={{
          pointerEvents: 'none',
          textAlign: 'center',
          transform: 'translate(-50%, 0)',
        }}
      >
        <div className="arch-node-label-3d" style={{
          color: 'white',
          fontFamily: 'Inter, sans-serif',
          opacity: hovered || isActive ? 1 : 0.8,
          transition: 'opacity 0.3s ease',
        }}>
          <div style={{ fontSize: '14px', fontWeight: 700, marginBottom: '4px' }}>{node.label}</div>
          <div style={{ fontSize: '11px', color: 'rgba(255,255,255,0.7)' }}>{node.sub}</div>
        </div>
      </Html>

      {/* Particle ring around active node */}
      {isActive && (
        <ParticleRing color={node.color} radius={1.5} count={30} />
      )}
    </group>
  )
}

// Particle ring for active node
function ParticleRing({ color, radius, count }) {
  const pointsRef = useRef()
  const [positions] = useMemo(() => {
    const pos = new Float32Array(count * 3)
    for (let i = 0; i < count; i++) {
      const angle = (i / count) * Math.PI * 2
      pos[i * 3] = Math.cos(angle) * radius
      pos[i * 3 + 1] = (Math.random() - 0.5) * 0.5
      pos[i * 3 + 2] = Math.sin(angle) * radius
    }
    return pos
  }, [count, radius])

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    if (pointsRef.current) {
      pointsRef.current.rotation.y = t * 0.5
      const positions = pointsRef.current.geometry.attributes.position.array
      for (let i = 0; i < count; i++) {
        const angle = (i / count) * Math.PI * 2 + t * 0.5
        positions[i * 3] = Math.cos(angle) * radius
        positions[i * 3 + 2] = Math.sin(angle) * radius
        positions[i * 3 + 1] = Math.sin(t * 3 + i) * 0.2
      }
      pointsRef.current.geometry.attributes.position.needsUpdate = true
    }
  })

  return (
    <points ref={pointsRef}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={count}
          array={positions}
          itemSize={3}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.08}
        color={color}
        transparent
        opacity={0.8}
        sizeAttenuation
      />
    </points>
  )
}

// Animated flow particles between nodes
function FlowParticles() {
  const pointsRef = useRef()
  const totalParticles = 80
  const [positions] = useMemo(() => {
    const pos = new Float32Array(totalParticles * 3)
    return pos
  }, [])

  const paths = useMemo(() => {
    const nodeMap = {}
    NODES.forEach(n => { nodeMap[n.id] = n.position })
    return CONNECTIONS.map(c => ({
      from: nodeMap[c.from],
      to: nodeMap[c.to],
    }))
  }, [])

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    if (pointsRef.current) {
      const positions = pointsRef.current.geometry.attributes.position.array
      const pathCount = paths.length

      for (let i = 0; i < totalParticles; i++) {
        const pathIndex = i % pathCount
        const path = paths[pathIndex]
        const progress = ((t * 0.5 + i / totalParticles) % 1)

        // Easing function for smooth movement
        const eased = progress < 0.5
          ? 2 * progress * progress
          : 1 - Math.pow(-2 * progress + 2, 2) / 2

        const x = path.from[0] + (path.to[0] - path.from[0]) * eased
        const y = path.from[1] + (path.to[1] - path.from[1]) * eased + Math.sin(t * 3 + i) * 0.1
        const z = path.from[2] + (path.to[2] - path.from[2]) * eased

        positions[i * 3] = x
        positions[i * 3 + 1] = y
        positions[i * 3 + 2] = z
      }
      pointsRef.current.geometry.attributes.position.needsUpdate = true
    }
  })

  return (
    <points ref={pointsRef}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={totalParticles}
          array={positions}
          itemSize={3}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.06}
        color="#ffffff"
        transparent
        opacity={0.8}
        sizeAttenuation
        blending={THREE.AdditiveBlending}
      />
    </points>
  )
}

// Connection lines with glow
function ConnectionLines() {
  const lineRefs = useRef([])

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    lineRefs.current.forEach((line, i) => {
      if (line && line.material) {
        line.material.opacity = 0.3 + Math.sin(t * 2 + i) * 0.2
      }
    })
  })

  const nodeMap = {}
  NODES.forEach(n => { nodeMap[n.id] = n.position })

  return (
    <group>
      {CONNECTIONS.map((conn, i) => {
        const from = nodeMap[conn.from]
        const to = nodeMap[conn.to]

        // Add curve for visual interest
        const curve = new THREE.QuadraticBezierCurve3(
          new THREE.Vector3(...from),
          new THREE.Vector3(
            (from[0] + to[0]) / 2,
            Math.max(from[1], to[1]) + 1.5,
            (from[2] + to[2]) / 2
          ),
          new THREE.Vector3(...to)
        )

        const curvePoints = curve.getPoints(20)
        const geometry = new THREE.BufferGeometry().setFromPoints(curvePoints)

        return (
          <line
            key={conn.from + '-' + conn.to}
            ref={(el) => { lineRefs.current[i] = el }}
            geometry={geometry}
          >
            <lineMaterial
              color="#ffffff"
              linewidth={2}
              transparent
              opacity={0.4}
              dashed={false}
            />
          </line>
        )
      })}
    </group>
  )
}

// Background atmosphere
function Atmosphere() {
  return (
    <group>
      {/* Large ambient glow spheres */}
      <mesh position={[-10, 5, -5]}>
        <sphereGeometry args={[8, 32, 32]} />
        <meshBasicMaterial color="#a855f7" transparent opacity={0.02} side={THREE.BackSide} />
      </mesh>
      <mesh position={[10, -5, -5]}>
        <sphereGeometry args={[6, 32, 32]} />
        <meshBasicMaterial color="#06b6d4" transparent opacity={0.02} side={THREE.BackSide} />
      </mesh>

      {/* Floating geometric shapes */}
      <Float speed={2} rotationIntensity={0.3} floatIntensity={0.8}>
        <mesh position={[-8, 3, -3]}>
          <octahedronGeometry args={[0.8, 0]} />
          <meshPhysicalMaterial color="#a855f7" roughness={0} metalness={1} transparent opacity={0.3} />
        </mesh>
      </Float>
      <Float speed={1.5} rotationIntensity={0.2} floatIntensity={0.5}>
        <mesh position={[8, -3, -3]}>
          <tetrahedronGeometry args={[0.6, 0]} />
          <meshPhysicalMaterial color="#06b6d4" roughness={0} metalness={1} transparent opacity={0.3} />
        </mesh>
      </Float>
      <Float speed={2.5} rotationIntensity={0.4} floatIntensity={0.6}>
        <mesh position={[0, 4, -4]}>
          <icosahedronGeometry args={[0.5, 0]} />
          <meshPhysicalMaterial color="#f59e0b" roughness={0} metalness={1} transparent opacity={0.4} />
        </mesh>
      </Float>

      {/* Stars */}
      <Stars radius={100} depth={50} count={2000} factor={4} saturation={0} fade speed={0.5} />
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
    targetRef.current.x += (mouseRef.current.x * 1.5 - targetRef.current.x) * 0.03
    targetRef.current.y += (mouseRef.current.y * 1 - targetRef.current.y) * 0.03

    camera.position.x += (targetRef.current.x - camera.position.x) * 0.05
    camera.position.y += (targetRef.current.y - camera.position.y) * 0.05
    camera.lookAt(0, 0, 0)
  })

  return null
}

export default function ArchitectureDiagram3D() {
  const [activeNode, setActiveNode] = useState(null)

  return (
    <div className="arch-3d-container" style={{ width: '100%', height: '500px' }}>
      <Canvas
        camera={{ position: [0, 0, 15], fov: 50, near: 0.1, far: 100 }}
        dpr={[1, Math.min(window.devicePixelRatio, 2)]}
        gl={{
          antialias: true,
          alpha: true,
          powerPreference: 'high-performance',
        }}
        style={{ background: 'transparent', width: '100%', height: '100%' }}
      >
        <ambientLight intensity={0.3} />
        <pointLight position={[-5, 5, 5]} intensity={1} color="#a855f7" />
        <pointLight position={[5, 5, 5]} intensity={1} color="#06b6d4" />
        <pointLight position={[0, -5, 5]} intensity={0.5} color="#f59e0b" />

        <Atmosphere />
        <CameraController />
        <ConnectionLines />
        <FlowParticles />

        {NODES.map((node, i) => (
          <FlowNode
            key={node.id}
            node={node}
            index={i}
            isActive={activeNode === node.id}
            onClick={() => setActiveNode(activeNode === node.id ? null : node.id)}
          />
        ))}

        {/* Post-processing bloom */}
        <EffectComposer>
          <RenderPass />
          <UnrealBloomPass
            strength={0.8}
            radius={0.4}
            threshold={0.3}
          />
        </EffectComposer>
      </Canvas>
    </div>
  )
}