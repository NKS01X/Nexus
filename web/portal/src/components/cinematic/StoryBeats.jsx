import React from 'react'
import { motion } from 'framer-motion'

const BEATS = [
  {
    n: '01',
    title: 'A buyer asks',
    copy: 'An AI buyer discovers your products and places an order through a familiar conversational experience.',
  },
  {
    n: '02',
    title: 'Your store responds',
    copy: 'Merchant MCP turns your catalog, availability, orders, and checkout into tools any compatible agent can use.',
  },
  {
    n: '03',
    title: 'Every request is checked',
    copy: 'Aegis applies your spend, quantity, velocity, category, and location rules before payment is attempted.',
  },
  {
    n: '04',
    title: 'Bad intent stops',
    copy: 'Prompt injection, quantity abuse, and unexpected purchase patterns are blocked before they reach your rails.',
  },
  {
    n: '05',
    title: 'You stay in control',
    copy: 'Allowed orders move forward. Exceptions are routed to your approval queue with a complete audit trail.',
  },
]

export default function StoryBeats() {
  return (
    <section className="story-beats" id="story">
      <p className="section-label">How it works</p>
      <h2 className="section-title">A better storefront, with a guardrail built in</h2>
      <p className="section-subtitle">
        Connect once and give AI buyers a reliable way to discover, purchase, and track orders
        while Aegis stops malicious intent before it becomes a payment.
      </p>
      <div className="story-grid">
        {BEATS.map((beat, i) => (
          <motion.article
            key={beat.n}
            className="story-card"
            initial={{ opacity: 0, y: 24 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, amount: 0.4 }}
            transition={{ duration: 0.7, delay: i * 0.08, ease: [0.16, 1, 0.3, 1] }}
          >
            <span className="story-n">{beat.n}</span>
            <h3>{beat.title}</h3>
            <p>{beat.copy}</p>
          </motion.article>
        ))}
      </div>
    </section>
  )
}
