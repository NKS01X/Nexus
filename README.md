# Nexus

**AI agent commerce infrastructure for merchants.**

Nexus is a self-hosted platform that makes any merchant transactable by AI shopping agents — exposing a standards-compliant MCP storefront, enforcing deterministic purchase policies, and maintaining a tamper-evident audit trail of every agent action.

[![Go](https://img.shields.io/badge/go-1.22+-blue)](https://go.dev/) [![License](https://img.shields.io/badge/license-MIT-green)](LICENSE) [![Build](https://img.shields.io/badge/build-passing-brightgreen)]()

---

## The Problem

AI shopping agents are becoming a real sales channel. But merchants have no safe way to accept them today:

- An agent manipulated via prompt injection can request 500 units of a ₹5,000 product.
- There is no standard interface for agents to browse catalogs and place orders.
- There is no audit trail when something goes wrong.

---

## What Nexus Does

Nexus ships two services that work together:

| Service | Purpose |
|---|---|
| **Merchant MCP Server** | Exposes your catalog and checkout as MCP tools any AI agent can call |
| **Policy Gateway** | Enforces hard caps in deterministic Go code before any payment is attempted |

Every purchase flows: `AI Agent → Merchant MCP → Policy Gateway → Razorpay → Audit Log`

Blocked purchases are routed to a human approval queue. Every decision — allowed or blocked — is appended to a SHA-256 hash-chained log.

---

## Key Features

- **MCP-native storefront** — `search_products`, `get_product`, `check_availability`, `purchase`, `get_order_status` exposed as standard MCP tools
- **Deterministic policy engine** — spend cap, per-SKU cap, velocity cap, category allowlist, geo rules, SKU blocklist; all compiled Go, no LLM in the enforcement path
- **Idempotent purchases** — replay-safe via idempotency keys, prevents duplicate charges
- **Hash-chained audit log** — append-only, SHA-256 linked, verifiable on demand
- **Human approval queue** — blocked requests surface in a dashboard for manual review
- **Red team suite** — 7 automated attack vectors run on every deploy to confirm nothing regressed
- **Razorpay integration** — connects to the official `razorpay-mcp-server` binary in test mode

---

## Architecture

```mermaid
graph TD
    subgraph Buyer["AI Buyer"]
        A[AI Shopping Agent]
    end

    subgraph Nexus["Nexus — Merchant Infrastructure"]
        B[Merchant MCP Server]
        C[Policy Gateway]
        D[Approval Queue]
        E[Audit Log]
        F[Razorpay MCP Client]
    end

    subgraph Razorpay["Razorpay"]
        G[Razorpay MCP Server]
        H[Razorpay Test API]
    end

    subgraph Storage["Storage"]
        I[(PostgreSQL)]
    end

    A -->|MCP: search / purchase| B
    B -->|forward purchase| C
    C -->|allowed| F
    C -->|blocked| D
    C -->|every decision| E
    F -->|MCP| G
    G -->|HTTPS| H
    B --- I
    C --- I
    D --- I
    E --- I
```

### Request Flow

| Step | Actor | Action |
|---|---|---|
| 1 | AI Agent | Calls `purchase` on Merchant MCP Server |
| 2 | Merchant MCP | Looks up product price from catalog, forwards to Policy Gateway |
| 3 | Policy Gateway | Evaluates all caps against live DB state |
| 4a | Gateway (allowed) | Creates Razorpay order, captures payment, confirms inventory |
| 4b | Gateway (blocked) | Enqueues to Approval Queue, returns `PENDING_APPROVAL` |
| 5 | Audit Log | Appends SHA-256 chained entry for every outcome |
| 6 | Human Reviewer | Approves or rejects from dashboard; approved triggers payment |

### Policy Engine Checks

| Rule | Default |
|---|---|
| Spend cap | ₹3,000 per buyer per session |
| Per-SKU cap | Configurable per variant |
| Velocity cap | 10 requests / 60 seconds |
| Category allowlist | `footwear`, `apparel` |
| SKU blocklist | Configurable |
| Geo restriction | Pincode-level |


---

## Tech Stack

| Layer | Choice |
|---|---|
| Language | Go 1.22 |
| Database | PostgreSQL 15 |
| Protocol | Model Context Protocol (MCP) |
| Payments | Razorpay MCP Server |
| Deployment | Docker + Docker Compose |

---

## Quick Start

**Prerequisites:** Go 1.22+, PostgreSQL 15+, Docker

```bash
git clone https://github.com/razorpay/nexus.git
cd nexus

cp config.yaml.example config.yaml
# Edit config.yaml — set your DATABASE_DSN and RAZORPAY_KEY_ID / RAZORPAY_KEY_SECRET

docker compose up -d

go run cmd/migrate/main.go config.yaml
go run cmd/seed-catalog/main.go config.yaml

go run cmd/aegis-gateway/main.go config.yaml &
go run cmd/merchant-mcp/main.go config.yaml &
```

The Merchant MCP Server is now live on `:8082`. Point any MCP-compatible AI agent at it.

---

## Configuration

Copy `config.yaml.example` to `config.yaml`. Key fields:

```yaml
database:
  dsn: "postgres://user:pass@localhost:5432/nexus?sslmode=disable"

razorpay:
  key_id: "${RAZORPAY_KEY_ID}"
  key_secret: "${RAZORPAY_KEY_SECRET}"

policy:
  spend_cap_paisa: 300000
  velocity_cap:
    max_requests: 10
    window_seconds: 60
  allowed_categories: ["footwear", "apparel"]
```

Full reference in `config.yaml.example`.

---

## Running the Demo

```bash
go run cmd/demo-buyer/main.go config.yaml
```

This runs a scripted AI buyer that searches the catalog, selects a product, checks availability, and attempts a purchase. You will see the policy decision, audit entry, and order ID printed to stdout.

---

## Red Team Suite

```bash
go run cmd/redteam/main.go config.yaml
```

Runs 7 automated attacks against a live stack:

| Attack | Expected outcome |
|---|---|
| Prompt injection — quantity | Blocked: per-SKU cap |
| Prompt injection — price | Blocked: catalog-price enforced |
| Velocity abuse | Blocked: velocity cap |
| Category escape | Blocked: category not allowed |
| Geo bypass | Blocked: pincode restricted |
| Idempotency replay | Deduplicated: same order returned |
| Hash chain integrity | Verified: chain unbroken |

---

## Approval Dashboard

```bash
go run cmd/dashboard/main.go config.yaml
```

Opens at `http://localhost:8083`. Lists all pending approval queue items with full policy context. Approve or reject with one click.

---

## Project Layout

```
cmd/
  aegis-gateway/    policy gateway entrypoint
  merchant-mcp/     AI-facing MCP server entrypoint
  dashboard/        approval review UI
  demo-buyer/       scripted AI buyer
  redteam/          automated attack suite
  seed-catalog/     synthetic catalog seeder
  migrate/          database migration runner

internal/app/
  service/          business logic (policy engine, gateway, audit, queue)
  repository/       PostgreSQL implementations
  model/            domain types
  mcp/              MCP tool definitions
  integration/      end-to-end test suite

internal/pkg/
  mcp/              MCP server implementations
  razorpay_mcp/     Razorpay MCP client
  config/           config loading
  logger/           structured logger

migrations/         SQL migration files
web/templates/      approval dashboard HTML
```

---

## Contributing

Open a PR. Keep commits conventional (`feat:`, `fix:`, `chore:`).

---

## License

[MIT](LICENSE)
