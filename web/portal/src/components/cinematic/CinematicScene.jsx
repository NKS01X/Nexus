import React, { useEffect, useMemo, useRef } from 'react'
import { useFrame, useThree } from '@react-three/fiber'
import {
  ContactShadows,
  Environment,
  Html,
  MeshTransmissionMaterial,
  RoundedBox,
} from '@react-three/drei'
import { EffectComposer, Bloom, DepthOfField, Vignette, ChromaticAberration } from '@react-three/postprocessing'
import * as THREE from 'three'
import { DURATION, GATE, sampleCamera, screenMode } from './timeline'
import { createScreenTexture, drawLaptopScreen } from './screenTexture'

const INTENT = 'Buy 1 red shoe for me.'
const PARTICLE_COUNT = 2200
const _pos = new THREE.Vector3()
const _look = new THREE.Vector3()
const _dummy = new THREE.Object3D()
const _color = new THREE.Color()
const RED = new THREE.Color('#ef4444')
const DUST = new THREE.Color('#fb7185')

function typedIntent(t) {
  const start = 0.45
  const end = 2.35
  if (t <= start) return ''
  if (t >= end) return INTENT
  const u = (t - start) / (end - start)
  return INTENT.slice(0, Math.floor(u * INTENT.length))
}

function CameraDirector({ clockRef, dofRef }) {
  const { camera } = useThree()
  useFrame(() => {
    const t = clockRef.current
    const sampled = sampleCamera(t, _pos, _look)
    camera.position.copy(_pos)
    camera.lookAt(_look)
    camera.fov = sampled.fov
    camera.updateProjectionMatrix()
    const pass = dofRef.current
    if (pass) {
      if (typeof pass.bokehScale !== 'undefined') pass.bokehScale = t > 3 && t < 13.2 ? 3.6 : 2.1
      if (typeof pass.focalLength !== 'undefined') pass.focalLength = sampled.dof
    }
  })
  return null
}

function WorldAtmosphere({ clockRef }) {
  const { scene } = useThree()
  const fog = useMemo(() => new THREE.FogExp2('#1c1612', 0.045), [])
  const background = useMemo(() => new THREE.Color('#1c1612'), [])
  useEffect(() => {
    scene.fog = fog
    scene.background = background
    return () => {
      scene.fog = null
    }
  }, [scene, fog, background])

  useFrame(() => {
    if (!scene.background || !('setRGB' in scene.background)) {
      scene.background = background
    }
    const t = clockRef.current
    const intoVoid = THREE.MathUtils.smoothstep(t, 3.05, 4.6)
    const back = THREE.MathUtils.smoothstep(t, 12.6, 14.4)
    const k = intoVoid * (1 - back)
    fog.color.setRGB(
      THREE.MathUtils.lerp(0.11, 0.012, k),
      THREE.MathUtils.lerp(0.086, 0.016, k),
      THREE.MathUtils.lerp(0.07, 0.03, k),
    )
    fog.density = THREE.MathUtils.lerp(0.038, 0.018, k)
    scene.background.setRGB(
      THREE.MathUtils.lerp(0.11, 0.01, k),
      THREE.MathUtils.lerp(0.082, 0.012, k),
      THREE.MathUtils.lerp(0.068, 0.02, k),
    )
  })
  return null
}

function SeatedFigure() {
  const armRef = useRef()
  useFrame(({ clock }) => {
    if (!armRef.current) return
    const t = clock.getElapsedTime()
    armRef.current.rotation.x = -0.18 + Math.sin(t * 8.5) * 0.045
  })
  const cloth = '#1a1c22'
  const skin = '#c4a07a'
  return (
    <group position={[-0.92, 0.02, 0.38]} rotation={[0, 0.55, 0]}>
      <mesh position={[0, 0.62, 0]} castShadow>
        <capsuleGeometry args={[0.16, 0.42, 6, 12]} />
        <meshStandardMaterial color={cloth} roughness={0.82} metalness={0.08} />
      </mesh>
      <mesh position={[0, 1.02, 0.02]} castShadow>
        <sphereGeometry args={[0.145, 24, 24]} />
        <meshStandardMaterial color={skin} roughness={0.55} metalness={0.05} />
      </mesh>
      <mesh position={[0, 1.14, -0.02]}>
        <sphereGeometry args={[0.15, 16, 16, 0, Math.PI * 2, 0, 1.1]} />
        <meshStandardMaterial color="#161616" roughness={0.7} />
      </mesh>
      <group ref={armRef} position={[0.2, 0.72, 0.08]} rotation={[-0.2, 0.2, -0.4]}>
        <mesh position={[0, -0.16, 0.18]} castShadow>
          <capsuleGeometry args={[0.045, 0.28, 4, 8]} />
          <meshStandardMaterial color={cloth} roughness={0.8} />
        </mesh>
        <mesh position={[0.02, -0.28, 0.38]}>
          <boxGeometry args={[0.1, 0.04, 0.12]} />
          <meshStandardMaterial color={skin} roughness={0.5} />
        </mesh>
      </group>
      <mesh position={[-0.18, 0.68, 0.12]} rotation={[-0.55, -0.15, 0.5]} castShadow>
        <capsuleGeometry args={[0.045, 0.3, 4, 8]} />
        <meshStandardMaterial color={cloth} roughness={0.8} />
      </mesh>
    </group>
  )
}

