import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useToast, useAuth } from '../App'
import Navbar from './Navbar'

function LoginGate({ onLogin }) {
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
        <div className="onboarding-card" style={{ maxWidth: '420px', width: '100%' }}>
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
              <h2 className="onboarding-title" style={{ fontSize: '20px', marginBottom: '0' }}>Admin Access</h2>
            </div>
          </div>
          <p className="onboarding-subtitle" style={{ marginBottom: '28px' }}>
            Enter your admin key to view provisioned merchants.
          </p>

          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label className="form-label">Admin Key</label>
              <input
                type="password"
                className="form-input"
                placeholder="Enter admin key"
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

export default function MerchantsPage() {
  const [merchants, setMerchants] = useState([])
  const [loading, setLoading] = useState(true)
  const { showToast } = useToast()
  const { isAuthenticated, authFetch, logout } = useAuth()

  useEffect(() => {
    if (isAuthenticated) {
      fetchMerchants()
    }
  }, [isAuthenticated])

  const fetchMerchants = async () => {
    setLoading(true)
    try {
      const res = await authFetch('/api/merchants')
      const data = await res.json()
      setMerchants(Array.isArray(data) ? data : [])
    } catch (err) {
      showToast(err.message || 'Failed to load merchants', 'error')
    } finally {
      setLoading(false)
    }
  }

  const copyToClipboard = (text, label) => {
    navigator.clipboard.writeText(text)
    showToast(`${label} copied!`)
  }

  // Show login gate if not authenticated
  if (!isAuthenticated) {
    return <LoginGate onLogin={() => {}} />
  }

  if (loading) {
    return (
      <div className="merchants-page">
        <Navbar />
        <div className="container">
          <div className="loading"><div className="spinner" /></div>
        </div>
      </div>
    )
  }

  return (
    <div className="merchants-page">
      <Navbar />
      <div className="container">
        <div className="page-header">
          <div className="page-header-row">
            <div>
              <h1 className="page-title">Provisioned Merchants</h1>
              <p className="page-subtitle">
                {merchants.length} store{merchants.length !== 1 ? 's' : ''} connected to Aegis Merchant MCP
              </p>
            </div>
            <div style={{ display: 'flex', gap: '12px' }}>
              <button className="btn btn-ghost" onClick={logout} id="logout-btn">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4M16 17l5-5-5-5M21 12H9" />
                </svg>
                Logout
              </button>
              <Link to="/">
                <button className="btn btn-primary">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                    <path d="M12 5v14M5 12h14" />
                  </svg>
                  Add Merchant
                </button>
              </Link>
            </div>
          </div>
        </div>

        {merchants.length === 0 ? (
          <div className="card">
            <div className="empty-state">
              <div className="empty-icon">
                <svg width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2" />
                  <circle cx="9" cy="7" r="4" />
                  <path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75" />
                </svg>
              </div>
              <h2 className="empty-title">No merchants yet</h2>
              <p className="empty-desc">Provision your first merchant to get started</p>
              <Link to="/">
                <button className="btn btn-primary">Provision Merchant</button>
              </Link>
            </div>
          </div>
        ) : (
          <div className="card">
            <div style={{ overflowX: 'auto' }}>
              <table className="merchants-table">
                <thead>
                  <tr>
                    <th>Store ID</th>
                    <th>Name</th>
                    <th>Platform</th>
                    <th>MCP Endpoint</th>
                    <th>API Key</th>
                    <th>Status</th>
                    <th>Created</th>
                  </tr>
                </thead>
                <tbody>
                  {merchants.map((m) => (
                    <tr key={m.id}>
                      <td>
                        <code style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>{m.id}</code>
                      </td>
                      <td style={{ fontWeight: 600 }}>{m.name}</td>
                      <td>
                        <span className={`platform-badge ${m.platform}`}>{m.platform}</span>
                      </td>
                      <td>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                          <span style={{ fontFamily: "'SF Mono', 'Fira Code', monospace", fontSize: '13px', color: 'var(--text-secondary)' }}>
                            /mcp/{m.id}
                          </span>
                          <button
                            className="copy-btn"
                            onClick={() => copyToClipboard(m.mcp_url || `/mcp/${m.id}`, 'MCP URL')}
                            title="Copy full URL"
                          >
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                              <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                              <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
                            </svg>
                          </button>
                        </div>
                      </td>
                      <td>
                        <code style={{ fontSize: '12px', color: 'var(--text-tertiary)', letterSpacing: '0.02em' }}>
                          {m.api_key}
                        </code>
                      </td>
                      <td>
                        <span className={`status-badge ${m.status === 'active' ? 'status-active' : ''}`}>
                          {m.status}
                        </span>
                      </td>
                      <td style={{ color: 'var(--text-secondary)', fontSize: '13px' }}>
                        {new Date(m.created_at).toLocaleDateString('en-US', {
                          year: 'numeric', month: 'short', day: 'numeric',
                        })}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
