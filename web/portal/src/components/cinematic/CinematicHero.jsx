import React, { Suspense, useRef } from 'react'
import { Canvas, useFrame } from '@react-three/fiber'
import * as THREE from 'three'
import { CinematicWorld, DURATION } from './CinematicScene'

function ClockDriver({ clockRef }) {
  useFrame((_, dt) => {
    clockRef.current += Math.min(dt, 0.05)
    if (clockRef.current >= DURATION) {
      clockRef.current = 0
    }
  })
  return null
}

export default function CinematicHero({ onScrollToForm }) {
  const clockRef = useRef(0)

  return (
    <section className="cine-hero" id="hero">
      <div className="cine-stage">
        <Canvas
          dpr={[0.75, 1.1]}
          camera={{ position: [0.68, 1.22, 1.62], fov: 36, near: 0.05, far: 80 }}
          gl={{
            antialias: false,
            toneMapping: THREE.ACESFilmicToneMapping,
            toneMappingExposure: 0.96,
            powerPreference: 'high-performance',
          }}
        >
          <Suspense fallback={null}>
            <ClockDriver clockRef={clockRef} />
            <CinematicWorld clockRef={clockRef} />
          </Suspense>
        </Canvas>
      </div>

      <div className="cine-content">
        <p className="cine-kicker">Merchant MCP with policy enforcement</p>
        <h1>
          Sell through AI.
          <em> Stop malicious orders.</em>
        </h1>
        <p className="cine-description">
          Expose your catalog, inventory, and checkout as trusted MCP tools.
          Aegis checks quantity, spend, velocity, category, and location before payment,
          blocking prompt-injected or abusive requests before they reach your rails.
        </p>
        <div className="cine-actions">
          <button type="button" className="btn btn-primary" onClick={onScrollToForm}>
            Connect your store
          </button>
          <a className="btn btn-ghost" href="#architecture">See the flow</a>
        </div>
        <div className="cine-proof" aria-label="Aegis platform capabilities">
          <span><i className="proof-dot proof-dot-cyan" /> MCP-native storefront</span>
          <span><i className="proof-dot proof-dot-green" /> Rules before payment</span>
          <span><i className="proof-dot proof-dot-amber" /> Human review when needed</span>
        </div>
      </div>
    </section>
  )
}
