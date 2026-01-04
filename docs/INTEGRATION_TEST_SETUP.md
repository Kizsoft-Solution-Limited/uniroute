# Integration Test Setup Guide

## Prerequisites

To run Phase 2 integration tests with real data, you need:

1. **PostgreSQL** running and accessible
2. **Redis** running and accessible

## Setup Instructions

### 1. PostgreSQL Setup

```bash
# Create test database
createdb uniroute_test

# Or using psql
psql -U postgres -c "CREATE DATABASE uniroute_test;"
```

**Default connection string:**
```
postgres://postgres:postgres@localhost/uniroute_test?sslmode=disable
```

**To use different credentials:**
- Modify `getTestPostgresClient()` in `internal/security/integration_test.go`
- Or set environment variable: `TEST_DATABASE_URL`

### 2. Redis Setup

```bash
# Start Redis (if not already running)
redis-server

# Or using Docker
docker run -d -p 6379:6379 redis:latest
```

**Default connection:**
- Host: `localhost:6379`
- Database: `15` (for testing)

**To use different Redis:**
- Modify `getTestRedisClient()` in `internal/security/integration_test.go`
- Or set environment variable: `TEST_REDIS_URL`

### 3. Run Integration Tests

```bash
# Run all integration tests
CGO_ENABLED=0 go test ./internal/security -v -run Integration
CGO_ENABLED=0 go test ./internal/api/middleware -v -run Integration

# Run specific test
CGO_ENABLED=0 go test ./internal/security -v -run TestAPIKeyServiceV2_Integration

# Run with coverage
CGO_ENABLED=0 go test ./internal/security -v -run Integration -coverprofile=coverage.out
```

## Test Coverage

### Security Package Integration Tests

1. **TestAPIKeyServiceV2_Integration**
   - Creates API key in real PostgreSQL
   - Validates API key against database
   - Tests invalid key rejection

2. **TestAPIKeyServiceV2_Integration_Expiration**
   - Creates expired API key
   - Verifies expiration is enforced

3. **TestAPIKeyServiceV2_Integration_Inactive**
   - Creates API key
   - Deactivates it
   - Verifies inactive keys are rejected

4. **TestRateLimiter_Integration**
   - Tests rate limiting with real Redis
   - Verifies per-minute limits
   - Checks remaining request tracking

5. **TestRateLimiter_Integration_DayLimit**
   - Tests daily rate limits
   - Verifies day limit enforcement

6. **TestFullFlow_Integration**
   - Complete end-to-end test
   - API key creation → validation → rate limiting
   - Tests all components together

### Middleware Package Integration Tests

1. **TestAuthMiddlewareV2_Integration**
   - Tests API key authentication with real database
   - Verifies context setting
   - Tests invalid key rejection

2. **TestRateLimitMiddleware_Integration**
   - Tests rate limiting middleware with real Redis
   - Verifies HTTP status codes
   - Checks rate limit headers

3. **TestFullMiddlewareFlow_Integration**
   - Complete middleware chain test
   - Security headers → Auth → Rate limiting
   - Tests full request flow

## Test Data Management

- **PostgreSQL**: Tests automatically create tables and truncate data before each test
- **Redis**: Tests use DB 15 and flush it before each test
- **Isolation**: Each test is independent and cleans up after itself

## Troubleshooting

### PostgreSQL Connection Issues

```bash
# Check if PostgreSQL is running
pg_isready

# Check if database exists
psql -U postgres -l | grep uniroute_test

# Create database if missing
createdb uniroute_test
```

### Redis Connection Issues

```bash
# Check if Redis is running
redis-cli ping
# Should return: PONG

# Check Redis connection
redis-cli -h localhost -p 6379 ping
```

### Test Skipping

If PostgreSQL or Redis is not available, tests will skip with a message:
```
Skipping integration test: PostgreSQL not available at ...
```

This is expected behavior - tests gracefully skip when dependencies are unavailable.

## Environment Variables

You can customize test database connections:

```bash
export TEST_DATABASE_URL="postgres://user:pass@host/db?sslmode=disable"
export TEST_REDIS_URL="redis://localhost:6379/15"
```

## Continuous Integration

For CI/CD pipelines:

```yaml
# Example GitHub Actions
services:
  postgres:
    image: postgres:15
    env:
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: uniroute_test
    options: >-
      --health-cmd pg_isready
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5

  redis:
    image: redis:7
    options: >-
      --health-cmd "redis-cli ping"
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
```

