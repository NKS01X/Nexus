#!/bin/sh
# Single-service bootstrap for Render free tier: one web instance hosts the
# Aegis Gateway, Merchant MCP server, and Portal behind one public port.

set -e
cd "$(dirname "$0")/.."

# Git checkouts exclude config.yaml (contains secrets); seed it from the example.
if [ ! -f config.yaml ]; then
  cp config.yaml.example config.yaml
fi

./bin/aegis-gateway -config config.yaml &
./bin/merchant-mcp -config config.yaml &
exec ./bin/portal -config config.yaml
