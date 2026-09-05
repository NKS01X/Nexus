import React, { useState, useRef, useEffect, useMemo } from 'react'
import { useToast, safeJsonParse } from '../App'
import Navbar from './Navbar'
import { Canvas, useFrame, useThree } from '@react-three/fiber'
import { Float, Stars, Line } from '@react-three/drei'
import * as THREE from 'three'
import gsap from 'gsap'
import { LazyCanvas } from './CanvasWrapper'

// Lightweight inline markdown renderer — no external deps
function MarkdownView({ text }) {
  if (!text) return null

  // Escape HTML to prevent XSS
  const esc = (s) => s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')

  // Inline formats: **bold**, *italic*, `code`
  const inlineFormat = (raw) =>
    esc(raw)
      .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.+?)\*/g, '<em>$1</em>')
      .replace(/`([^`]+)`/g, '<code style="background:rgba(168,85,247,0.15);padding:1px 5px;border-radius:4px;font-family:monospace;font-size:12px">$1</code>')

  const lines = text.split('\n')
  const blocks = []
  let i = 0
  while (i < lines.length) {
    const line = lines[i]

    // Heading
    const hMatch = line.match(/^(#{1,3})\s+(.+)/)
    if (hMatch) {
      const level = hMatch[1].length
      const tag = `h${level + 2}` // h3-h5 so it doesn't clash with page h1
      blocks.push(<div key={i} dangerouslySetInnerHTML={{ __html: `<${tag} style="font-weight:700;margin:14px 0 6px;color:var(--text-primary)">${inlineFormat(hMatch[2])}</${tag}>` }} />)
      i++
      continue
    }

    // HR
    if (/^(-{3,}|\*{3,}|_{3,})$/.test(line.trim())) {
      blocks.push(<hr key={i} style={{ border: 'none', borderTop: '1px solid var(--border)', margin: '12px 0' }} />)
      i++
      continue
    }

    // Table: detect header row followed by separator
    if (i + 1 < lines.length && /^\|/.test(line) && /^\|[-:\s|]+\|/.test(lines[i + 1])) {
      const tableLines = []
      while (i < lines.length && lines[i].trim().startsWith('|')) {
        tableLines.push(lines[i])
        i++
      }
      const parseRow = (r) => r.split('|').slice(1, -1).map(c => c.trim())
      const headers = parseRow(tableLines[0])
      const rows = tableLines.slice(2).map(parseRow)
      blocks.push(
        <div key={i} style={{ overflowX: 'auto', margin: '12px 0' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px' }}>
            <thead>
              <tr>
                {headers.map((h, hi) => (
                  <th key={hi} dangerouslySetInnerHTML={{ __html: inlineFormat(h) }} style={{ padding: '8px 12px', textAlign: 'left', background: 'rgba(168,85,247,0.15)', color: 'var(--text-primary)', fontWeight: 700, borderBottom: '1px solid var(--border)', whiteSpace: 'nowrap' }} />
                ))}
              </tr>
            </thead>
            <tbody>
              {rows.map((row, ri) => (
                <tr key={ri} style={{ borderBottom: '1px solid rgba(255,255,255,0.05)' }}>
                  {row.map((cell, ci) => (
                    <td key={ci} dangerouslySetInnerHTML={{ __html: inlineFormat(cell) }} style={{ padding: '7px 12px', color: 'var(--text-secondary)', verticalAlign: 'top' }} />
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )
      continue
    }

    // Bullet list item
    const liMatch = line.match(/^[\s]*[-*+]\s+(.+)/)
    if (liMatch) {
      const items = []
      while (i < lines.length && /^[\s]*[-*+]\s+/.test(lines[i])) {
        const m = lines[i].match(/^[\s]*[-*+]\s+(.+)/)
        items.push(m ? m[1] : lines[i])
        i++
      }
      blocks.push(
        <ul key={i} style={{ paddingLeft: '20px', margin: '8px 0', listStyleType: 'disc' }}>
          {items.map((item, ii) => (
            <li key={ii} dangerouslySetInnerHTML={{ __html: inlineFormat(item) }} style={{ color: 'var(--text-secondary)', margin: '4px 0', lineHeight: 1.6 }} />
          ))}
        </ul>
      )
      continue
    }

    // Numbered list
    const nlMatch = line.match(/^[\s]*\d+\.\s+(.+)/)
    if (nlMatch) {
      const items = []
      while (i < lines.length && /^[\s]*\d+\.\s+/.test(lines[i])) {
        const m = lines[i].match(/^[\s]*\d+\.\s+(.+)/)
        items.push(m ? m[1] : lines[i])
        i++
      }
      blocks.push(
        <ol key={i} style={{ paddingLeft: '20px', margin: '8px 0' }}>
          {items.map((item, ii) => (
            <li key={ii} dangerouslySetInnerHTML={{ __html: inlineFormat(item) }} style={{ color: 'var(--text-secondary)', margin: '4px 0', lineHeight: 1.6 }} />
          ))}
        </ol>
      )
      continue
    }

    // Blank line
    if (line.trim() === '') {
      blocks.push(<div key={i} style={{ height: '8px' }} />)
      i++
      continue
    }

    // Paragraph
    blocks.push(
      <p key={i} dangerouslySetInnerHTML={{ __html: inlineFormat(line) }} style={{ color: 'var(--text-secondary)', lineHeight: 1.7, margin: '4px 0' }} />
    )
    i++
  }
  return <div style={{ fontSize: '14px' }}>{blocks}</div>
}

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
      const data = await safeJsonParse(resp)
      if (!resp.ok) throw new Error(data.error || data.message || 'Groq error')
      setSuggestion(data.completion)
      showToast('Got AI suggestion', 'success')
    } catch (e) {
      console.error(e)
      showToast(e.message || 'Failed to get suggestion', 'error')
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
      const resp = await fetch('/api/ai/purchase', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          prompt,
          store_id: storeId,
          buyer_id: 'demo-buyer',
        }),
      })

      const text = await resp.text()
      let data
      try { data = JSON.parse(text) } catch { data = { error: text || `HTTP ${resp.status}` } }

      if (!resp.ok) {
        showToast(`Error: ${data.error || resp.statusText}`, 'error')
        setResult({ error: data.error || resp.statusText })
        return
      }

      setResult(data)
      if (data.llm_decision === 'no_purchase') {
        showToast('LLM declined to purchase', 'info')
      } else {
        showToast('AI purchase attempt processed', 'success')
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
    <div className="merchants-page">
      <Navbar />
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
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#a855f7" strokeWidth="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4M12 8h.01"/></svg>
                <strong style={{ color: 'var(--accent-purple)', fontSize: '13px', letterSpacing: '0.05em', textTransform: 'uppercase' }}>AI Suggestion</strong>
              </div>
              <MarkdownView text={suggestion} />
            </div>
          )}

          {result && (() => {
            // Handle LLM-declined case
            if (result.llm_decision === 'no_purchase') {
              return (
                <div style={{ marginTop: '24px', padding: '24px', borderRadius: '16px', background: 'rgba(245,158,11,0.08)', border: '1px solid rgba(245,158,11,0.3)' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '12px' }}>
                    <div style={{ width: '36px', height: '36px', borderRadius: '50%', background: 'rgba(245,158,11,0.2)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" strokeWidth="2.5"><path d="M12 9v4M12 17h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/></svg>
                    </div>
                    <div style={{ fontWeight: 700, fontSize: '16px', color: '#f59e0b' }}>LLM Declined Purchase</div>
                  </div>
                  <p style={{ color: 'var(--text-secondary)', fontSize: '14px', lineHeight: 1.6 }}>{result.llm_reply}</p>
                  <button className="btn btn-secondary" style={{ marginTop: '20px', width: '100%' }} onClick={reset}>Try Another</button>
                </div>
              )
            }

            // Handle agentic purchase attempt
            const r = result.result || result || {}
            const allowed = r.allowed ?? r.Allowed
            const reason = r.reason ?? r.Reason ?? ''
            const status = r.status ?? r.Status ?? ''
            const orderId = r.order_id ?? r.OrderID ?? ''
            const ruleFired = r.rule_fired ?? r.RuleFired ?? ''
            const llmArgs = result.llm_args || {}

            return (
              <div style={{ marginTop: '24px', padding: '24px', borderRadius: '16px', background: allowed ? 'rgba(16,185,129,0.08)' : 'rgba(239,68,68,0.08)', border: `1px solid ${allowed ? 'rgba(16,185,129,0.3)' : 'rgba(239,68,68,0.3)'}` }}>
                {/* Status badge */}
                <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '16px' }}>
                  <div style={{ width: '36px', height: '36px', borderRadius: '50%', background: allowed ? 'rgba(16,185,129,0.2)' : 'rgba(239,68,68,0.2)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                    {allowed ? (
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2.5"><path d="M20 6L9 17l-5-5"/></svg>
                    ) : (
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#ef4444" strokeWidth="2.5"><path d="M18 6L6 18M6 6l12 12"/></svg>
                    )}
                  </div>
                  <div>
                    <div style={{ fontWeight: 700, fontSize: '16px', color: allowed ? '#10b981' : '#ef4444' }}>
                      {allowed ? 'Purchase Approved' : 'Purchase Blocked'}
                    </div>
                    <div style={{ fontSize: '12px', color: 'var(--text-tertiary)', marginTop: '2px' }}>{status}</div>
                  </div>
                </div>

                {/* LLM decision */}
                {(llmArgs.sku || llmArgs.quantity) && (
                  <div style={{ padding: '10px 12px', background: 'rgba(168,85,247,0.08)', borderRadius: '8px', marginBottom: '8px', fontSize: '13px', border: '1px solid rgba(168,85,247,0.2)' }}>
                    <span style={{ color: 'var(--accent-purple)', fontWeight: 600 }}>🤖 LLM decided: </span>
                    <span style={{ fontFamily: 'monospace', color: 'var(--text-secondary)' }}>
                      buy {llmArgs.quantity}× <strong style={{ color: 'var(--text-primary)' }}>{llmArgs.sku}</strong>
                    </span>
                  </div>
                )}

                {/* Details */}
                <div style={{ display: 'grid', gap: '8px', fontSize: '13px' }}>
                  {reason && (
                    <div style={{ display: 'flex', gap: '8px', padding: '10px 12px', background: 'rgba(255,255,255,0.04)', borderRadius: '8px' }}>
                      <span style={{ color: 'var(--text-tertiary)', flexShrink: 0 }}>Reason:</span>
                      <span style={{ color: 'var(--text-secondary)' }}>{reason}</span>
                    </div>
                  )}
                  {ruleFired && ruleFired !== 'none' && (
                    <div style={{ display: 'flex', gap: '8px', padding: '10px 12px', background: 'rgba(255,255,255,0.04)', borderRadius: '8px' }}>
                      <span style={{ color: 'var(--text-tertiary)', flexShrink: 0 }}>Rule fired:</span>
                      <span style={{ fontFamily: 'monospace', color: '#f59e0b', fontSize: '12px' }}>{ruleFired}</span>
                    </div>
                  )}
                  {orderId && (
                    <div style={{ display: 'flex', gap: '8px', padding: '10px 12px', background: 'rgba(16,185,129,0.08)', borderRadius: '8px' }}>
                      <span style={{ color: 'var(--text-tertiary)', flexShrink: 0 }}>Order ID:</span>
                      <span style={{ fontFamily: 'monospace', color: '#10b981', fontSize: '12px' }}>{orderId}</span>
                    </div>
                  )}
                </div>

                <button className="btn btn-secondary" style={{ marginTop: '20px', width: '100%' }} onClick={reset}>Try Another</button>
              </div>
            )
          })()}
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
  </div>
  )
}