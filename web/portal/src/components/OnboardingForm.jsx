import React, { useState, useEffect } from 'react'
import { useToast } from '../App'
import ResultCard from './ResultCard'

const PLATFORMS = [
  {
    value: 'shopify',
    label: 'Shopify',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
        <path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
        <polyline points="9 22 9 12 15 12 15 22" />
      </svg>
    ),
    color: '#95bf47',
  },
  {
    value: 'woocommerce',
    label: 'WooCommerce',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
        <circle cx="12" cy="12" r="10" />
        <path d="M14.31 8l5.74 9.94M9.69 8h11.48M7.38 12l5.74-9.94M9.69 16L3.95 6.06M14.31 16H2.83M16.62 12l-5.74 9.94" />
      </svg>
    ),
    color: '#9b5c8f',
  },
  {
    value: 'custom',
    label: 'Custom API',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
        <polyline points="16 18 22 12 16 6" />
        <polyline points="8 6 2 12 8 18" />
        <line x1="14" y1="4" x2="10" y2="20" />
      </svg>
    ),
    color: '#06b6d4',
  },
]

export default function OnboardingForm() {
  const [name, setName] = useState('')
  const [platform, setPlatform] = useState('shopify')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState(null)
  const [formVisible, setFormVisible] = useState(false)
  const { showToast } = useToast()

  // Animation for form entrance
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setTimeout(() => setFormVisible(true), 200)
          }
        })
      },
      { threshold: 0.2 }
    )
    const section = document.getElementById('onboard')
    if (section) observer.observe(section)
    return () => observer.disconnect()
  }, [])

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!name.trim()) return
    setLoading(true)
    try {
      const res = await fetch('/api/merchants', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name.trim(), platform }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.error || 'Failed to provision')
      setResult(data)
      showToast('Merchant provisioned successfully!')
    } catch (err) {
      showToast(err.message, 'error')
    } finally {
      setLoading(false)
    }
  }

  const handleReset = () => {
    setResult(null)
    setName('')
  }

  if (result) {
    return (
      <section className="onboarding-section" id="onboard">
        <ResultCard result={result} onReset={handleReset} />
      </section>
    )
  }

  return (
    <section className="onboarding-section" id="onboard">
      <form className={`onboarding-card ${formVisible ? 'visible' : ''}`} onSubmit={handleSubmit} style={{ position: 'relative', zIndex: 2 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '8px' }}>
          <div style={{
            width: '48px', height: '48px', borderRadius: '14px',
            background: 'linear-gradient(135deg, #a855f7, #06b6d4)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            boxShadow: '0 8px 24px rgba(168, 85, 247, 0.4)',
          }}>
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5">
              <path d="M12 5v14M5 12h14" />
            </svg>
          </div>
          <div>
            <h2 className="onboarding-title">Bring your store to AI buyers</h2>
          </div>
        </div>
        <p className="onboarding-subtitle">
          Connect your commerce platform and get a dedicated Merchant MCP endpoint for your store.
        </p>

        <div className="form-group">
          <label className="form-label">Store Name</label>
          <input
            type="text"
            className="form-input"
            placeholder="e.g., Kicks & Co."
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            id="store-name-input"
            autoFocus
          />
        </div>

        <div className="form-group">
          <label className="form-label">Platform</label>
          <div className="platform-tiles">
            {PLATFORMS.map((p) => (
              <div
                key={p.value}
                className={`platform-tile ${platform === p.value ? 'selected' : ''}`}
                onClick={() => setPlatform(p.value)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => e.key === 'Enter' && setPlatform(p.value)}
                id={`platform-tile-${p.value}`}
                style={{ borderLeft: platform === p.value ? `4px solid ${p.color}` : 'none' }}
              >
                <span className="platform-tile-icon" style={{ color: p.color }}>{p.icon}</span>
                {p.label}
              </div>
            ))}
          </div>
        </div>

        <button
          type="submit"
          className="btn btn-primary"
          style={{ width: '100%', marginTop: '8px' }}
          disabled={loading || !name.trim()}
          id="connect-store-btn"
        >
          {loading ? (
            <>
              <span className="spinner" style={{ width: 18, height: 18 }} />
              Provisioning...
            </>
          ) : (
            <>
              Connect Store
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                <path d="M5 12h14M12 5l7 7-7 7" />
              </svg>
            </>
          )}
        </button>
      </form>
    </section>
  )
}