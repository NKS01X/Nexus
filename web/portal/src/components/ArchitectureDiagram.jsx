import React, { useState } from 'react'
import { motion } from 'framer-motion'

const NODES = [
  {
    id: 'ai-agent',
    label: 'AI Agent',
    sub: 'MCP Client',
    color: '#a855f7',
    icon: 'spark',
  },
  {
    id: 'merchant-mcp',
    label: 'Merchant MCP',
    sub: 'Tools API',
    color: '#06b6d4',
    icon: 'square',
  },
  {
    id: 'policy-gateway',
    label: 'Policy Gateway',
    sub: 'Aegis Engine',
    color: '#f59e0b',
    icon: 'diamond',
  },
  {
    id: 'razorpay',
    label: 'Razorpay',
    sub: 'Payment Rails',
    color: '#ef4444',
    icon: 'stack',
  },
]

function ArchitectureIcon({ type }) {
  const paths = {
    spark: <path d="M12 3l1.55 5.45L19 10l-5.45 1.55L12 17l-1.55-5.45L5 10l5.45-1.55L12 3z" />,
    square: <><rect x="5" y="5" width="14" height="14" rx="1" /><path d="M8 8h8v8H8z" /></>,
    diamond: <path d="M12 3l9 9-9 9-9-9 9-9z" />,
    stack: <><rect x="5" y="4" width="14" height="16" rx="1" /><path d="M8 8h8M8 12h8M8 16h8" /></>,
    audit: <><path d="M9 7l3-3 3 3-3 3-3-3zM9 17l3-3 3 3-3 3-3-3z" /><path d="M12 10v4M7 12h10" /></>,
  }

  return (
    <svg width="25" height="25" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      {paths[type]}
    </svg>
  )
}

export default function ArchitectureDiagram() {
  const [activeNode, setActiveNode] = useState(null)

  return (
    <section className="arch-section" id="architecture">
      <div className="section-label">Architecture</div>
      <h2 className="section-title">Your commerce stack, ready for agents</h2>
      <p className="section-subtitle">
        Merchant MCP connects AI buyers to your store. Aegis evaluates every request,
        stops malicious or out-of-policy orders, and only forwards approved transactions
        to your payment provider.
      </p>

      <div className="architecture-flow" aria-label="Aegis purchase architecture">
        {NODES.map((node, index) => (
          <React.Fragment key={node.id}>
            <motion.button
              type="button"
              className={`architecture-node ${activeNode === node.id ? 'active' : ''}`}
              style={{ '--node-color': node.color }}
              onClick={() => setActiveNode(activeNode === node.id ? null : node.id)}
            >
              <span className="architecture-node-icon"><ArchitectureIcon type={node.icon} /></span>
              <span className="architecture-node-name">{node.label}</span>
              <span className="architecture-node-sub">{node.sub}</span>
              {activeNode === node.id && (
                <span className="architecture-node-state">verified</span>
              )}
            </motion.button>
            {index < NODES.length - 1 && (
              <span className="architecture-connector" aria-hidden="true">
                <span />
              </span>
            )}
          </React.Fragment>
        ))}

        <motion.button
          type="button"
          className={`architecture-audit ${activeNode === 'audit-log' ? 'active' : ''}`}
          style={{ '--node-color': '#10b981' }}
          onClick={() => setActiveNode(activeNode === 'audit-log' ? null : 'audit-log')}
        >
          <span className="architecture-node-icon"><ArchitectureIcon type="audit" /></span>
          <span className="architecture-node-name">Audit Log</span>
          <span className="architecture-node-sub">Hash-Chained</span>
        </motion.button>
      </div>
    </section>
  )
}
