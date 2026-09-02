import React, { useState, useEffect } from 'react'
import { useAuth, useToast, safeJsonParse } from '../App'

export default function AuditTrail() {
  const [entries, setEntries] = useState([])
  const [loading, setLoading] = useState(true)
  const [filterBuyer, setFilterBuyer] = useState('')
  const [expandedRow, setExpandedRow] = useState(null)
  const { authFetch } = useAuth()
  const { showToast } = useToast()

  useEffect(() => {
    fetchTrail()
  }, [])

  const fetchTrail = async () => {
    setLoading(true)
    try {
      const url = filterBuyer ? `/api/audit/trail?buyer_id=${encodeURIComponent(filterBuyer)}` : '/api/audit/entries'
      const res = await authFetch(url)
      const data = await safeJsonParse(res)
      if (!res.ok) throw new Error(data.error || 'Failed to fetch audit trail')
      setEntries(Array.isArray(data) ? data : [])
    } catch (err) {
      showToast(err.message, 'error')
    } finally {
      setLoading(false)
    }
  }

  const handleFilter = (e) => {
    e.preventDefault()
    fetchTrail()
  }

  const copyHash = (hash) => {
    if (!hash) return
    navigator.clipboard.writeText(hash)
    showToast('Hash copied to clipboard')
  }

  const toggleRow = (id) => {
    setExpandedRow(expandedRow === id ? null : id)
  }

  const getActionColor = (action) => {
    switch (action) {
      case 'PURCHASE_ATTEMPT': return 'rgba(56,189,248,0.2)'; // blue
      case 'PAYMENT_EXECUTED': return 'rgba(34,197,94,0.2)'; // green
      case 'PURCHASE_BLOCKED': return 'rgba(239,68,68,0.2)'; // red
      case 'ESCALATED': return 'rgba(245,158,11,0.2)'; // amber
      case 'PURCHASE_APPROVED': return 'rgba(168,85,247,0.2)'; // purple
      case 'PURCHASE_REJECTED': return 'rgba(239,68,68,0.2)'; // red
      default: return 'rgba(255,255,255,0.1)';
    }
  }

  const getActionTextColor = (action) => {
    switch (action) {
      case 'PURCHASE_ATTEMPT': return '#38bdf8';
      case 'PAYMENT_EXECUTED': return '#22c55e';
      case 'PURCHASE_BLOCKED': return '#ef4444';
      case 'ESCALATED': return '#f59e0b';
      case 'PURCHASE_APPROVED': return '#a855f7';
      case 'PURCHASE_REJECTED': return '#ef4444';
      default: return '#e2e8f0';
    }
  }

  return (
    <div className="card">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
        <h2 className="onboarding-title" style={{ fontSize: '18px', margin: 0 }}>Decision Trail</h2>
        <form onSubmit={handleFilter} style={{ display: 'flex', gap: '8px' }}>
          <input 
            type="text" 
            className="form-input" 
            style={{ width: '200px', height: '32px', fontSize: '13px' }}
            placeholder="Filter by buyer ID..." 
            value={filterBuyer}
            onChange={(e) => setFilterBuyer(e.target.value)}
          />
          <button type="submit" className="btn btn-primary" style={{ padding: '0 12px', height: '32px', fontSize: '13px' }}>Filter</button>
          {filterBuyer && (
            <button type="button" className="btn btn-ghost" style={{ padding: '0 12px', height: '32px', fontSize: '13px' }} onClick={() => { setFilterBuyer(''); setTimeout(fetchTrail, 0); }}>Clear</button>
          )}
        </form>
      </div>

      {loading ? (
        <div style={{ padding: '40px', textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading...</div>
      ) : entries.length === 0 ? (
        <div style={{ padding: '40px', textAlign: 'center', color: 'var(--text-tertiary)' }}>No audit entries found.</div>
      ) : (
        <div style={{ overflowX: 'auto' }}>
          <table className="merchants-table" style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr>
                <th style={{ width: '140px' }}>Timestamp</th>
                <th style={{ width: '180px' }}>Action</th>
                <th>Buyer / Session</th>
                <th>Hash Chain</th>
                <th style={{ width: '40px' }}></th>
              </tr>
            </thead>
            <tbody>
              {entries.map((entry) => (
                <React.Fragment key={entry.id}>
                  <tr style={{ cursor: 'pointer', borderBottom: '1px solid rgba(255,255,255,0.05)' }} onClick={() => toggleRow(entry.id)}>
                    <td style={{ color: 'var(--text-secondary)', fontSize: '12px' }}>
                      {new Date(entry.timestamp).toLocaleString()}
                    </td>
                    <td>
                      <span className="audit-action-badge" style={{ 
                        background: getActionColor(entry.action), 
                        color: getActionTextColor(entry.action),
                        padding: '4px 8px', borderRadius: '4px', fontSize: '11px', fontWeight: 600, letterSpacing: '0.05em'
                      }}>
                        {entry.action}
                      </span>
                    </td>
                    <td>
                      <div style={{ fontSize: '13px', fontWeight: 500 }}>{entry.buyer_id}</div>
                      <div style={{ fontSize: '11px', color: 'var(--text-tertiary)' }}>{entry.session_id}</div>
                    </td>
                    <td>
                      <div className="audit-hash" style={{ 
                        display: 'flex', alignItems: 'center', gap: '6px',
                        fontFamily: 'monospace', fontSize: '12px', color: 'var(--text-tertiary)' 
                      }}>
                        <span title={entry.hash}>{entry.hash ? entry.hash.substring(0, 12) + '...' : 'N/A'}</span>
                        {entry.hash && (
                          <svg onClick={(e) => { e.stopPropagation(); copyHash(entry.hash); }} style={{ cursor: 'pointer' }} width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
                        )}
                      </div>
                    </td>
                    <td>
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ transform: expandedRow === entry.id ? 'rotate(180deg)' : 'none', transition: 'transform 0.2s', color: 'var(--text-tertiary)' }}>
                        <path d="M6 9l6 6 6-6"/>
                      </svg>
                    </td>
                  </tr>
                  {expandedRow === entry.id && (
                    <tr style={{ background: 'rgba(0,0,0,0.2)' }}>
                      <td colSpan="5" style={{ padding: '16px' }}>
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                          {entry.buyer_reasoning && (
                            <div>
                              <h4 style={{ margin: '0 0 8px 0', fontSize: '12px', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>AI Buyer Reasoning</h4>
                              <div className="reasoning-text" style={{ 
                                padding: '12px', background: 'rgba(255,255,255,0.03)', borderRadius: '6px', 
                                borderLeft: '3px solid var(--accent-purple)', fontSize: '13px', color: 'var(--text-secondary)' 
                              }}>
                                {entry.buyer_reasoning}
                              </div>
                            </div>
                          )}
                          {entry.policy_decision && (
                            <div>
                              <h4 style={{ margin: '0 0 8px 0', fontSize: '12px', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Policy Decision</h4>
                              <pre style={{ 
                                margin: 0, padding: '12px', background: 'rgba(0,0,0,0.3)', borderRadius: '6px', 
                                fontSize: '12px', color: 'var(--text-tertiary)', overflowX: 'auto' 
                              }}>
                                {JSON.stringify(entry.policy_decision, null, 2)}
                              </pre>
                            </div>
                          )}
                          {(entry.request || entry.response) && (
                            <div style={{ display: 'flex', gap: '16px' }}>
                              {entry.request && (
                                <div style={{ flex: 1, minWidth: 0 }}>
                                  <h4 style={{ margin: '0 0 8px 0', fontSize: '12px', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Request Data</h4>
                                  <pre style={{ margin: 0, padding: '12px', background: 'rgba(0,0,0,0.3)', borderRadius: '6px', fontSize: '11px', color: 'var(--text-tertiary)', overflowX: 'auto', maxHeight: '200px' }}>
                                    {JSON.stringify(entry.request, null, 2)}
                                  </pre>
                                </div>
                              )}
                              {entry.response && (
                                <div style={{ flex: 1, minWidth: 0 }}>
                                  <h4 style={{ margin: '0 0 8px 0', fontSize: '12px', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Response Data</h4>
                                  <pre style={{ margin: 0, padding: '12px', background: 'rgba(0,0,0,0.3)', borderRadius: '6px', fontSize: '11px', color: 'var(--text-tertiary)', overflowX: 'auto', maxHeight: '200px' }}>
                                    {JSON.stringify(entry.response, null, 2)}
                                  </pre>
                                </div>
                              )}
                            </div>
                          )}
                        </div>
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
