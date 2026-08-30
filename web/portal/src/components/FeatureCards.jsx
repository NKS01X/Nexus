import React from 'react'

const FEATURES = [
  {
    id: 'mcp-storefront',
    color: '#a855f7',
    title: 'Your store, available to AI',
    desc: 'Expose search, product details, availability, purchase, and order status as standard MCP tools.',
    icon: (
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
        <rect x="2" y="3" width="20" height="14" rx="2" />
        <path d="M8 21h8M12 17v4" />
      </svg>
    ),
  },
  {
    id: 'policy-engine',
    color: '#06b6d4',
    title: 'Stop unsafe orders before payment',
    desc: 'Aegis enforces spend caps, quantity limits, velocity controls, category rules, SKU blocks, and geography before an order reaches your payment rails.',
    icon: (
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    ),
  },
  {
    id: 'audit-trail',
    color: '#f59e0b',
    title: 'Review exceptions with confidence',
    desc: 'Prompt injection and abusive requests are blocked or forwarded to your approval queue, with every decision preserved in a tamper-evident audit trail.',
    icon: (
      <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
        <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71" />
        <path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71" />
      </svg>
    ),
  },
]

export default function FeatureCards() {
  return (
    <section className="features-section" id="features">
      <div className="features-heading">
        <div className="section-label">For modern merchants</div>
        <h2 className="section-title">Everything you need to sell through AI</h2>
        <p className="section-subtitle">
          One merchant-facing integration for discovery, conversion, policy control, and operational confidence.
        </p>
      </div>

      <div className="features-grid">
        {FEATURES.map((feature) => (
          <article className="feature-card visible" key={feature.id}>
            <div className="feature-icon" style={{ color: feature.color }}>
              {feature.icon}
            </div>
            <h3 className="feature-title">{feature.title}</h3>
            <p className="feature-desc">{feature.desc}</p>
          </article>
        ))}
      </div>
    </section>
  )
}
