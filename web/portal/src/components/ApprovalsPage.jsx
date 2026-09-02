import React, { useState, useEffect } from 'react'
import { useToast, useAuth, safeJsonParse } from '../App'
import Navbar from './Navbar'
import AuditTrail from './AuditTrail'
import LoginGate from './LoginGate'

export default function ApprovalsPage() {

  const [pending, setPending] = useState([])
  const [loading, setLoading] = useState(true)
  const [verifyResult, setVerifyResult] = useState(null)
  const [verifying, setVerifying] = useState(false)
  const { showToast } = useToast()
  const { isAuthenticated, authFetch } = useAuth()

  useEffect(() => {
    if (isAuthenticated) {
      fetchApprovals()
    }
  }, [isAuthenticated])

  const fetchApprovals = async () => {
    setLoading(true)
    try {
      const res = await authFetch('/api/approvals')
      const data = await safeJsonParse(res)
      if (!res.ok) throw new Error(data.error || 'Failed to fetch approvals')
      setPending(Array.isArray(data) ? data : [])
    } catch (err) {
      showToast(err.message, 'error')
    } finally {
      setLoading(false)
    }
  }

  const handleAction = async (id, action, note) => {
    try {
      const res = await authFetch(`/api/approvals/${action}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id, note })
      })
      const err = await safeJsonParse(res)
      if (!res.ok) {
        throw new Error(err.error || `Failed to ${action}`)
      }
      showToast(`Request ${action}d successfully`)
      fetchApprovals()
    } catch (err) {
      showToast(err.message, 'error')
    }
  }

  const verifyChain = async () => {
    setVerifying(true)
    try {
      const res = await authFetch('/api/audit/verify')
      const data = await safeJsonParse(res)
      if (!res.ok) throw new Error(data.error || 'Failed to verify chain')
      setVerifyResult(data)
      if (data.chain_valid) {
        showToast('Audit chain is valid', 'success')
      } else {
        showToast('Audit chain validation failed!', 'error')
      }
    } catch (err) {
      showToast(err.message, 'error')
    } finally {
      setVerifying(false)
    }
  }

  if (!isAuthenticated) {
    return <LoginGate subtitle="Enter your admin key to view and manage approval requests." />
  }


  return (
    <div className="merchants-page">
      <Navbar />
      <div className="container">
        <div className="page-header">
          <div className="page-header-row">
            <div>
              <h1 className="page-title">Approval Queue</h1>
              <p className="page-subtitle">
                {pending.length} pending request{pending.length !== 1 ? 's' : ''} requiring review
              </p>
            </div>
            <div style={{ display: 'flex', gap: '12px' }}>
              <button className="btn btn-ghost" onClick={fetchApprovals}>
                Refresh
              </button>
            </div>
          </div>
        </div>

        {loading ? (
          <div className="loading"><div className="spinner" /></div>
        ) : pending.length === 0 ? (
          <div className="card" style={{ marginBottom: '24px' }}>
            <div className="empty-state">
              <div className="empty-icon">
                <svg width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h2 className="empty-title">All caught up</h2>
              <p className="empty-desc">No pending approvals require your attention.</p>
            </div>
          </div>
        ) : (
          <div className="card" style={{ marginBottom: '24px' }}>
            <div style={{ overflowX: 'auto' }}>
              <table className="merchants-table">
                <thead>
                  <tr>
                    <th>Buyer ID</th>
                    <th>Amount</th>
                    <th>SKU</th>
                    <th>Rule Fired</th>
                    <th>Reasoning</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {pending.map((p) => (
                    <tr key={p.id}>
                      <td><code style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>{p.buyer_id}</code></td>
                      <td style={{ fontWeight: 600 }}>₹{(p.purchase_request.amount_paisa / 100).toFixed(2)}</td>
                      <td><code style={{ fontSize: '13px', color: 'var(--text-tertiary)' }}>{p.purchase_request.sku}</code></td>
                      <td><span className="platform-badge" style={{ background: 'rgba(239,68,68,0.1)', color: 'var(--accent-red)', borderColor: 'rgba(239,68,68,0.2)' }}>{p.policy_decision.rule_fired}</span></td>
                      <td>
                        {p.buyer_reasoning ? (
                          <div className="reasoning-text" style={{ fontSize: '12px', maxWidth: '200px', whiteSpace: 'normal' }}>
                            {p.buyer_reasoning}
                          </div>
                        ) : (
                          <span style={{ color: 'var(--text-tertiary)' }}>No reasoning</span>
                        )}
                      </td>
                      <td>
                        <div className="approval-actions" style={{ display: 'flex', gap: '8px' }}>
                          <button 
                            className="btn btn-primary" 
                            style={{ padding: '4px 12px', fontSize: '12px', height: 'auto' }}
                            onClick={() => handleAction(p.id, 'approve', 'Approved by admin')}
                          >
                            Approve
                          </button>
                          <button 
                            className="btn btn-ghost" 
                            style={{ padding: '4px 12px', fontSize: '12px', height: 'auto', color: 'var(--accent-red)' }}
                            onClick={() => handleAction(p.id, 'reject', 'Rejected by admin')}
                          >
                            Reject
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        <div className="card" style={{ marginBottom: '24px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
            <h2 className="onboarding-title" style={{ fontSize: '18px', margin: 0 }}>Audit Chain Integrity</h2>
            <button className="btn btn-ghost" onClick={verifyChain} disabled={verifying}>
              {verifying ? 'Verifying...' : 'Verify Chain'}
            </button>
          </div>
          {verifyResult && (
            <div className={`verify-result ${verifyResult.chain_valid ? 'status-active' : ''}`} style={{ 
              padding: '12px 16px', 
              borderRadius: '8px', 
              background: verifyResult.chain_valid ? 'rgba(34,197,94,0.1)' : 'rgba(239,68,68,0.1)',
              border: `1px solid ${verifyResult.chain_valid ? 'rgba(34,197,94,0.2)' : 'rgba(239,68,68,0.2)'}`,
              color: verifyResult.chain_valid ? 'var(--accent-green)' : 'var(--accent-red)',
              display: 'flex',
              alignItems: 'center',
              gap: '8px'
            }}>
              {verifyResult.chain_valid ? (
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M20 6L9 17l-5-5"/></svg>
              ) : (
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="10"/><path d="M15 9l-6 6M9 9l6 6"/></svg>
              )}
              <span style={{ fontWeight: 500 }}>
                {verifyResult.chain_valid ? 'Chain is valid and untampered.' : 'Chain verification failed! Integrity compromised.'}
              </span>
              <span style={{ marginLeft: 'auto', fontSize: '12px', color: 'var(--text-tertiary)' }}>
                Verified at: {new Date(verifyResult.verified_at).toLocaleString()}
              </span>
            </div>
          )}
        </div>

        <AuditTrail />
      </div>
    </div>
  )
}