function DeskSet({ clockRef }) {
  const { canvas, texture } = useMemo(() => createScreenTexture(), [])
  const lastMode = useRef('')
  const lastTyped = useRef('')

  useEffect(() => {
    drawLaptopScreen(canvas, { mode: 'intent', typed: '', time: 0 })
    texture.needsUpdate = true
  }, [canvas, texture])

  useFrame(() => {
    const t = clockRef.current
    const mode = screenMode(t)
    const typed = typedIntent(t)
    if (mode !== lastMode.current || typed !== lastTyped.current) {
      lastMode.current = mode
      lastTyped.current = typed
      drawLaptopScreen(canvas, { mode, typed, time: t })
      texture.needsUpdate = true
    }
  })

  return (
    <group>
      <mesh position={[0.15, -0.04, 0.15]} receiveShadow>
        <boxGeometry args={[3.4, 0.08, 1.7]} />
        <meshStandardMaterial color="#3a2a22" roughness={0.35} metalness={0.18} />
      </mesh>
      <mesh position={[0.15, -0.5, 0.15]}>
        <boxGeometry args={[3.15, 0.82, 1.5]} />
        <meshStandardMaterial color="#2a1f1a" roughness={0.5} metalness={0.1} />
      </mesh>

      <group position={[0.16, 0.02, 0.08]}>
        <RoundedBox args={[0.92, 0.03, 0.58]} radius={0.02} smoothness={4} castShadow>
          <meshPhysicalMaterial color="#c9ced6" metalness={0.92} roughness={0.22} clearcoat={0.6} />
        </RoundedBox>
        <mesh position={[0, 0.018, 0.04]}>
          <boxGeometry args={[0.78, 0.006, 0.42]} />
          <meshStandardMaterial color="#1a1d24" roughness={0.4} />
        </mesh>
        <group position={[0, 0.46, -0.26]} rotation={[-0.18, 0, 0]}>
          <RoundedBox args={[0.9, 0.58, 0.03]} radius={0.012} smoothness={4}>
            <meshPhysicalMaterial color="#9aa3b0" metalness={0.95} roughness={0.18} />
          </RoundedBox>
          <mesh position={[0, 0.01, 0.02]}>
            <planeGeometry args={[0.8, 0.5]} />
            <meshBasicMaterial map={texture} toneMapped={false} />
          </mesh>
          <mesh position={[0, 0.01, 0.018]}>
            <planeGeometry args={[0.82, 0.52]} />
            <MeshTransmissionMaterial
              samples={4}
              resolution={128}
              transmission={0.28}
              roughness={0.08}
              thickness={0.2}
              ior={1.5}
              chromaticAberration={0.015}
              color="#dbe4f0"
              toneMapped={false}
            />
          </mesh>
          <mesh position={[0, -0.3, 0.0]}>
            <boxGeometry args={[0.18, 0.03, 0.02]} />
            <meshStandardMaterial color="#111" metalness={0.7} roughness={0.3} emissive="#60a5fa" emissiveIntensity={0.25} />
          </mesh>
        </group>
      </group>

      <mesh position={[1.18, 0.42, 0.22]}>
        <cylinderGeometry args={[0.03, 0.05, 0.7, 12]} />
        <meshStandardMaterial color="#d7c4a3" metalness={0.6} roughness={0.3} />
      </mesh>
      <pointLight position={[1.18, 0.82, 0.22]} intensity={2.4} color="#ffc39a" distance={5} decay={2} />

      <SeatedFigure />
      <ContactShadows position={[0, -0.08, 0]} opacity={0.45} scale={6} blur={2.4} far={2.5} />
    </group>
  )
}

