#!/bin/bash
# demo-purchase.sh — Cinematic blocked purchase demo for Scene 1 of the hackathon video
# Shows an AI agent discovering a product and getting blocked by Aegis policy

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
DIM='\033[2m'
BOLD='\033[1m'
RESET='\033[0m'
MAGENTA='\033[0;35m'

typeout() {
  local text="$1"
  local delay="${2:-0.03}"
  echo -ne "$text" | while IFS= read -r -n1 char; do
    echo -n "$char"
    sleep "$delay"
  done
  echo ""
}

clear 2>/dev/null || true

echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════════════════════╗${RESET}"
echo -e "${BOLD}${CYAN}║            NEXUS — AI BUYER AGENT v1.0                   ║${RESET}"
echo -e "${BOLD}${CYAN}║        Aegis Policy Engine — LIVE ENFORCEMENT            ║${RESET}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════════════════════╝${RESET}"
echo ""
sleep 0.5

echo -e "${DIM}  Agent initialized. Loading policy rules from Aegis Gateway...${RESET}"
sleep 0.5
echo -e "${DIM}  Spend cap:       ₹3,000 / session${RESET}"
echo -e "${DIM}  Categories:      footwear, apparel${RESET}"
echo -e "${DIM}  Geo:             IN — pincodes 560001, 560002${RESET}"
echo -e "${DIM}  Velocity:        10 req / 60s${RESET}"
echo ""
sleep 0.8

echo -e "${WHITE}  GOAL${RESET}   ${BOLD}Buy high-value electronics for demo_attacker_1${RESET}"
echo -e "${WHITE}  MAX    ₹3,00,000${RESET}"
echo -e "${WHITE}  CAT    electronics${RESET}"
echo ""
sleep 0.5

echo -ne "${YELLOW}  Searching catalog for 'headphones'...${RESET}"
sleep 1.2
echo ""
echo ""
echo -e "  ${BOLD}Found 1 product:${RESET}"
echo -e "  ${GREEN}▸${RESET}  Sony WH-1000XM5 Wireless Headphones"
echo -e "     ${DIM}SKU: HEADPHONES-WL-001-BLK  ·  ₹29,990  ·  In stock: 5${RESET}"
echo ""
sleep 0.7

echo -ne "${YELLOW}  Checking availability for HEADPHONES-WL-001-BLK...${RESET}"
sleep 0.8
echo ""
echo -e "  ${GREEN}✓${RESET}  Available: 5  ·  Reserved: 0"
echo ""
sleep 0.5

echo -ne "${YELLOW}  Fetching product details...${RESET}"
sleep 0.6
echo ""
echo -e "  ${BOLD}Sony WH-1000XM5 Wireless Headphones${RESET}"
echo -e "  ${DIM}Industry-leading noise canceling, 30-hour battery life${RESET}"
echo -e "  ${DIM}Offers: black ₹29,990  ·  silver ₹29,990${RESET}"
echo ""
sleep 0.7

echo -e "  ${BOLD}${MAGENTA}⟳  Selected:${RESET}  HEADPHONES-WL-001-BLK  (black)  —  ${BOLD}₹29,990${RESET}"
echo ""
sleep 0.5

echo -ne "${YELLOW}  Initiating purchase via Aegis Gateway...${RESET}"
sleep 1.5
echo ""
echo ""

echo -e "${BOLD}${CYAN}  ══════════════════════════════════════${RESET}"
echo -e "${BOLD}${CYAN}         AEGIS POLICY DECISION${RESET}"
echo -e "${BOLD}${CYAN}  ══════════════════════════════════════${RESET}"
echo ""
echo -e "  Allowed:    ${BOLD}${RED}false${RESET}"
echo -e "  Status:     ${BOLD}${RED}BLOCKED${RESET}"
echo -e "  Rule:       ${BOLD}category_blocked${RESET}"
echo -e "  Reason:     ${DIM}Category 'electronics' not in allowed list [footwear, apparel]${RESET}"
echo ""
echo -e "${BOLD}${RED}  ✗  PURCHASE BLOCKED — Zero payment processed${RESET}"
echo ""
echo -e "${DIM}  Audit entry written · Trace ID: trc_$(date +%s)${RESET}"
echo ""
echo -e "${BOLD}${CYAN}  ══════════════════════════════════════${RESET}"
echo ""
