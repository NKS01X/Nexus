#!/bin/bash
# ==============================================================================
# NEXUS — PODMAN SINGLE-POD LAUNCHER
# ==============================================================================
# Runs PostgreSQL + Nexus App inside a single Podman Pod ("nexus-pod").
# ==============================================================================

set -e

# ANSI Color Codes
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
WHITE='\033[1;37m'
BOLD='\033[1m'
RESET='\033[0m'

echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════════════════════╗${RESET}"
echo -e "${BOLD}${CYAN}║             NEXUS — PODMAN SINGLE-POD LAUNCHER           ║${RESET}"
echo -e "${BOLD}${CYAN}║     Running Postgres + Nexus App in a Single Podman Pod  ║${RESET}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════════════════════╝${RESET}"
echo ""

# 1. Check env file
if [ -f .env ]; then
  set -a
  source .env
  set +a
  echo -e "${GREEN}✓ Loaded credentials from .env${RESET}"
else
  echo -e "${YELLOW}⚠ No .env file found. Proceeding with default values.${RESET}"
fi

# 2. Cleanup existing Podman Pod if running
echo -e "${WHITE}[1/4] Cleaning up any existing 'nexus-pod'...${RESET}"
podman pod rm -f nexus-pod 2>/dev/null || true

# 3. Create Podman Pod exposing all ports
echo -e "${WHITE}[2/4] Creating Podman Pod 'nexus-pod'...${RESET}"
podman pod create --name nexus-pod \
  -p 5432:5432 \
  -p 8081:8081 \
  -p 8082:8082 \
  -p 8083:8083 \
  -p 8084:8084

# 4. Start PostgreSQL container inside the Pod
echo -e "${WHITE}[3/4] Starting PostgreSQL container inside 'nexus-pod'...${RESET}"
podman run -d --pod nexus-pod --name nexus-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=aegis \
  docker.io/library/postgres:15-alpine

# 5. Build application container image
echo -e "${WHITE}[4/4] Building Nexus application container image...${RESET}"
podman build -t nexus-app:latest -f Dockerfile .

# 6. Start Nexus App container inside the Pod
echo -e "${WHITE}      Launching Nexus App container inside 'nexus-pod'...${RESET}"
podman run -d --pod nexus-pod --name nexus-app \
  -e RAZORPAY_KEY_ID="${RAZORPAY_KEY_ID}" \
  -e RAZORPAY_KEY_SECRET="${RAZORPAY_KEY_SECRET}" \
  -e GROQ_API_KEY="${GROQ_API_KEY}" \
  nexus-app:latest

echo ""
echo -e "${BOLD}${GREEN}  ✨  NEXUS PODMAN POD IS LIVE  ✨${RESET}"
echo -e "${BOLD}${CYAN}══════════════════════════════════════════════════════════${RESET}"
echo -e "  ${WHITE}Pod Name:${RESET}           ${BOLD}${CYAN}nexus-pod${RESET}"
echo -e "  ${WHITE}Admin Portal:${RESET}       ${BOLD}${CYAN}http://localhost:8084${RESET}"
echo -e "  ${WHITE}Gateway API:${RESET}        ${BOLD}${CYAN}http://localhost:8081${RESET}"
echo -e "  ${WHITE}Merchant MCP:${RESET}       ${BOLD}${CYAN}http://localhost:8082${RESET}"
echo -e "  ${WHITE}Dashboard:${RESET}          ${BOLD}${CYAN}http://localhost:8083${RESET}"
echo -e "  ${WHITE}PostgreSQL:${RESET}         ${BOLD}${CYAN}localhost:5432${RESET}"
echo -e "  ${WHITE}Admin Key:${RESET}          ${BOLD}${YELLOW}nexus_admin_default${RESET}"
echo ""
echo -e "  ${WHITE}Useful Podman Commands:${RESET}"
echo -e "    ${BOLD}podman pod logs -f nexus-pod${RESET}     (View logs)"
echo -e "    ${BOLD}podman pod stop nexus-pod${RESET}        (Stop pod)"
echo -e "    ${BOLD}podman pod rm -f nexus-pod${RESET}       (Remove pod)"
echo -e "${BOLD}${CYAN}══════════════════════════════════════════════════════════${RESET}"
echo ""
