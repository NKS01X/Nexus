#!/bin/bash
set -e

echo "=== Running Demo Buyer (Triggering Blocked Purchase) ==="
echo "This will simulate an AI buyer attempting to buy an expensive electronics item"
echo "which should be blocked by the Aegis Policy Engine."
echo ""

cat << 'EOF' > scripts/blocked-goal.json
{
  "query": "headphones",
  "max_price_inr": 300000,
  "category": "electronics",
  "buyer_id": "demo_attacker_1",
  "buyer_pincode": "400001"
}
EOF

go run cmd/demo-buyer/main.go config.yaml scripts/blocked-goal.json
