import React, { useEffect, useRef } from 'react'

const NODES = [
  {
    label: 'AI Agent',
    sub: 'MCP Client',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 2a4 4 0 014 4v2a4 4 0 01-8 0V6a4 4 0 014-4z" />
        <path d="M6 21v-2a4 4 0 014-4h4a4 4 0 014 4v2" />
        <circle cx="12" cy="6" r="1" fill="currentColor" stroke="none" />
      </svg>
    ),
  },
  {
    label: 'Merchant MCP',
    sub: 'Tools API',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <rect x="2" y="3" width="20" height="14" rx="2" />
        <path d="M8 21h8M12 17v4" />
      </svg>
    ),
  },
  {
    label: 'Policy Gateway',
    sub: 'Aegis Engine',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    ),
  },
  {
    label: 'Razorpay',
    sub: 'Payment Rails',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <rect x="1" y="4" width="22" height="16" rx="2" ry="2" />
        <path d="M1 10h22" />
      </svg>
    ),
  },
  {
    label: 'Audit Log',
    sub: 'Hash-Chained',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71" />
        <path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71" />
      </svg>
    ),
  },
]

export default function ArchitectureDiagram() {
  const sectionRef = useRef()
  const nodesRef = useRef([])
  const arrowsRef = useRef([])

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const nodes = nodesRef.current
            const arrows = arrowsRef.current
            nodes.forEach((node, i) => {
              if (node) {
                setTimeout(() => node.classList.add('visible'), i * 150)
              }
            })
            arrows.forEach((arrow, i) => {
              if (arrow) {
                setTimeout(() => arrow.classList.add('visible'), i * 150 + 100)
              }
            })
            observer.disconnect()
          }
        })
      },
      { threshold: 0.3 }
    )
    if (sectionRef.current) observer.observe(sectionRef.current)
    return () => observer.disconnect()
  }, [])

  return (
    <section className="arch-section" ref={sectionRef} id="architecture">
      <div className="section-label">Architecture</div>
      <h2 className="section-title">Every Purchase, Verified</h2>
      <p className="section-subtitle">
        AI agent requests flow through deterministic policy checks before any payment is attempted.
        Every decision is hash-chained.
      </p>
      <div className="arch-flow">
        {NODES.map((node, i) => (
          <React.Fragment key={node.label}>
            <div
              className="arch-node"
              ref={(el) => (nodesRef.current[i] = el)}
            >
              <div className="arch-node-icon">{node.icon}</div>
              <div className="arch-node-label">{node.label}</div>
              <div className="arch-node-sub">{node.sub}</div>
            </div>
            {i < NODES.length - 1 && (
              <div
                className="arch-arrow"
                ref={(el) => (arrowsRef.current[i] = el)}
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M5 12h14M12 5l7 7-7 7" />
                </svg>
              </div>
            )}
          </React.Fragment>
        ))}
      </div>
    </section>
  )
}
