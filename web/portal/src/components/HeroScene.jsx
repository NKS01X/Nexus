import React, { useRef, useMemo, useEffect } from 'react'
import { Canvas, useFrame } from '@react-three/fiber'
import { Float } from '@react-three/drei'
import * as THREE from 'three'

function NexusMesh() {
  const meshRef = useRef()
  const wireRef = useRef()
  const glowRef = useRef()

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
    if (wireRef.current) {
      wireRef.current.rotation.y = t * 0.12
      wireRef.current.rotation.x = Math.sin(t * 0.08) * 0.15
    }
    if (glowRef.current) {
      glowRef.current.rotation.y = t * 0.08
      glowRef.current.rotation.z = t * 0.05
      const scale = 1 + Math.sin(t * 0.5) * 0.04
      glowRef.current.scale.setScalar(scale)
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
                           sin(position.z * 2.5 + uTime * 0.5) * 0.08;
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
      color += fresnel * 0.3;
      float alpha = 0.12 + fresnel * 0.15 + abs(vDisplacement) * 2.0;
      gl_FragColor = vec4(color, alpha);
    }
  `

  return (
    <Float speed={1.5} rotationIntensity={0.2} floatIntensity={0.5}>
      <group>
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
        <mesh ref={wireRef}>
          <icosahedronGeometry args={[2.25, 2]} />
          <meshBasicMaterial
            color="#a855f7"
            wireframe
            transparent
            opacity={0.08}
          />
        </mesh>

        {/* Outer glow ring */}
        <mesh ref={glowRef}>
          <torusGeometry args={[3.0, 0.015, 16, 100]} />
          <meshBasicMaterial
            color="#06b6d4"
            transparent
            opacity={0.2}
          />
        </mesh>

        {/* Second glow ring */}
        <mesh rotation={[Math.PI / 3, Math.PI / 6, 0]}>
          <torusGeometry args={[2.8, 0.01, 16, 80]} />
          <meshBasicMaterial
            color="#a855f7"
            transparent
            opacity={0.12}
          />
        </mesh>
      </group>
    </Float>
  )
}

function Particles() {
  const particlesRef = useRef()
  const count = 80

  const positions = useMemo(() => {
    const pos = new Float32Array(count * 3)
    for (let i = 0; i < count; i++) {
      pos[i * 3] = (Math.random() - 0.5) * 16
      pos[i * 3 + 1] = (Math.random() - 0.5) * 16
      pos[i * 3 + 2] = (Math.random() - 0.5) * 10
    }
    return pos
  }, [])

  useFrame((state) => {
    const t = state.clock.getElapsedTime()
    if (particlesRef.current) {
      particlesRef.current.rotation.y = t * 0.02
    }
  })

  return (
    <points ref={particlesRef}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={count}
          array={positions}
          itemSize={3}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.03}
        color="#a855f7"
        transparent
        opacity={0.4}
        sizeAttenuation
      />
    </points>
  )
}

export default function HeroScene() {
  return (
    <div className="hero-canvas-container">
      <Canvas
        camera={{ position: [0, 0, 7], fov: 55, near: 0.1, far: 50 }}
        dpr={[1, Math.min(window.devicePixelRatio, 2)]}
        gl={{
          antialias: true,
          alpha: true,
          powerPreference: 'high-performance',
        }}
        style={{ background: 'transparent' }}
      >
        <ambientLight intensity={0.15} />
        <pointLight position={[5, 5, 5]} intensity={0.4} color="#a855f7" />
        <pointLight position={[-5, -3, 3]} intensity={0.3} color="#06b6d4" />
        <NexusMesh />
        <Particles />
      </Canvas>
    </div>
  )
}
