import React, { useState } from 'react'
import { useToast, useAuth } from '../App'
import Navbar from './Navbar'

export default function LoginGate({ onLogin, title = "Admin Access Required", subtitle = "Enter your admin key to access this page." }) {
  const [adminKey, setAdminKey] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const { login } = useAuth()
  const { showToast } = useToast()

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!adminKey.trim()) return
    setLoading(true)
    setError('')
    try {
      await login(adminKey.trim())
      showToast('Authenticated successfully')
      if (onLogin) onLogin()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="merchants-page">
      <Navbar />
      <div className="container" style={{ display: 'flex', justifyContent: 'center', paddingTop: '80px' }}>
        <div className="onboarding-card visible" style={{ maxWidth: '420px', width: '100%', opacity: 1, transform: 'none' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '8px' }}>
            <div style={{
              width: '40px', height: '40px', borderRadius: '12px',
              background: 'rgba(168,85,247,0.12)', border: '1px solid rgba(168,85,247,0.25)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: 'var(--accent-purple)',
            }}>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                <path d="M7 11V7a5 5 0 0110 0v4" />
              </svg>
            </div>
            <div>
              <h2 className="onboarding-title" style={{ fontSize: '20px', marginBottom: '0' }}>{title}</h2>
            </div>
          </div>
          <p className="onboarding-subtitle" style={{ marginBottom: '20px' }}>
            {subtitle}
          </p>

          <div style={{
            display: 'flex', alignItems: 'center', justifyBetween: 'space-between',
            padding: '10px 14px', borderRadius: '10px', marginBottom: '20px',
            background: 'rgba(168,85,247,0.06)', border: '1px solid rgba(168,85,247,0.18)',
            fontSize: '12px', color: 'var(--text-secondary)'
          }}>
            <span>Default Admin Key: <code style={{ color: 'var(--text-primary)', fontWeight: 600 }}>nexus_admin_default</code></span>
            <button
              type="button"
              className="btn btn-ghost"
              style={{ padding: '2px 8px', fontSize: '11px', height: 'auto', marginLeft: 'auto' }}
              onClick={() => setAdminKey('nexus_admin_default')}
            >
              Fill Key
            </button>
          </div>

          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label className="form-label">Admin Key</label>
              <input
                type="password"
                className="form-input"
                placeholder="Enter admin key (e.g. nexus_admin_default)"
                value={adminKey}
                onChange={(e) => setAdminKey(e.target.value)}
                required
                autoFocus
                id="admin-key-input"
              />
            </div>

            {error && (
              <div style={{
                padding: '10px 14px', borderRadius: '10px', marginBottom: '16px',
                background: 'rgba(239,68,68,0.08)', border: '1px solid rgba(239,68,68,0.2)',
                color: 'var(--accent-red)', fontSize: '13px', fontWeight: 500,
              }}>
                {error}
              </div>
            )}

            <button
              type="submit"
              className="btn btn-primary"
              style={{ width: '100%' }}
              disabled={loading || !adminKey.trim()}
              id="admin-login-btn"
            >
              {loading ? (
                <>
                  <span className="spinner" style={{ width: 18, height: 18 }} />
                  Authenticating...
                </>
              ) : (
                <>
                  Authenticate
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                    <path d="M5 12h14M12 5l7 7-7 7" />
                  </svg>
                </>
              )}
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