function DataStreams({ clockRef }) {
  const lineRef = useRef()
  const positions = useMemo(() => {
    const n = 240
    const arr = new Float32Array(n * 3)
    for (let i = 0; i < n; i++) {
      const z = -2 - (i / n) * 22
      arr[i * 3] = Math.sin(i * 0.37) * (0.4 + i * 0.004)
      arr[i * 3 + 1] = Math.cos(i * 0.21) * 0.25
      arr[i * 3 + 2] = z
    }
    return arr
  }, [])

  useFrame(({ clock }) => {
    const t = clockRef.current
    const vis = THREE.MathUtils.smoothstep(t, 3.4, 5.0) * (1 - THREE.MathUtils.smoothstep(t, 13.0, 14.6))
    if (!lineRef.current) return
    lineRef.current.material.opacity = vis * 0.22
    lineRef.current.rotation.z = clock.getElapsedTime() * 0.02
  })

  return (
    <points ref={lineRef} position={[0, 0.1, 0]}>
      <bufferGeometry>
        <bufferAttribute attach="attributes-position" count={positions.length / 3} array={positions} itemSize={3} />
      </bufferGeometry>
      <pointsMaterial color="#94a3b8" size={0.035} transparent opacity={0.18} sizeAttenuation depthWrite={false} />
    </points>
  )
}

function BlueFilament({ clockRef }) {
  const meshRef = useRef()
  const curve = useMemo(
    () =>
      new THREE.CatmullRomCurve3([
        new THREE.Vector3(0.16, 0.55, -0.35),
        new THREE.Vector3(0.05, 0.28, -4.2),
        new THREE.Vector3(-0.05, 0.12, -10),
        new THREE.Vector3(0, 0.18, -16.8),
      ]),
    [],
  )
  const geom = useMemo(() => new THREE.TubeGeometry(curve, 64, 0.028, 12, false), [curve])

  useFrame(() => {
    const t = clockRef.current
    const grow = THREE.MathUtils.smoothstep(t, 4.6, 6.35)
    const fade = 1 - THREE.MathUtils.smoothstep(t, 6.55, 7.3)
    if (!meshRef.current) return
    meshRef.current.scale.set(1, 1, Math.max(0.001, grow))
    meshRef.current.material.opacity = 0.95 * fade
    meshRef.current.material.emissiveIntensity = 3.2 * fade
  })

  return (
    <mesh ref={meshRef} geometry={geom} scale={[1, 1, 0.001]}>
      <meshStandardMaterial
        color="#60a5fa"
        emissive="#3b82f6"
        emissiveIntensity={3}
        transparent
        opacity={0.95}
        roughness={0.2}
        metalness={0.1}
        depthWrite={false}
      />
    </mesh>
  )
}

function RedRipple({ clockRef }) {
  const ref = useRef()
  useFrame(() => {
    const t = clockRef.current
    const u = THREE.MathUtils.clamp((t - 5.55) / 0.9, 0, 1)
    if (!ref.current) return
    const s = 0.2 + u * 4.8
    ref.current.scale.setScalar(s)
    ref.current.material.opacity = (1 - u) * 0.55
    ref.current.visible = t > 5.45 && t < 7.2
  })
  return (
    <mesh ref={ref} position={[0, 0.15, -8.4]} rotation={[Math.PI / 2, 0, 0]}>
      <torusGeometry args={[1, 0.035, 12, 80]} />
      <meshBasicMaterial color="#ef4444" transparent opacity={0} depthWrite={false} />
    </mesh>
  )
}

