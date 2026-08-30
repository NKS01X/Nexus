import React, { useRef, useMemo, useEffect } from 'react'
import { Canvas, useFrame, useThree } from '@react-three/fiber'
import { Float, Sparkles, MeshDistortMaterial, Environment, Stars } from '@react-three/drei'
import * as THREE from 'three'
import { LazyCanvas } from './CanvasWrapper'

// Animated torus knot with custom shader
function AnimatedTorus() {
  const meshRef = useRef()

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    if (meshRef.current) {
      meshRef.current.rotation.x = t * 0.3
      meshRef.current.rotation.y = t * 0.2
      meshRef.current.position.y = Math.sin(t * 0.5) * 0.1
    }
  })

  return (
    <mesh ref={meshRef} position={[2.5, 0, -1]}>
      <torusKnotGeometry args={[0.8, 0.25, 128, 32]} />
      <MeshDistortMaterial
        color="#a855f7"
        emissive="#7c3aed"
        emissiveIntensity={0.3}
        roughness={0.2}
        metalness={0.8}
        distort={0.3}
        speed={2}
      />
    </mesh>
  )
}

// Floating icosahedron with custom shader
function FloatingIcosahedron() {
  const meshRef = useRef()
  const uniforms = useMemo(() => ({
    uTime: { value: 0 },
    uColor1: { value: new THREE.Color('#a855f7') },
    uColor2: { value: new THREE.Color('#06b6d4') },
  }), [])

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    uniforms.uTime.value = t
    if (meshRef.current) {
      meshRef.current.rotation.y = t * 0.12
      meshRef.current.rotation.x = Math.sin(t * 0.08) * 0.15
    }
  })

  const vertexShader = `
    uniform float uTime;
    varying vec3 vPosition;
    varying vec3 vNormal;
    varying float vDisplacement;

    void main() {
      vPosition = position;
      vNormal = normal;
      float displacement = sin(position.x * 3.0 + uTime * 0.6) *
                           cos(position.y * 2.0 + uTime * 0.4) *
                           sin(position.z * 2.5 + uTime * 0.5) * 0.12;
      vDisplacement = displacement;
      vec3 newPosition = position + normal * displacement;
      gl_Position = projectionMatrix * modelViewMatrix * vec4(newPosition, 1.0);
    }
  `

  const fragmentShader = `
    uniform float uTime;
    uniform vec3 uColor1;
    uniform vec3 uColor2;
    varying vec3 vPosition;
    varying vec3 vNormal;
    varying float vDisplacement;

    void main() {
      float mixFactor = (vPosition.y + 1.5) / 3.0;
      vec3 color = mix(uColor1, uColor2, mixFactor);
      float fresnel = pow(1.0 - abs(dot(vNormal, vec3(0.0, 0.0, 1.0))), 2.5);
      color += fresnel * 0.4;
      float alpha = 0.18 + fresnel * 0.2 + abs(vDisplacement) * 2.0;
      gl_FragColor = vec4(color, alpha);
    }
  `

  return (
    <Float speed={1.5} rotationIntensity={0.2} floatIntensity={0.5}>
      <group position={[0, 0, 0]}>
        {/* Solid translucent core */}
        <mesh ref={meshRef}>
          <icosahedronGeometry args={[2.2, 3]} />
          <shaderMaterial
            uniforms={uniforms}
            vertexShader={vertexShader}
            fragmentShader={fragmentShader}
            transparent
            depthWrite={false}
            side={THREE.DoubleSide}
          />
        </mesh>

        {/* Wireframe overlay */}
        <mesh>
          <icosahedronGeometry args={[2.25, 2]} />
          <meshBasicMaterial
            color="#a855f7"
            wireframe
            transparent
            opacity={0.08}
          />
        </mesh>

        {/* Outer glow ring */}
        <mesh rotation={[Math.PI / 3, Math.PI / 6, 0]}>
          <torusGeometry args={[3.0, 0.015, 16, 100]} />
          <meshBasicMaterial
            color="#06b6d4"
            transparent
            opacity={0.25}
          />
        </mesh>

        {/* Second glow ring */}
        <mesh rotation={[Math.PI / 2, Math.PI / 4, 0]}>
          <torusGeometry args={[2.8, 0.01, 16, 80]} />
          <meshBasicMaterial
            color="#a855f7"
            transparent
            opacity={0.15}
          />
        </mesh>
      </group>
    </Float>
  )
}

