<div align="center">
  <br />
  <h1>🚀 Nexus</h1>
  <p><b>Enterprise-Grade Agentic Commerce Infrastructure for Merchants</b></p>
  <p><i>Razorpay Buildathon 2026 — Track 01: AI Growth & Agentic Commerce</i></p>
  <br />

  [![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
  [![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-4169E1?style=flat&logo=postgresql)](https://www.postgresql.org/)
  [![React](https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react)](https://reactjs.org/)
  [![MCP](https://img.shields.io/badge/Protocol-MCP-purple?style=flat)](https://modelcontextprotocol.io/)
  [![Podman](https://img.shields.io/badge/Container-Podman-892CA0?style=flat&logo=podman)](https://podman.io/)
  [![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
</div>

---

## 🎥 Platform Overview

*(Watch the Nexus platform in action, featuring the 3D control plane, multi-tenant merchant onboarding, the deterministic policy engine, human review queue, and automated Red Team security lab.)*

![Nexus Platform Demo](assets/demo.webp)

---

## 💡 The Problem

As AI shopping agents evolve from simple search assistants into autonomous purchasing delegates, merchants face three critical infrastructure vulnerabilities:

1. **Unbounded Financial Risk:** LLMs can be manipulated via prompt injection or context drift to place unauthorized, high-value transactions (e.g., ordering 500 units of a $10,000 item).
2. **Missing Universal Standard:** Traditional e-commerce platforms lack a standardized, machine-readable protocol for AI agents to discover catalogs, check inventory, and execute multi-step checkouts.
3. **Zero Compliance & Auditability:** Non-deterministic LLM actions leave merchants without cryptographically verifiably audit trails for dispute resolution, compliance, or fraud prevention.

---

## ⚡ The Solution: Nexus

**Nexus** is a high-performance, self-hosted commerce gateway that renders any merchant natively transactable by AI agents end-to-end.

By combining an open **Model Context Protocol (MCP)** interface with a zero-LLM **Aegis Policy Engine**, Nexus guarantees that **every monetary action is explainable, bounded, and cryptographically gated**.

```
AI Agent / LLM ──► Merchant MCP Server ──► Aegis Policy Gateway ──► Razorpay API ──► Audit Ledger
                                                   │
                                            (Policy Violation)
                                                   ▼
                                         Human Approval Queue
```

### Key Engineering Guarantees:
- ⚡ **Zero LLM in the Enforcement Path:** Policy evaluation is compiled Go code with sub-millisecond latency. Zero hallucination risk during payment capture.
- 🛡️ **Bounded Execution:** Enforces session spend caps, per-SKU quantity limits, rate limiting (velocity), category allowlists, and regional geo-fencing.
- ⛓️ **Cryptographic Audit Chain:** Hash-chained append-only ledger verifies log integrity against tampering.
- 👨‍⚖️ **Graceful Human-in-the-Loop:** Blocked anomalous transactions do not drop; they seamlessly route to a merchant review queue for manual override.

---

## 🎯 Track 01 Alignment Checklist

| Requirement | Implementation in Nexus | Verification |
|---|---|---|
| **Make merchant transactable by AI buyer** | Exposes standardized MCP endpoints for product search, availability checking, and instant checkout. | `/mcp/{store_id}` tool handlers |
| **Explainable monetary actions** | Every policy decision outputs detailed rationale, rule IDs, and parameter evaluation states. | `policy_decision.rule_fired` in audit log |
| **Bounded execution** | Configurable spend caps, per-SKU quotas, rate limits, allowlists, and geo-fencing. | `internal/app/service/policy_engine.go` |
| **Gated transactions & audit trail** | SHA-256 hash-chained immutable audit log with cryptographic chain verification. | `internal/app/service/audit_service.go` |
| **One failure handled gracefully** | Blocked transactions generate a pending request in the Human Approval Queue instead of failing silently. | `/approvals` dashboard & backend API |

---

## 🛠️ Exposed MCP Tools

The **Merchant MCP Server** acts as the universal protocol layer between external AI agents and the merchant's payment infrastructure:

| Tool Name | Description | Key Parameters |
| :--- | :--- | :--- |
| `search_products` | Search catalog with structured price & category filters | `query`, `category`, `min_price`, `max_price` |
| `get_product` | Fetch detailed SKU specifications and inventory availability | `product_id` |
| `check_availability` | Real-time stock reservation check | `sku` |
| `purchase` | Submit purchase request to Aegis Policy Engine | `buyer_id`, `session_id`, `sku`, `quantity`, `idempotency_key` |
| `get_order_status` | Query status of pending, approved, or completed order | `order_id` |

---

## 🏗️ Architecture & Core Components

![Nexus System Architecture](assets/architecture.png)

Nexus operates as a cohesive suite of microservices:

1. **Merchant MCP Server (`:8082`):** Handles external LLM JSON-RPC connections, validates schemas, and translates tool calls.
2. **Aegis Policy Gateway (`:8081`):** Evaluates deterministic business rules against PostgreSQL state before authorizing Razorpay charges.
3. **Human Approval Service (`:8083`):** Receives policy-blocked transactions, holds them in state, and exposes resolution endpoints.
4. **Admin Portal & Portal Server (`:8084`):** React 18 dashboard offering merchant onboarding, live policy management, review queue, and security test lab.
5. **Razorpay Integration Engine:** Executes payment intent creation and capture using official Razorpay APIs.

---

## 🚨 Automated Security Test Lab (Red Team Suite)

Nexus includes an automated Red Team simulation suite to validate policy enforcement against malicious or prompt-injected AI agents:

```bash
go run cmd/redteam/main.go config.yaml
```

| Attack Vector | Enforcement Layer | Result |
|---|---|---|
| **Excessive Quantity Injection** | Per-SKU Quota Rule | **Blocked** → Sent to Approval Queue |
| **Price Tampering Attack** | Catalog Truth Verification | **Blocked** → Price mismatch rejected |
| **Velocity Spam Attack** | Sliding-window Rate Limiter | **Blocked** → HTTP 429 / Rate Capped |
| **Category Escape** | Category Allowlist Rule | **Blocked** → Unauthorized category denied |
| **Geo-fencing Bypass** | Pincode / Country Rule | **Blocked** → Regional restriction applied |
| **Replay Attack** | Idempotency Key Store | **Deduplicated** → Cached state returned |
| **Audit Chain Tampering** | Cryptographic Hash Chain | **Detected** → Chain validation alert fired |

---

## 🚀 Quick Start Guide

### Prerequisites
- **Go 1.22+**
- **PostgreSQL 15+** (or Podman / Docker)
- **Node.js 20+**

---

### Option A: Single-Pod Deployment via Podman (Recommended)

Run the entire Nexus platform (PostgreSQL + All Services + Web Portal) in a single unified Podman pod:

```bash
# 1. Clone repository
git clone https://github.com/razorpay/nexus.git
cd nexus

# 2. Copy environment secrets
cp .env.example .env
# Edit .env to add RAZORPAY_KEY_ID, RAZORPAY_KEY_SECRET, and GROQ_API_KEY

# 3. Launch Podman Pod
make podman-up
```

*To stop the pod:*
```bash
make podman-down
```

---

### Option B: Local Native Development

```bash
# 1. Configure environment
cp config.yaml.example config.yaml

# 2. Run automated demo harness (starts DB, compiles binaries, seeds catalog)
./scripts/demo.sh
```

---

## 🌐 Platform Portals & Endpoints

Once running, access the platform endpoints:

- 📊 **Merchant Dashboard:** [http://localhost:8084](http://localhost:8084) *(Admin Key: `nexus_admin_default`)*
- 🏬 **Store Management:** [http://localhost:8084/merchants](http://localhost:8084/merchants)
- 👨‍⚖️ **Human Approval Queue:** [http://localhost:8084/approvals](http://localhost:8084/approvals)
- 🧪 **Security Test Lab:** [http://localhost:8084/redteam](http://localhost:8084/redteam)
- 🤖 **AI Checkout Simulator:** [http://localhost:8084/ai-purchase](http://localhost:8084/ai-purchase)
- 🎨 **3D Aegis Control Plane:** [http://localhost:8084/aegis-demo](http://localhost:8084/aegis-demo)
- 🔌 **MCP Merchant Endpoint:** `http://localhost:8082/mcp/{store_id}`

---

## 📁 Repository Structure

```
cmd/
  ├── aegis-gateway/     # Policy engine gateway entrypoint
  ├── merchant-mcp/      # Model Context Protocol server
  ├── portal/            # Portal backend & static web server
  ├── redteam/           # Security attack simulation suite
  ├── demo-buyer/        # AI purchase scenario harness
  └── seed-catalog/      # Product catalog seeder
internal/
  ├── app/
  │   ├── service/       # Business logic (Policy, Audit, Approval, MCP)
  │   ├── repository/    # PostgreSQL data access layer
  │   └── model/         # Domain entities & strict schemas
  └── pkg/               # Shared loggers, config parsers, clients
migrations/              # Versioned SQL database migrations
web/portal/              # React 18 + Three.js + Vite merchant portal
scripts/                 # Podman, deployment, and testing automation
```

---