function ThreatParticles({ clockRef }) {
  const meshRef = useRef()
  const sim = useMemo(() => {
    const count = PARTICLE_COUNT
    const pos = new Float32Array(count * 3)
    const vel = new Float32Array(count * 3)
    const life = new Float32Array(count)
    const size = new Float32Array(count)
    const phase = new Uint8Array(count)
    for (let i = 0; i < count; i++) {
      pos[i * 3] = 0
      pos[i * 3 + 1] = 0.15
      pos[i * 3 + 2] = -6
      vel[i * 3] = 0
      vel[i * 3 + 1] = 0
      vel[i * 3 + 2] = 0
      life[i] = 0
      size[i] = 0.018 + Math.random() * 0.028
      phase[i] = 0
    }
    return { count, pos, vel, life, size, phase, spawned: false, shatterStarted: false }
  }, [])

  useFrame((_, dt) => {
    const t = clockRef.current
    const mesh = meshRef.current
    if (!mesh) return
    const { count, pos, vel, life, size, phase } = sim
    const step = Math.min(dt, 0.033)

    if (t >= 5.5 && !sim.spawned) {
      sim.spawned = true
      for (let i = 0; i < count; i++) {
        const along = Math.random()
        pos[i * 3] = (Math.random() - 0.5) * 0.12
        pos[i * 3 + 1] = 0.12 + (Math.random() - 0.5) * 0.1
        pos[i * 3 + 2] = -5.2 - along * 4.5
        const burst = 0.55 + Math.random() * 2.8
        vel[i * 3] = (Math.random() - 0.5) * burst
        vel[i * 3 + 1] = (Math.random() - 0.5) * burst * 0.85
        vel[i * 3 + 2] = -4.2 - Math.random() * 6.5
        phase[i] = 2
        life[i] = 1
      }
    }

    if (t < 5.45) {
      mesh.visible = false
      return
    }
    mesh.visible = true

    const impact = t >= 8.05
    for (let i = 0; i < count; i++) {
      if (phase[i] === 0) continue
      const ix = i * 3
      vel[ix + 1] -= 0.35 * step
      pos[ix] += vel[ix] * step
      pos[ix + 1] += vel[ix + 1] * step
      pos[ix + 2] += vel[ix + 2] * step

      if (phase[i] === 2 && impact) {
        const x = pos[ix]
        const y = pos[ix + 1]
        const z = pos[ix + 2]
        if (z <= GATE.z && z >= GATE.z - GATE.depth - 0.35 && Math.abs(x) < GATE.halfW && Math.abs(y) < GATE.halfH) {
          phase[i] = 3
          vel[ix] += (Math.random() - 0.5) * 7.5
          vel[ix + 1] += Math.random() * 5.2
          vel[ix + 2] = Math.abs(vel[ix + 2]) * 0.22 + Math.random() * 3.4
          life[i] = 0.55 + Math.random() * 0.9
          size[i] *= 0.55
          sim.shatterStarted = true
        }
      }

      if (phase[i] === 3) {
        vel[ix] *= 0.94
        vel[ix + 1] *= 0.94
        vel[ix + 2] *= 0.92
        life[i] -= step * 0.85
        size[i] *= 0.985
        if (life[i] <= 0) {
          phase[i] = 0
          size[i] = 0
        }
      }

      _dummy.position.set(pos[ix], pos[ix + 1], pos[ix + 2])
      const jagged = phase[i] === 2 ? 1 : Math.max(life[i], 0)
      const s = size[i] * (phase[i] === 2 ? 1.15 : jagged)
      _dummy.scale.setScalar(Math.max(0.0001, s * 18))
      _dummy.rotation.set(pos[ix] * 4, pos[ix + 2] * 2, pos[ix + 1] * 3)
      _dummy.updateMatrix()
      mesh.setMatrixAt(i, _dummy.matrix)
      if (phase[i] === 2) _color.copy(RED)
      else _color.copy(DUST).multiplyScalar(Math.max(life[i], 0))
      mesh.setColorAt(i, _color)
    }
    mesh.instanceMatrix.needsUpdate = true
    if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true
  })

  return (
    <instancedMesh ref={meshRef} args={[null, null, PARTICLE_COUNT]} frustumCulled={false}>
      <tetrahedronGeometry args={[0.055, 0]} />
      <meshStandardMaterial
        vertexColors
        emissive="#ef4444"
        emissiveIntensity={1.8}
        roughness={0.25}
        metalness={0.35}
        toneMapped={false}
      />
    </instancedMesh>
  )
}

function Hologram({ clockRef }) {
  const group = useRef()
  const htmlRef = useRef()
  useFrame(() => {
    const t = clockRef.current
    const vis = THREE.MathUtils.smoothstep(t, 8.55, 9.1) * (1 - THREE.MathUtils.smoothstep(t, 12.05, 13.1))
    if (!group.current) return
    group.current.visible = vis > 0.03
    group.current.scale.setScalar(0.9 + vis * 0.12)
    if (htmlRef.current) htmlRef.current.style.opacity = String(vis)
  })

  return (
    <group ref={group} position={[0, 1.85, -18.85]} visible={false}>
      <mesh>
        <planeGeometry args={[4.4, 1.85]} />
        <meshBasicMaterial color="#7f1d1d" transparent opacity={0.16} depthWrite={false} side={THREE.DoubleSide} />
      </mesh>
      <Html transform center distanceFactor={6.5} occlude={false} style={{ pointerEvents: 'none' }}>
        <div ref={htmlRef} className="cine-holo" style={{ opacity: 0 }}>
          <div className="cine-holo-kicker">Aegis Policy Gateway</div>
          <div className="cine-holo-title">QUANTITY CAP EXCEEDED</div>
          <div className="cine-holo-line">REQUEST BLOCKED</div>
          <div className="cine-holo-sub">FORWARDED TO APPROVAL QUEUE</div>
        </div>
      </Html>
    </group>
  )
}

