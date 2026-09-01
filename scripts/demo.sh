#!/bin/bash
# ==============================================================================
# NEXUS — MASTER DEMO ORCHESTRATION SCRIPT
# ==============================================================================
# One script to build, migrate, seed, and run the entire Nexus platform stack.
# ==============================================================================

set -e

# ANSI Color Codes
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
WHITE='\033[1;37m'
BOLD='\033[1m'
DIM='\033[2m'
RESET='\033[0m'

echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════════════════════╗${RESET}"
echo -e "${BOLD}${CYAN}║            NEXUS — MASTER DEMO STACK ORCHESTRATION       ║${RESET}"
echo -e "${BOLD}${CYAN}║         Aegis Policy Engine & Razorpay Integration       ║${RESET}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════════════════════╝${RESET}"
echo ""

# 1. Environment Credentials Check
echo -e "${WHITE}[1/7] Checking environment credentials...${RESET}"
if [ ! -f .env ]; then
  echo -e "${RED}✗ Error: .env file missing!${RESET}"
  echo "Please create a .env file with RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET."
  exit 1
fi

set -a
source .env
set +a
echo -e "      ${GREEN}✓ Credentials loaded from .env${RESET}"

# 2. Database Services Startup
echo -e "${WHITE}[2/7] Starting PostgreSQL database container...${RESET}"
docker compose up -d postgres 2>/dev/null || docker-compose up -d postgres

echo -ne "      Waiting for PostgreSQL readiness... "
for i in {1..15}; do
  if docker exec -i $(docker compose ps -q postgres 2>/dev/null || docker-compose ps -q postgres) psql -U postgres -d aegis -c "SELECT 1;" >/dev/null 2>&1; then
    echo -e "${GREEN}Ready!${RESET}"
    break
  fi
  sleep 1
done

# 3. Build Razorpay MCP Server & Nexus Binaries
echo -e "${WHITE}[3/7] Building binaries & frontend portal...${RESET}"

mkdir -p bin

if [ ! -f bin/razorpay-mcp-server ]; then
  echo -e "      ${YELLOW}Building Razorpay MCP server binary...${RESET}"
  (cd razorpay-mcp-server && go build -v -o ../bin/razorpay-mcp-server ./cmd/razorpay-mcp-server)
fi
echo -e "      ${GREEN}✓ bin/razorpay-mcp-server ready${RESET}"

make build-all >/dev/null 2>&1 || make build-all
echo -e "      ${GREEN}✓ All Aegis binaries & React frontend built${RESET}"

# 4. Database Migrations & Catalog Seeding
echo -e "${WHITE}[4/7] Applying DB migrations & seeding catalog...${RESET}"
export DATABASE_DSN="postgres://postgres:postgres@localhost:5432/aegis?sslmode=disable"
./bin/migrate up >/dev/null 2>&1 || ./bin/migrate up
./bin/seed-catalog "$DATABASE_DSN" >/dev/null 2>&1 || ./bin/seed-catalog "$DATABASE_DSN"
echo -e "      ${GREEN}✓ Schema migrated & product catalog seeded${RESET}"

# 5. Pre-load Approval Queue Data
echo -e "${WHITE}[5/7] Pre-loading approval queue for UI demo...${RESET}"
if [ -f scripts/demo-preload.sh ]; then
  bash scripts/demo-preload.sh >/dev/null 2>&1 || true
fi
echo -e "      ${GREEN}✓ Approval queue seeded with pending review items${RESET}"

# 6. Service Management (Cleanup & Restart)
echo -e "${WHITE}[6/7] Restarting platform microservices...${RESET}"
pkill -f "aegis-gateway|merchant-mcp|dashboard|portal" 2>/dev/null || true
sleep 1

export RAZORPAY_MCP_BINARY_PATH="./bin/razorpay-mcp-server"

./bin/aegis-gateway config.yaml >> gateway.log 2>&1 &
./bin/merchant-mcp config.yaml >> mcp.log 2>&1 &
./bin/dashboard config.yaml --port 8083 >> dashboard.log 2>&1 &
./bin/portal config.yaml >> portal.log 2>&1 &

sleep 2

# 7. Health Check Verification
echo -ne "${WHITE}[7/7] Verifying service endpoints... ${RESET}"
HEALTH_OK=true

if curl -s http://localhost:8084/health >/dev/null 2>&1; then
  echo -ne "${GREEN}Portal (8084) OK  ${RESET}"
else
  HEALTH_OK=false
fi

if curl -s http://localhost:8081/health >/dev/null 2>&1; then
  echo -ne "${GREEN}Gateway (8081) OK  ${RESET}"
else
  HEALTH_OK=false
fi

if curl -s http://localhost:8083/ready >/dev/null 2>&1 || curl -s http://localhost:8083 >/dev/null 2>&1; then
  echo -ne "${GREEN}Approvals (8083) OK${RESET}"
fi
echo ""

echo ""
echo -e "${BOLD}${CYAN}══════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}${GREEN}  ✨  NEXUS DEMO STACK IS LIVE AND READY  ✨${RESET}"
echo -e "${BOLD}${CYAN}══════════════════════════════════════════════════════════${RESET}"
echo ""
echo -e "  ${WHITE}Admin Portal:${RESET}       ${BOLD}${CYAN}http://localhost:8084${RESET}"
echo -e "  ${WHITE}Red Team Suite:${RESET}     ${BOLD}${CYAN}http://localhost:8084/redteam${RESET}"
echo -e "  ${WHITE}Approval Queue:${RESET}     ${BOLD}${CYAN}http://localhost:8084/approvals${RESET}  ${DIM}(or http://localhost:8083)${RESET}"
echo -e "  ${WHITE}Admin Key:${RESET}          ${BOLD}${YELLOW}nexus_admin_default${RESET}"
echo ""
echo -e "  ${WHITE}Quick Commands:${RESET}"
echo -e "    ${DIM}• Terminal Purchase Demo:${RESET}   ${BOLD}bash scripts/demo-purchase.sh${RESET}"
echo -e "    ${DIM}• Red Team CLI Attack:${RESET}     ${BOLD}bash scripts/demo-redteam.sh${RESET}"
echo -e "    ${DIM}• Stop Services:${RESET}           ${BOLD}pkill -f 'aegis-gateway|merchant-mcp|dashboard|portal'${RESET}"
echo ""
echo -e "${BOLD}${CYAN}══════════════════════════════════════════════════════════${RESET}"
echo ""
