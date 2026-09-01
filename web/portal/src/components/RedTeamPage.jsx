import React, { useState } from 'react'
import { useAuth, useToast } from '../App'
import Navbar from './Navbar'

export default function RedTeamPage() {
  const [running, setRunning] = useState(false)
  const [results, setResults] = useState(null)
  const { authFetch, isAuthenticated } = useAuth()
  const { showToast } = useToast()

  const runAttacks = async () => {
    setRunning(true)
    setResults(null)
    try {
      const res = await authFetch('/api/redteam/run', { method: 'POST' })
      const data = await res.json()
      
      if (!res.ok) {
        // If it returns JSON with attacks we still display it even if HTTP status isn't 200
        if (data.attacks) {
          setResults(data)
          showToast('Attacks completed with vulnerabilities', 'error')
        } else {
          throw new Error(data.error || 'Failed to run attacks')
        }
      } else {
        setResults(data)
        showToast('All attacks blocked successfully!', 'success')
      }
    } catch (err) {
      showToast(err.message, 'error')
    } finally {
      setRunning(false)
    }
  }

  if (!isAuthenticated) {
    return (
      <div className="merchants-page">
        <Navbar />
        <div className="container" style={{ display: 'flex', justifyContent: 'center', paddingTop: '80px' }}>
          <div className="onboarding-card" style={{ maxWidth: '420px', width: '100%', textAlign: 'center' }}>
            <h2 className="onboarding-title">Admin Access Required</h2>
            <p className="onboarding-subtitle">Please login from the Merchants page first.</p>
          </div>
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
              <h1 className="page-title">Red Team Attack Simulator</h1>
              <p className="page-subtitle">
                Test Aegis Policy Engine against adversarial attacks and prompt injections
              </p>
            </div>
            <div>
              <button 
                className="btn btn-primary" 
                onClick={runAttacks} 
                disabled={running}
                style={{ 
                  background: 'linear-gradient(135deg, #ef4444, #b91c1c)', 
                  borderColor: 'rgba(239, 68, 68, 0.4)',
                  boxShadow: '0 4px 16px rgba(239, 68, 68, 0.3)'
                }}
              >
                {running ? (
                  <>
                    <span className="spinner" style={{ width: 16, height: 16, borderTopColor: '#fff' }} />
                    Running Simulation...
                  </>
                ) : (
                  <>
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>
                    Launch Attack Suite
                  </>
                )}
              </button>
            </div>
          </div>
        </div>

        {!results && !running && (
          <div className="card" style={{ padding: '60px 24px', textAlign: 'center' }}>
            <div style={{ opacity: 0.5, marginBottom: '20px', color: 'var(--accent-red)' }}>
              <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5 12 2"></polygon><line x1="12" y1="22" x2="12" y2="15.5"></line><polyline points="22 8.5 12 15.5 2 8.5"></polyline><polyline points="2 15.5 12 8.5 22 15.5"></polyline><line x1="12" y1="2" x2="12" y2="8.5"></line></svg>
            </div>
            <h2 style={{ fontSize: '24px', marginBottom: '12px' }}>System Ready for Simulation</h2>
            <p style={{ color: 'var(--text-secondary)', maxWidth: '500px', margin: '0 auto' }}>
              Launch the automated test suite to simulate adversarial prompt injections, velocity abuse, and idempotency replays against the policy engine in real-time.
            </p>
          </div>
        )}

        {running && (
          <div className="card" style={{ padding: '60px 24px', textAlign: 'center' }}>
            <div className="spinner" style={{ width: 48, height: 48, margin: '0 auto 24px', borderTopColor: 'var(--accent-red)' }}></div>
            <h2 style={{ fontSize: '20px', marginBottom: '8px', color: 'var(--accent-red)' }}>Executing Attack Vectors...</h2>
            <p style={{ color: 'var(--text-secondary)' }}>Fuzzing MCP endpoints and triggering limit boundaries</p>
          </div>
        )}

        {results && !running && (
          <div>
            <div className="platform-tiles" style={{ marginBottom: '24px' }}>
              <div className="platform-tile" style={{ borderLeft: '4px solid #38bdf8' }}>
                <span style={{ fontSize: '24px', color: 'var(--text-primary)', fontWeight: 700 }}>{results.summary?.total || 0}</span>
                <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>Total Attacks</span>
              </div>
              <div className="platform-tile" style={{ borderLeft: '4px solid #22c55e' }}>
                <span style={{ fontSize: '24px', color: 'var(--text-primary)', fontWeight: 700 }}>{results.summary?.blocked || 0}</span>
                <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>Blocked Safely</span>
              </div>
              <div className="platform-tile" style={{ borderLeft: '4px solid #ef4444' }}>
                <span style={{ fontSize: '24px', color: 'var(--text-primary)', fontWeight: 700 }}>{results.summary?.vulnerable || 0}</span>
                <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>Vulnerabilities</span>
              </div>
            </div>

            <div className="card">
              <div style={{ overflowX: 'auto' }}>
                <table className="merchants-table">
                  <thead>
                    <tr>
                      <th style={{ width: '40px' }}>Result</th>
                      <th style={{ width: '250px' }}>Attack Vector</th>
                      <th>Description & Details</th>
                    </tr>
                  </thead>
                  <tbody>
                    {results.attacks?.map((attack, i) => (
                      <tr key={i} style={{ background: attack.passed ? 'rgba(239, 68, 68, 0.05)' : 'transparent' }}>
                        <td style={{ textAlign: 'center' }}>
                          {attack.passed ? (
                            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" color="var(--accent-red)"><circle cx="12" cy="12" r="10"/><path d="M15 9l-6 6M9 9l6 6"/></svg>
                          ) : (
                            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" color="var(--accent-green)"><path d="M22 11.08V12a10 10 0 11-5.93-9.14"/><path d="M22 4L12 14.01l-3-3"/></svg>
                          )}
                        </td>
                        <td style={{ fontWeight: 600, color: attack.passed ? 'var(--accent-red)' : 'var(--text-primary)' }}>
                          {attack.name}
                        </td>
                        <td>
                          <div style={{ marginBottom: '8px' }}>{attack.description}</div>
                          {attack.details && (
                            <code style={{ 
                              display: 'block', 
                              fontSize: '11px', 
                              color: 'var(--text-tertiary)', 
                              background: 'rgba(0,0,0,0.2)', 
                              padding: '8px', 
                              borderRadius: '4px' 
                            }}>
                              {attack.details}
                            </code>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