function Gateway() {
  return (
    <group position={[0, 0.35, GATE.z]}>
      <mesh position={[-2.15, 0.1, 0]} castShadow>
        <boxGeometry args={[0.38, 4.6, 0.7]} />
        <meshPhysicalMaterial color="#8a919c" metalness={0.96} roughness={0.18} clearcoat={0.8} />
      </mesh>
      <mesh position={[2.15, 0.1, 0]} castShadow>
        <boxGeometry args={[0.38, 4.6, 0.7]} />
        <meshPhysicalMaterial color="#8a919c" metalness={0.96} roughness={0.18} clearcoat={0.8} />
      </mesh>
      <mesh position={[0, 2.45, 0]}>
        <boxGeometry args={[4.7, 0.32, 0.72]} />
        <meshPhysicalMaterial color="#9aa3b2" metalness={0.97} roughness={0.14} clearcoat={1} />
      </mesh>
      <mesh position={[0, 0.2, 0.02]}>
        <boxGeometry args={[3.65, 4.05, 0.18]} />
        <MeshTransmissionMaterial
          backside
          samples={6}
          resolution={256}
          transmission={0.92}
          roughness={0.08}
          thickness={1.35}
          ior={1.42}
          chromaticAberration={0.04}
          anisotropy={0.15}
          distortion={0.12}
          distortionScale={0.18}
          temporalDistortion={0.08}
          color="#d7e4f5"
        />
      </mesh>
      <mesh position={[0, -1.95, 0]}>
        <boxGeometry args={[4.8, 0.22, 1.1]} />
        <meshStandardMaterial color="#6b7280" metalness={0.9} roughness={0.25} />
      </mesh>
    </group>
  )
}

function SceneLights({ clockRef }) {
  const warm = useRef()
  const cool = useRef()
  const gate = useRef()
  useFrame(() => {
    const t = clockRef.current
    const k = THREE.MathUtils.smoothstep(t, 3.1, 4.8) * (1 - THREE.MathUtils.smoothstep(t, 12.8, 14.5))
    if (warm.current) warm.current.intensity = THREE.MathUtils.lerp(1.6, 0.15, k)
    if (cool.current) cool.current.intensity = THREE.MathUtils.lerp(0.2, 1.1, k)
    if (gate.current) gate.current.intensity = THREE.MathUtils.lerp(0, 2.4, THREE.MathUtils.smoothstep(t, 7.6, 8.6))
  })
  return (
    <>
      <hemisphereLight args={['#ffd7b8', '#1a120e', 0.55]} />
      <ambientLight intensity={0.18} />
      <spotLight ref={warm} position={[1.4, 3.2, 2.2]} intensity={1.6} color="#ffc8a0" angle={0.55} penumbra={0.8} castShadow />
      <directionalLight ref={cool} position={[-4, 4, -8]} intensity={0.2} color="#93c5fd" />
      <pointLight ref={gate} position={[0, 1.2, GATE.z + 1.2]} intensity={0} color="#e8eef7" distance={12} />
      <pointLight position={[0, 0.4, -10]} intensity={0.6} color="#3b82f6" distance={14} />
    </>
  )
}

function PostFX({ dofRef }) {
  return (
    <EffectComposer disableNormalPass>
      <DepthOfField ref={dofRef} focusDistance={0.02} focalLength={0.02} bokehScale={1.5} height={360} />
      <Bloom intensity={0.38} luminanceThreshold={0.7} luminanceSmoothing={0.22} mipmapBlur />
      <ChromaticAberration offset={[0.00025, 0.0004]} />
      <Vignette eskil={false} offset={0.18} darkness={0.7} />
    </EffectComposer>
  )
}

export function CinematicWorld({ clockRef }) {
  const dofRef = useRef()
  return (
    <>
      <CameraDirector clockRef={clockRef} dofRef={dofRef} />
      <WorldAtmosphere clockRef={clockRef} />
      <SceneLights clockRef={clockRef} />
      <Environment preset="studio" />
      <DeskSet clockRef={clockRef} />
      <DataStreams clockRef={clockRef} />
      <BlueFilament clockRef={clockRef} />
      <RedRipple clockRef={clockRef} />
      <ThreatParticles clockRef={clockRef} />
      <Gateway />
      <Hologram clockRef={clockRef} />
      <PostFX dofRef={dofRef} />
    </>
  )
}

export { DURATION }
