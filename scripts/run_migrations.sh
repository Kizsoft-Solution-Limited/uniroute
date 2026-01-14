#!/bin/bash

# Migration runner script for UniRoute
# Checks which migrations have been applied and runs missing ones

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔄 UniRoute Migration Runner${NC}"
echo ""

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo -e "${RED}❌ Error: DATABASE_URL not set${NC}"
    echo "   Please set DATABASE_URL in your .env file or environment"
    exit 1
fi

echo -e "${BLUE}📊 Checking database connection...${NC}"
if ! psql "$DATABASE_URL" -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}❌ Error: Cannot connect to database${NC}"
    echo "   Please check your DATABASE_URL and ensure PostgreSQL is running"
    exit 1
fi
echo -e "${GREEN}✅ Database connection successful${NC}"
echo ""

# Function to check if a migration has been applied
check_migration() {
    local migration_name=$1
    local check_query=$2
    
    if psql "$DATABASE_URL" -tAc "$check_query" 2>/dev/null | grep -q "1"; then
        return 0  # Migration applied
    else
        return 1  # Migration not applied
    fi
}

# Function to run a migration
run_migration() {
    local migration_file=$1
    local migration_name=$(basename "$migration_file")
    
    echo -e "${YELLOW}⏳ Running: $migration_name${NC}"
    if psql "$DATABASE_URL" -f "$migration_file" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Applied: $migration_name${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed: $migration_name${NC}"
        return 1
    fi
}

# Track which migrations need to be run
MIGRATIONS_TO_RUN=()

# Check each migration
echo -e "${BLUE}🔍 Checking migration status...${NC}"

# Migration 001: Initial schema (users, api_keys)
if check_migration "001" "SELECT 1 FROM information_schema.tables WHERE table_name = 'users' AND table_schema = 'public';"; then
    echo -e "${GREEN}✅ 001_initial_schema.sql (already applied)${NC}"
else
    echo -e "${YELLOW}⏸️  001_initial_schema.sql (pending)${NC}"
    MIGRATIONS_TO_RUN+=("migrations/001_initial_schema.sql")
fi

# Migration 002: Analytics schema (requests table)
if check_migration "002" "SELECT 1 FROM information_schema.tables WHERE table_name = 'requests' AND table_schema = 'public';"; then
    echo -e "${GREEN}✅ 002_analytics_schema.sql (already applied)${NC}"
else
    echo -e "${YELLOW}⏸️  002_analytics_schema.sql (pending)${NC}"
    MIGRATIONS_TO_RUN+=("migrations/002_analytics_schema.sql")
fi

# Migration 003: Tunnel schema (tunnels table)
if check_migration "003" "SELECT 1 FROM information_schema.tables WHERE table_name = 'tunnels' AND table_schema = 'public';"; then
    echo -e "${GREEN}✅ 003_tunnel_schema.sql (already applied)${NC}"
else
    echo -e "${YELLOW}⏸️  003_tunnel_schema.sql (pending)${NC}"
    MIGRATIONS_TO_RUN+=("migrations/003_tunnel_schema.sql")
fi

# Migration 004: User provider keys
if check_migration "004" "SELECT 1 FROM information_schema.tables WHERE table_name = 'user_provider_keys' AND table_schema = 'public';"; then
    echo -e "${GREEN}✅ 004_user_provider_keys.sql (already applied)${NC}"
else
    echo -e "${YELLOW}⏸️  004_user_provider_keys.sql (pending)${NC}"
    MIGRATIONS_TO_RUN+=("migrations/004_user_provider_keys.sql")
fi

# Migration 005: Webhook testing (check for query_string column in tunnel_requests)
if check_migration "005" "SELECT 1 FROM information_schema.columns WHERE table_name = 'tunnel_requests' AND column_name = 'query_string' AND table_schema = 'public';"; then
    echo -e "${GREEN}✅ 005_webhook_testing_schema.sql (already applied)${NC}"
else
    echo -e "${YELLOW}⏸️  005_webhook_testing_schema.sql (pending)${NC}"
    MIGRATIONS_TO_RUN+=("migrations/005_webhook_testing_schema.sql")
fi

# Migration 006: User name and password reset
if check_migration "006" "SELECT 1 FROM information_schema.tables WHERE table_name = 'password_reset_tokens' AND table_schema = 'public';"; then
    echo -e "${GREEN}✅ 006_add_user_name_and_password_reset.sql (already applied)${NC}"
else
    echo -e "${YELLOW}⏸️  006_add_user_name_and_password_reset.sql (pending)${NC}"
    MIGRATIONS_TO_RUN+=("migrations/006_add_user_name_and_password_reset.sql")
fi

# Migration 007: Email verification
if check_migration "007" "SELECT 1 FROM information_schema.tables WHERE table_name = 'email_verification_tokens' AND table_schema = 'public';"; then
    echo -e "${GREEN}✅ 007_email_verification.sql (already applied)${NC}"
else
    echo -e "${YELLOW}⏸️  007_email_verification.sql (pending)${NC}"
    MIGRATIONS_TO_RUN+=("migrations/007_email_verification.sql")
fi

# Migration 008: Error logging
if check_migration "008" "SELECT 1 FROM information_schema.tables WHERE table_name = 'error_logs' AND table_schema = 'public';"; then
    echo -e "${GREEN}✅ 008_error_logging.sql (already applied)${NC}"
else
    echo -e "${YELLOW}⏸️  008_error_logging.sql (pending)${NC}"
    MIGRATIONS_TO_RUN+=("migrations/008_error_logging.sql")
fi

# Migration 009: Add user role (check if role column exists OR roles column exists)
# If roles column exists, 010 has been run, so 009 is not needed
if check_migration "009" "SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND (column_name = 'role' OR column_name = 'roles') AND table_schema = 'public';"; then
    if check_migration "010" "SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'roles' AND table_schema = 'public';"; then
        echo -e "${GREEN}✅ 009_add_user_role.sql (skipped - 010 already applied)${NC}"
    else
        echo -e "${GREEN}✅ 009_add_user_role.sql (already applied)${NC}"
    fi
else
    echo -e "${YELLOW}⏸️  009_add_user_role.sql (pending)${NC}"
    MIGRATIONS_TO_RUN+=("migrations/009_add_user_role.sql")
fi

# Migration 010: Change role to roles array
if check_migration "010" "SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'roles' AND data_type = 'ARRAY' AND table_schema = 'public';"; then
    echo -e "${GREEN}✅ 010_change_role_to_roles_array.sql (already applied)${NC}"
else
    echo -e "${YELLOW}⏸️  010_change_role_to_roles_array.sql (pending)${NC}"
    MIGRATIONS_TO_RUN+=("migrations/010_change_role_to_roles_array.sql")
fi

# Migration 011: Routing strategy persistence (system_settings table and users.routing_strategy column)
if check_migration "011" "SELECT 1 FROM information_schema.tables WHERE table_name = 'system_settings' AND table_schema = 'public' AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'routing_strategy' AND table_schema = 'public');"; then
    echo -e "${GREEN}✅ 011_routing_strategy_persistence.sql (already applied)${NC}"
else
    echo -e "${YELLOW}⏸️  011_routing_strategy_persistence.sql (pending)${NC}"
    MIGRATIONS_TO_RUN+=("migrations/011_routing_strategy_persistence.sql")
fi

echo ""

# Run pending migrations
if [ ${#MIGRATIONS_TO_RUN[@]} -eq 0 ]; then
    echo -e "${GREEN}✅ All migrations are up to date!${NC}"
    exit 0
fi

echo -e "${BLUE}🚀 Running ${#MIGRATIONS_TO_RUN[@]} pending migration(s)...${NC}"
echo ""

FAILED=0
for migration in "${MIGRATIONS_TO_RUN[@]}"; do
    if [ ! -f "$migration" ]; then
        echo -e "${RED}❌ Error: Migration file not found: $migration${NC}"
        FAILED=1
        continue
    fi
    
    if ! run_migration "$migration"; then
        FAILED=1
        echo -e "${RED}❌ Migration failed. Stopping.${NC}"
        break
    fi
    echo ""
done

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All migrations completed successfully!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some migrations failed. Please check the errors above.${NC}"
    exit 1
fi

