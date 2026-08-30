import React, { useRef, useEffect, useState } from 'react'
import { Canvas, useFrame, useThree } from '@react-three/fiber'
import { Html, Float, MeshWobbleMaterial } from '@react-three/drei'
import * as THREE from 'three'
import gsap from 'gsap'

const FEATURES = [
  {
    id: 'mcp-storefront',
    icon: (
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <rect x="2" y="3" width="20" height="14" rx="2" />
        <path d="M8 21h8M12 17v4" />
      </svg>
    ),
    title: 'MCP-Native Storefront',
    desc: 'search_products, purchase, get_order_status — exposed as standard MCP tools any AI agent can call.',
    color: '#a855f7',
    position: [-4.5, 0, 0],
  },
  {
    id: 'policy-engine',
    icon: (
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    ),
    title: 'Deterministic Policy Engine',
    desc: 'Spend caps, velocity limits, category allowlists, SKU blocklists — all compiled Go, no LLM in the enforcement path.',
    color: '#06b6d4',
    position: [0, 0, 0],
  },
  {
    id: 'audit-trail',
    icon: (
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71" />
        <path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71" />
      </svg>
    ),
    title: 'Hash-Chained Audit Trail',
    desc: 'Every decision — allowed or blocked — appended to a SHA-256 linked log. Tamper-evident and verifiable on demand.',
    color: '#f59e0b',
    position: [4.5, 0, 0],
  },
]

// 3D Feature Card
function FeatureCard3D({ feature, index, isVisible }) {
  const groupRef = useRef()
  const [hovered, setHovered] = useState(false)
  const [clicked, setClicked] = useState(false)

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    if (groupRef.current) {
      groupRef.current.position.y = feature.position[1] + Math.sin(t * 1.2 + index * 2) * 0.08
    }
  })

  useEffect(() => {
    if (isVisible) {
      gsap.from(groupRef.current?.position, {
        y: feature.position[1] + 2,
        opacity: 0,
        duration: 0.8,
        delay: index * 0.15,
        ease: 'power3.out',
      })
      gsap.from(groupRef.current?.rotation, {
        x: -0.5,
        duration: 0.8,
        delay: index * 0.15,
        ease: 'power3.out',
      })
    }
  }, [isVisible, index])

  return (
    <group
      ref={groupRef}
      position={feature.position}
      onPointerOver={() => setHovered(true)}
      onPointerOut={() => setHovered(false)}
      onClick={() => setClicked(!clicked)}
    >
      <Float speed={1.5} rotationIntensity={0.1} floatIntensity={0.4}>
        {/* Card background with glass effect */}
        <mesh
          scale={[4.2, 5.5, 0.3]}
          position={[0, 0, 0]}
          onPointerOver={() => setHovered(true)}
          onPointerOut={() => setHovered(false)}
        >
          <boxGeometry args={[1, 1, 1]} />
          <meshPhysicalMaterial
            color={feature.color}
            roughness={0.1}
            metalness={0.2}
            transmission={0.5}
            thickness={0.5}
            clearcoat={1}
            clearcoatRoughness={0.1}
            transparent
            opacity={hovered ? 0.25 : 0.15}
            side={THREE.DoubleSide}
          />
        </mesh>

        {/* Card border glow */}
        <mesh
          scale={[4.3, 5.6, 0.35]}
          position={[0, 0, -0.01]}
        >
          <boxGeometry args={[1, 1, 1]} />
          <meshBasicMaterial
            color={feature.color}
            wireframe
            transparent
            opacity={hovered ? 0.4 : 0.2}
          />
        </mesh>

        {/* Floating geometric accent */}
        <mesh position={[0, 2.2, 0.3]} scale={hovered ? 1.2 : 1}>
          <icosahedronGeometry args={[0.5, 1]} />
          <MeshWobbleMaterial
            color={feature.color}
            factor={hovered ? 0.3 : 0.15}
            speed={2}
            roughness={0}
            metalness={1}
          />
        </mesh>

        {/* Accent ring */}
        <mesh position={[0, 2.2, 0.2]} rotation={[0, 0, Math.PI / 4]} scale={hovered ? 1.1 : 1}>
          <torusGeometry args={[0.7, 0.03, 8, 32]} />
          <meshBasicMaterial color={feature.color} transparent opacity={hovered ? 0.6 : 0.4} />
        </mesh>
      </Float>

      {/* HTML Content Overlay */}
      <Html
        position={[-2, -1.5, 0.3]}
        style={{
          pointerEvents: 'none',
          width: '360px',
          transform: 'translate(0, 0)',
        }}
      >
        <div
          className="feature-card-3d"
          style={{
            background: 'rgba(12, 12, 29, 0.8)',
            backdropFilter: 'blur(20px) saturate(1.3)',
            border: '1px solid rgba(255,255,255,0.06)',
            borderRadius: '20px',
            padding: '32px',
            color: 'white',
            fontFamily: 'Inter, sans-serif',
            boxShadow: '0 4px 20px -2px rgba(0,0,0,0.5), 0 2px 4px -1px rgba(255,255,255,0.02)',
            transition: 'all 0.4s cubic-bezier(0.16, 1, 0.3, 1)',
            opacity: hovered ? 1 : 0.95,
            transform: hovered ? 'translateY(-4px) scale(1.01)' : 'none',
            borderColor: hovered ? feature.color : 'rgba(255,255,255,0.06)',
            boxShadow: hovered ? '0 25px 50px -12px rgba(0,0,0,0.7), inset 0 1px 1px rgba(255,255,255,0.06), 0 0 30px -5px ' + feature.color + '40' : '0 4px 20px -2px rgba(0,0,0,0.5), 0 2px 4px -1px rgba(255,255,255,0.02)',
          }}
        >
          <div
            style={{
              width: '52px',
              height: '52px',
              borderRadius: '14px',
              background: 'rgba(168,85,247,0.1)',
              border: '1px solid rgba(168,85,247,0.2)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              marginBottom: '20px',
              color: feature.color,
              transition: 'all 0.4s cubic-bezier(0.16, 1, 0.3, 1)',
              background: hovered ? 'linear-gradient(135deg, ' + feature.color + ', ' + feature.color + 'CC)' : 'rgba(168,85,247,0.1)',
              borderColor: hovered ? 'transparent' : 'rgba(168,85,247,0.2)',
              color: hovered ? 'white' : feature.color,
              boxShadow: hovered ? '0 4px 16px ' + feature.color + '4D' : 'none',
            }}
          >
            {feature.icon}
          </div>
          <div style={{ fontSize: '18px', fontWeight: 700, marginBottom: '10px', letterSpacing: '-0.01em' }}>
            {feature.title}
          </div>
          <div style={{ fontSize: '14px', color: '#7a7a95', lineHeight: '1.7' }}>
            {feature.desc}
          </div>
        </div>
      </Html>

      {/* Particle trail on hover */}
      {hovered && <HoverParticles color={feature.color} position={feature.position} />}
    </group>
  )
}

