import React, { useEffect, useRef } from 'react'

const FEATURES = [
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <rect x="2" y="3" width="20" height="14" rx="2" />
        <path d="M8 21h8M12 17v4" />
      </svg>
    ),
    title: 'MCP-Native Storefront',
    desc: 'search_products, purchase, get_order_status — exposed as standard MCP tools any AI agent can call.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    ),
    title: 'Deterministic Policy Engine',
    desc: 'Spend caps, velocity limits, category allowlists, SKU blocklists — all compiled Go, no LLM in the enforcement path.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71" />
        <path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71" />
      </svg>
    ),
    title: 'Hash-Chained Audit Trail',
    desc: 'Every decision — allowed or blocked — appended to a SHA-256 linked log. Tamper-evident and verifiable on demand.',
  },
]

export default function FeatureCards() {
  const cardsRef = useRef([])

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('visible')
          }
        })
      },
      { threshold: 0.15, rootMargin: '0px 0px -60px 0px' }
    )
    cardsRef.current.forEach((card) => {
      if (card) observer.observe(card)
    })
    return () => observer.disconnect()
  }, [])

  return (
    <section className="features-section" id="features">
      <div style={{ textAlign: 'center' }}>
        <div className="section-label">Capabilities</div>
        <h2 className="section-title">Built for Agent Commerce</h2>
        <p className="section-subtitle">
          Three layers of infrastructure that make AI purchasing safe, auditable, and merchant-controlled.
        </p>
      </div>
      <div className="features-grid">
        {FEATURES.map((f, i) => (
          <div
            key={f.title}
            className="feature-card"
            ref={(el) => (cardsRef.current[i] = el)}
            style={{ transitionDelay: `${i * 100}ms` }}
          >
            <div className="feature-icon">{f.icon}</div>
            <div className="feature-title">{f.title}</div>
            <div className="feature-desc">{f.desc}</div>
          </div>
        ))}
      </div>
    </section>
  )
}
