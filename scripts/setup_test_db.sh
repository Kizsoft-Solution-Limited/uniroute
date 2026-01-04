#!/bin/bash

# Setup script for Phase 2 integration tests
# This script helps set up PostgreSQL and Redis for integration testing

set -e

echo "🚀 Setting up test databases for Phase 2 integration tests..."

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# PostgreSQL Setup
echo ""
echo "📊 Setting up PostgreSQL test database..."

# Check if PostgreSQL is running
if ! pg_isready > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  PostgreSQL doesn't seem to be running.${NC}"
    echo "   Please start PostgreSQL and try again."
    echo "   On macOS: brew services start postgresql"
    echo "   On Linux: sudo systemctl start postgresql"
    exit 1
fi

# Get PostgreSQL connection details
PG_USER="${PGUSER:-postgres}"
PG_PASSWORD="${PGPASSWORD:-}"
PG_HOST="${PGHOST:-localhost}"
PG_PORT="${PGPORT:-5432}"

echo "   Using PostgreSQL at ${PG_HOST}:${PG_PORT}"

# Create test database
if [ -z "$PG_PASSWORD" ]; then
    # Try without password (trust authentication)
    PGPASSWORD="" psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d postgres -c "CREATE DATABASE uniroute_test;" 2>/dev/null || echo "   Database might already exist (this is OK)"
else
    PGPASSWORD="$PG_PASSWORD" psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d postgres -c "CREATE DATABASE uniroute_test;" 2>/dev/null || echo "   Database might already exist (this is OK)"
fi

echo -e "${GREEN}✅ PostgreSQL test database ready${NC}"
echo "   Database: uniroute_test"
echo "   Connection: postgres://${PG_USER}:${PG_PASSWORD}@${PG_HOST}:${PG_PORT}/uniroute_test?sslmode=disable"

# Redis Setup
echo ""
echo "🔴 Setting up Redis for testing..."

# Check if Redis is running
if ! redis-cli ping > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Redis doesn't seem to be running.${NC}"
    echo "   Please start Redis and try again."
    echo "   On macOS: brew services start redis"
    echo "   On Linux: sudo systemctl start redis"
    echo "   Or use Docker: docker run -d -p 6379:6379 redis:latest"
    exit 1
fi

# Test Redis connection
if redis-cli ping | grep -q PONG; then
    echo -e "${GREEN}✅ Redis is running${NC}"
    echo "   Host: localhost:6379"
    echo "   Test DB: 15"
else
    echo -e "${RED}❌ Redis connection failed${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ All test databases are ready!${NC}"
echo ""
echo "You can now run integration tests:"
echo "  go test ./internal/security -v -run Integration"
echo "  go test ./internal/api/middleware -v -run Integration"
echo ""

