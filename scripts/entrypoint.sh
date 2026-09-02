#!/bin/sh
set -e

echo "[Nexus Pod] Waiting for PostgreSQL database at localhost:5432..."
until pg_isready -h localhost -p 5432 -U postgres >/dev/null 2>&1; do
  sleep 1
done

echo "[Nexus Pod] PostgreSQL is ready!"
echo "[Nexus Pod] Running database migrations..."
/usr/local/bin/migrate up || true

echo "[Nexus Pod] Seeding product catalog..."
export DATABASE_DSN="postgres://postgres:postgres@localhost:5432/aegis?sslmode=disable"
/usr/local/bin/seed-catalog "$DATABASE_DSN" || true

echo "[Nexus Pod] Starting Aegis Gateway on :8081..."
/usr/local/bin/aegis-gateway --config /app/config.yaml > /app/gateway.log 2>&1 &

echo "[Nexus Pod] Starting Merchant MCP on :8082..."
/usr/local/bin/merchant-mcp --config /app/config.yaml > /app/mcp.log 2>&1 &

echo "[Nexus Pod] Starting Approval Dashboard on :8083..."
/usr/local/bin/dashboard --config /app/config.yaml --port 8083 > /app/dashboard.log 2>&1 &

echo "[Nexus Pod] Starting Admin Portal on :8084..."
exec /usr/local/bin/portal --config /app/config.yaml