// Animated particles with trails
function AnimatedParticles() {
  const particlesRef = useRef()
  const count = 150
  const positionsRef = useRef()
  const velocitiesRef = useRef()
  const sizesRef = useRef()

  // Initialize refs
  useEffect(() => {
    const pos = new Float32Array(count * 3)
    const vel = new Float32Array(count * 3)
    const siz = new Float32Array(count)
    for (let i = 0; i < count; i++) {
      pos[i * 3] = (Math.random() - 0.5) * 20
      pos[i * 3 + 1] = (Math.random() - 0.5) * 20
      pos[i * 3 + 2] = (Math.random() - 0.5) * 15
      vel[i * 3] = (Math.random() - 0.5) * 0.002
      vel[i * 3 + 1] = (Math.random() - 0.5) * 0.002
      vel[i * 3 + 2] = (Math.random() - 0.5) * 0.002
      siz[i] = Math.random() * 0.05 + 0.01
    }
    positionsRef.current = pos
    velocitiesRef.current = vel
    sizesRef.current = siz
  }, [])

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    if (particlesRef.current && positionsRef.current && velocitiesRef.current) {
      const positions = particlesRef.current.geometry.attributes.position.array
      const velocities = velocitiesRef.current
      for (let i = 0; i < count; i++) {
        positions[i * 3] += velocities[i * 3]
        positions[i * 3 + 1] += velocities[i * 3 + 1]
        positions[i * 3 + 2] += velocities[i * 3 + 2]

        // Reset particles that go too far
        if (Math.abs(positions[i * 3]) > 10) velocities[i * 3] *= -1
        if (Math.abs(positions[i * 3 + 1]) > 10) velocities[i * 3 + 1] *= -1
        if (Math.abs(positions[i * 3 + 2]) > 7.5) velocities[i * 3 + 2] *= -1
      }
      particlesRef.current.geometry.attributes.position.needsUpdate = true
      particlesRef.current.rotation.y = t * 0.01
    }
  })

  return (
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
        size={0.04}
        color="#a855f7"
        transparent
        opacity={0.6}
        sizeAttenuation
      />
    </points>
  )
}

// Glowing rings with animation
function GlowingRings() {
  const ring1Ref = useRef()
  const ring2Ref = useRef()
  const ring3Ref = useRef()

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    if (ring1Ref.current) {
      ring1Ref.current.rotation.x = t * 0.2
      ring1Ref.current.rotation.y = t * 0.15
    }
    if (ring2Ref.current) {
      ring2Ref.current.rotation.x = t * 0.15
      ring2Ref.current.rotation.z = t * 0.1
    }
    if (ring3Ref.current) {
      ring3Ref.current.rotation.y = t * 0.1
      ring3Ref.current.rotation.z = t * 0.08
    }
  })

  return (
    <group position={[-2.5, 0.5, -2]}>
      <mesh ref={ring1Ref}>
        <torusGeometry args={[1.2, 0.02, 16, 100]} />
        <meshBasicMaterial color="#a855f7" transparent opacity={0.4} />
      </mesh>
      <mesh ref={ring2Ref}>
        <torusGeometry args={[1.0, 0.015, 16, 80]} />
        <meshBasicMaterial color="#06b6d4" transparent opacity={0.3} />
      </mesh>
      <mesh ref={ring3Ref}>
        <torusGeometry args={[0.8, 0.01, 16, 60]} />
        <meshBasicMaterial color="#10b981" transparent opacity={0.25} />
      </mesh>
    </group>
  )
}

// Mouse-following camera with parallax
function CameraRig() {
  const { camera } = useThree()
  const mouseRef = useRef({ x: 0, y: 0 })

  useEffect(() => {
    const handleMouseMove = (event) => {
      mouseRef.current.x = (event.clientX / window.innerWidth) * 2 - 1
      mouseRef.current.y = -(event.clientY / window.innerHeight) * 2 + 1
    }
    window.addEventListener('mousemove', handleMouseMove)
    return () => window.removeEventListener('mousemove', handleMouseMove)
  }, [])

  useFrame(() => {
    const targetX = mouseRef.current.x * 0.5
    const targetY = mouseRef.current.y * 0.3
    camera.position.x += (targetX - camera.position.x) * 0.02
    camera.position.y += (targetY - camera.position.y) * 0.02
    camera.lookAt(0, 0, 0)
  })

  return null
}

export default function HeroScene() {
  return (
    <div className="hero-canvas-container">
      <LazyCanvas
        threshold={0.1}
        rootMargin='100px'
        canvasProps={{
          camera: { position: [0, 0, 7], fov: 55, near: 0.1, far: 50 }
        }}
      >
        <ambientLight intensity={0.15} />
        <pointLight position={[5, 5, 5]} intensity={0.5} color="#a855f7" />
        <pointLight position={[-5, -3, 3]} intensity={0.4} color="#06b6d4" />
        <pointLight position={[0, -5, 5]} intensity={0.3} color="#10b981" />

        <CameraRig />
        <FloatingIcosahedron />
        <AnimatedTorus />
        <GlowingRings />
        <AnimatedParticles />

        {/* Background stars */}
        <Stars radius={100} depth={50} count={3000} factor={4} saturation={0} fade speed={1} />

        {/* Sparkles around main element */}
        <Sparkles
          count={50}
          scale={8}
          size={2}
          speed={0.4}
          color="#a855f7"
        />
      </LazyCanvas>
    </div>
  )
}
