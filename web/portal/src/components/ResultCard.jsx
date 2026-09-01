import React, { useState } from 'react'
import { Link } from 'react-router-dom'
import { useToast } from '../App'

export default function ResultCard({ result, onReset }) {
  const { showToast } = useToast()
  const [copiedField, setCopiedField] = useState(null)

  const copyToClipboard = (text, label) => {
    navigator.clipboard.writeText(text)
    setCopiedField(label)
    showToast(`${label} copied to clipboard!`)
    setTimeout(() => setCopiedField(null), 2000)
  }

  const CopyButton = ({ text, label }) => (
    <button
      className={`copy-btn ${copiedField === label ? 'copied' : ''}`}
      onClick={() => copyToClipboard(text, label)}
      title={`Copy ${label}`}
    >
      {copiedField === label ? (
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
          <polyline points="20 6 9 17 4 12" />
        </svg>
      ) : (
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
          <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
        </svg>
      )}
    </button>
  )

  return (
    <div className="result-card">
      <div className="result-header">
        <div className="result-check">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <polyline points="20 6 9 17 4 12" />
          </svg>
        </div>
        <div>
          <div className="result-title">MCP Endpoint Ready</div>
          <div className="result-subtitle">Your store is now AI-transactable</div>
        </div>
      </div>

      <div className="result-rows">
        <div className="result-row">
          <span className="result-label">Store ID</span>
          <span className="result-value">{result.id}</span>
        </div>
        <div className="result-row">
          <span className="result-label">Name</span>
          <span className="result-value">{result.name}</span>
        </div>
        <div className="result-row">
          <span className="result-label">Platform</span>
          <span className={`platform-badge ${result.platform}`}>{result.platform}</span>
        </div>
        <div className="result-row">
          <span className="result-label">MCP URL</span>
          <div className="result-value-row">
            <span className="result-value" style={{ maxWidth: '100%' }}>{result.mcp_url}</span>
            <CopyButton text={result.mcp_url} label="MCP URL" />
          </div>
        </div>
        <div className="result-row">
          <span className="result-label">API Key</span>
          <div className="result-value-row">
            <span className="result-value" style={{ maxWidth: '100%' }}>{result.api_key}</span>
            <CopyButton text={result.api_key} label="API Key" />
          </div>
        </div>
      </div>

      <div style={{
        marginTop: '16px', padding: '12px 16px', borderRadius: '12px',
        background: 'rgba(245, 158, 11, 0.08)', border: '1px solid rgba(245, 158, 11, 0.2)',
        display: 'flex', alignItems: 'flex-start', gap: '10px',
        fontSize: '13px', color: 'var(--accent-amber)', lineHeight: '1.5',
      }}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ flexShrink: 0, marginTop: '2px' }}>
          <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
          <line x1="12" y1="9" x2="12" y2="13" />
          <line x1="12" y1="17" x2="12.01" y2="17" />
        </svg>
        <span>Save your API key now. It will only be shown once and cannot be retrieved later.</span>
      </div>

      <div style={{ marginTop: '28px', display: 'flex', gap: '12px', justifyContent: 'center' }}>
        <button className="btn btn-secondary" onClick={onReset}>
          Provision Another
        </button>
        <Link to="/merchants">
          <button className="btn btn-primary">View All Merchants</button>
        </Link>
      </div>
    </div>
  )
}
