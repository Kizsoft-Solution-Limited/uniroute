#!/bin/bash

# UniRoute Environment Variables Setup Script
# This script generates secure secrets and exports environment variables

set -e

echo "🚀 UniRoute Environment Variables Setup"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Generate secure secrets
echo -e "${BLUE}Generating secure secrets...${NC}"
API_KEY_SECRET=$(openssl rand -hex 32)
JWT_SECRET=$(openssl rand -hex 32)

echo -e "${GREEN}✅ Secrets generated${NC}"
echo ""

# Default values
PORT="${PORT:-8084}"
ENV="${ENV:-production}"
OLLAMA_BASE_URL="${OLLAMA_BASE_URL:-http://localhost:11434}"

# Export environment variables
export PORT="$PORT"
export ENV="$ENV"
export OLLAMA_BASE_URL="$OLLAMA_BASE_URL"
export API_KEY_SECRET="$API_KEY_SECRET"
export JWT_SECRET="$JWT_SECRET"

# Optional Phase 2 variables (if set)
if [ -n "$DATABASE_URL" ]; then
    export DATABASE_URL="$DATABASE_URL"
    echo -e "${GREEN}✅ DATABASE_URL set${NC}"
fi

if [ -n "$REDIS_URL" ]; then
    export REDIS_URL="$REDIS_URL"
    echo -e "${GREEN}✅ REDIS_URL set${NC}"
fi

if [ -n "$IP_WHITELIST" ]; then
    export IP_WHITELIST="$IP_WHITELIST"
    echo -e "${GREEN}✅ IP_WHITELIST set${NC}"
fi

echo ""
echo -e "${GREEN}✅ Environment variables exported!${NC}"
echo ""
echo "Current configuration:"
echo "  PORT=$PORT"
echo "  ENV=$ENV"
echo "  OLLAMA_BASE_URL=$OLLAMA_BASE_URL"
echo "  API_KEY_SECRET=*** (generated)"
echo "  JWT_SECRET=*** (generated)"
[ -n "$DATABASE_URL" ] && echo "  DATABASE_URL=$DATABASE_URL"
[ -n "$REDIS_URL" ] && echo "  REDIS_URL=$REDIS_URL"
[ -n "$IP_WHITELIST" ] && echo "  IP_WHITELIST=$IP_WHITELIST"
echo ""
echo -e "${YELLOW}⚠️  Note: These variables are only set in this shell session.${NC}"
echo -e "${YELLOW}   To persist, add them to your shell profile or use a process manager.${NC}"
echo ""
echo "To run UniRoute with these variables:"
echo "  ./bin/uniroute"
echo ""
echo "Or source this script:"
echo "  source scripts/setup_env.sh"

