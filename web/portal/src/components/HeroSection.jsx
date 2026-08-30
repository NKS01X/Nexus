import React from 'react'
import HeroScene from './HeroScene'

export default function HeroSection({ onScrollToForm }) {
  return (
    <section className="hero-wrapper" id="hero">
      <HeroScene />
      <div className="hero-content">
        <div className="hero-badge">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
          </svg>
          Merchant MCP for AI commerce
        </div>
        <h1 className="hero-title">
          Make Your Store{' '}
          <span className="gradient-text">Transactable by AI</span>
        </h1>
        <p className="hero-subtitle">
          Expose your catalog as MCP tools, enforce deterministic purchase policies,
          and maintain a tamper-evident audit trail — in 60 seconds.
        </p>
        <div className="hero-ctas">
          <button className="btn btn-primary" onClick={onScrollToForm}>
            Connect Your Store
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <path d="M5 12h14M12 5l7 7-7 7" />
            </svg>
          </button>
          <a href="https://github.com/nickhil-razorpay/nexus" target="_blank" rel="noopener noreferrer" className="btn btn-ghost">
            View on GitHub
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6M15 3h6v6M10 14L21 3" />
            </svg>
          </a>
        </div>
      </div>
      <div className="scroll-indicator">
        <span>Scroll</span>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M12 5v14M5 12l7 7 7-7" />
        </svg>
      </div>
    </section>
  )
}
