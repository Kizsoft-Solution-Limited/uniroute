#!/bin/bash

# UniRoute Production Environment Variables Setup
# This script sets up all environment variables for production deployment

set -e

echo "UniRoute Production Environment Setup"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check for required tools
if ! command -v openssl &> /dev/null; then
    echo -e "${RED}Error: openssl not found. Please install openssl.${NC}"
    exit 1
fi

# Generate secure secrets
echo -e "${BLUE}Generating secure secrets...${NC}"
API_KEY_SECRET=$(openssl rand -hex 32)
JWT_SECRET=$(openssl rand -hex 32)

echo -e "${GREEN}Secrets generated${NC}"
echo ""

# Required variables
export PORT="${PORT:-8084}"
export ENV="${ENV:-production}"
export OLLAMA_BASE_URL="${OLLAMA_BASE_URL:-http://localhost:11434}"
export API_KEY_SECRET="$API_KEY_SECRET"
export JWT_SECRET="$JWT_SECRET"

# Phase 2 variables (required for full features)
if [ -z "$DATABASE_URL" ]; then
    echo -e "${YELLOW}Warning: DATABASE_URL not set. Phase 2 features will be disabled.${NC}"
    echo "   Set it with: export DATABASE_URL='postgres://user:password@host/db?sslmode=require'"
else
    export DATABASE_URL="$DATABASE_URL"
    echo -e "${GREEN}DATABASE_URL set${NC}"
fi

if [ -z "$REDIS_URL" ]; then
    echo -e "${YELLOW}Warning: REDIS_URL not set. Rate limiting will be disabled.${NC}"
    echo "   Set it with: export REDIS_URL='redis://host:6379'"
else
    export REDIS_URL="$REDIS_URL"
    echo -e "${GREEN}REDIS_URL set${NC}"
fi

# Optional variables
if [ -n "$IP_WHITELIST" ]; then
    export IP_WHITELIST="$IP_WHITELIST"
    echo -e "${GREEN}IP_WHITELIST set${NC}"
fi

echo ""
echo -e "${GREEN}Production environment configured${NC}"
echo ""
echo "Configuration Summary:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  PORT:              $PORT"
echo "  ENV:               $ENV"
echo "  OLLAMA_BASE_URL:   $OLLAMA_BASE_URL"
echo "  API_KEY_SECRET:    *** (32-byte hex, generated)"
echo "  JWT_SECRET:        *** (32-byte hex, generated)"
[ -n "$DATABASE_URL" ] && echo "  DATABASE_URL:      $DATABASE_URL" || echo "  DATABASE_URL:      (not set - Phase 1 mode)"
[ -n "$REDIS_URL" ] && echo "  REDIS_URL:         $REDIS_URL" || echo "  REDIS_URL:         (not set - no rate limiting)"
[ -n "$IP_WHITELIST" ] && echo "  IP_WHITELIST:      $IP_WHITELIST" || echo "  IP_WHITELIST:      (not set - allow all)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${YELLOW}IMPORTANT: Save these secrets securely.${NC}"
echo "   They are only shown once and cannot be recovered."
echo ""
echo "To save to a file (optional):"
echo "  echo 'API_KEY_SECRET=$API_KEY_SECRET' >> .env.production"
echo "  echo 'JWT_SECRET=$JWT_SECRET' >> .env.production"
echo ""
echo "To run UniRoute:"
echo "  ./bin/uniroute"
echo ""