// Particle effect on hover
function HoverParticles({ color, position }) {
  const pointsRef = useRef()
  const count = 30
  const [positions] = useMemo(() => {
    const pos = new Float32Array(count * 3)
    for (let i = 0; i < count; i++) {
      pos[i * 3] = (Math.random() - 0.5) * 4
      pos[i * 3 + 1] = (Math.random() - 0.5) * 5
      pos[i * 3 + 2] = (Math.random() - 0.5) * 1 + 0.5
    }
    return pos
  }, [count])

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    if (pointsRef.current) {
      pointsRef.current.rotation.y = t * 0.3
      const positions = pointsRef.current.geometry.attributes.position.array
      for (let i = 0; i < count; i++) {
        positions[i * 3 + 1] += 0.01
        if (positions[i * 3 + 1] > 2.5) {
          positions[i * 3 + 1] = -2.5
          positions[i * 3] = (Math.random() - 0.5) * 4
          positions[i * 3 + 2] = (Math.random() - 0.5) * 1 + 0.5
        }
      }
      pointsRef.current.geometry.attributes.position.needsUpdate = true
    }
  })

  return (
    <points ref={pointsRef} position={position}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={count}
          array={positions}
          itemSize={3}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.04}
        color={color}
        transparent
        opacity={0.8}
        sizeAttenuation
        blending={THREE.AdditiveBlending}
      />
    </points>
  )
}

// Background atmosphere
function FeatureAtmosphere() {
  return (
    <group>
      <Float speed={3} rotationIntensity={0.2} floatIntensity={1}>
        <mesh position={[-6, 3, -3]}>
          <octahedronGeometry args={[0.6, 0]} />
          <meshPhysicalMaterial color="#a855f7" roughness={0} metalness={1} transparent opacity={0.15} />
        </mesh>
      </Float>
      <Float speed={2.5} rotationIntensity={0.3} floatIntensity={0.8}>
        <mesh position={[6, -2, -3]}>
          <tetrahedronGeometry args={[0.5, 0]} />
          <meshPhysicalMaterial color="#06b6d4" roughness={0} metalness={1} transparent opacity={0.15} />
        </mesh>
      </Float>
      <Float speed={4} rotationIntensity={0.1} floatIntensity={0.6}>
        <mesh position={[0, 4, -4]}>
          <icosahedronGeometry args={[0.4, 0]} />
          <meshPhysicalMaterial color="#f59e0b" roughness={0} metalness={1} transparent opacity={0.2} />
        </mesh>
      </Float>
    </group>
  )
}

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
    targetRef.current.x += (mouseRef.current.x * 1 - targetRef.current.x) * 0.02
    targetRef.current.y += (mouseRef.current.y * 0.5 - targetRef.current.y) * 0.02

    camera.position.x += (targetRef.current.x - camera.position.x) * 0.03
    camera.position.y += (targetRef.current.y - camera.position.y) * 0.03
    camera.lookAt(0, 0, 0)
  })

  return null
}

export default function FeatureCards() {
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setVisible(true)
          }
        })
      },
      { threshold: 0.15, rootMargin: '0px 0px -100px 0px' }
    )
    const section = document.getElementById('features')
    if (section) observer.observe(section)
    return () => observer.disconnect()
  }, [])

  return (
    <section className="features-section" id="features">
      <div style={{ textAlign: 'center', marginBottom: '60px' }}>
        <div className="section-label">Capabilities</div>
        <h2 className="section-title">Built for Agent Commerce</h2>
        <p className="section-subtitle">
          Three layers of infrastructure that make AI purchasing safe, auditable, and merchant-controlled.
        </p>
      </div>

      <div style={{ width: '100%', height: '550px', position: 'relative' }}>
        <Canvas
          camera={{ position: [0, 0, 12], fov: 50, near: 0.1, far: 100 }}
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

          <FeatureAtmosphere />
          <CameraController />

          {FEATURES.map((feature, i) => (
            <FeatureCard3D
              key={feature.id}
              feature={feature}
              index={i}
              isVisible={visible}
            />
          ))}
        </Canvas>
      </div>
    </section>
  )
}