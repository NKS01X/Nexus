#!/bin/bash
# demo-preload.sh — Seeds the approval queue with realistic pending items
# Run this once BEFORE starting screen recording for Scene 2 (approval queue)

set -e

DSN="${DATABASE_DSN:-postgres://postgres:postgres@localhost:5432/aegis?sslmode=disable}"
DOCKER_PG=$(podman compose ps -q postgres 2>/dev/null)

run_sql() {
  if [ -n "$DOCKER_PG" ]; then
    podman exec -i "$DOCKER_PG" psql -U postgres -d aegis -q -c "$1"
  else
    psql "$DSN" -q -c "$1"
  fi
}

echo "Seeding approval queue for demo..."

# Clear old demo queue entries first
run_sql "DELETE FROM approval_queue WHERE buyer_id IN ('demo_reviewer_1','demo_reviewer_2','demo_reviewer_3');"

# Seed 3 realistic pending approval items matching schema
run_sql "
INSERT INTO approval_queue (id, buyer_id, session_id, purchase_request, policy_decision, buyer_reasoning, status, created_at, expires_at)
VALUES
  ('demo-queue-001', 'demo_reviewer_1', 'sess_demo_001',
   '{\"sku\":\"WATCH-SMART-001-TIT-44\",\"amount_paisa\":9000000,\"quantity\":1}'::jsonb,
   '{\"allowed\":false,\"rule_fired\":\"spend_cap_exceeded\",\"reason\":\"Purchase amount ₹90,000 exceeds session spend cap of ₹3,000.\"}'::jsonb,
   'This Apple Watch Ultra 2 is essential for triathlon training performance monitoring.',
   'PENDING', NOW() - INTERVAL '4 minutes', NOW() + INTERVAL '24 hours'),

  ('demo-queue-002', 'demo_reviewer_2', 'sess_demo_002',
   '{\"sku\":\"MAT-YOGA-001-PUR\",\"amount_paisa\":1200000,\"quantity\":1}'::jsonb,
   '{\"allowed\":false,\"rule_fired\":\"spend_cap_exceeded\",\"reason\":\"Purchase amount ₹12,000 exceeds session spend cap.\"}'::jsonb,
   'Manduka PRO mat with lifetime guarantee — best long-term value for daily practice.',
   'PENDING', NOW() - INTERVAL '12 minutes', NOW() + INTERVAL '24 hours'),

  ('demo-queue-003', 'demo_reviewer_3', 'sess_demo_003',
   '{\"sku\":\"SNEAKERS-LTD-001-CHI-42\",\"amount_paisa\":15000000,\"quantity\":1}'::jsonb,
   '{\"allowed\":false,\"rule_fired\":\"spend_cap_exceeded\",\"reason\":\"Purchase amount ₹1,50,000 exceeds session spend cap.\"}'::jsonb,
   'Jordan 1 Chicago — limited edition, will appreciate in value. Treat as investment.',
   'PENDING', NOW() - INTERVAL '2 minutes', NOW() + INTERVAL '24 hours')
ON CONFLICT (id) DO NOTHING;
"

echo ""
echo "Approval queue seeded with 3 pending items:"
echo "  - Apple Watch Ultra 2 (₹90,000) — spend cap exceeded"
echo "  - Manduka PRO Yoga Mat (₹12,000) — spend cap exceeded"
echo "  - Jordan 1 Chicago (₹1,50,000) — spend cap exceeded"
echo ""
echo "Open http://localhost:8084/approvals or http://localhost:8083 to view."
