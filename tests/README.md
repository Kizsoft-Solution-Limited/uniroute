# UniRoute Testing Guide

This directory contains all tests organized by type and purpose.

## Directory Structure

```
tests/
├── README.md              # This file
├── unit/                  # Unit tests (fast, isolated)
│   ├── api/              # API handler unit tests
│   ├── middleware/       # Middleware unit tests
│   ├── security/         # Security component tests
│   ├── providers/        # Provider tests
│   ├── gateway/          # Gateway/router tests
│   └── storage/          # Storage layer tests
├── integration/          # Integration tests (require services)
│   ├── api/              # API integration tests
│   ├── auth/             # Authentication flow tests
│   ├── database/         # Database integration tests
│   └── redis/            # Redis integration tests
├── e2e/                  # End-to-end tests (full system)
│   ├── auth_flow.go     # Complete auth flow
│   ├── chat_flow.go     # Complete chat flow
│   └── admin_flow.go     # Admin operations flow
├── fixtures/             # Test data and fixtures
│   ├── users.json       # Sample user data
│   ├── requests.json    # Sample API requests
│   └── responses.json   # Sample API responses
└── testutil/             # Test utilities and helpers
    ├── setup.go          # Test setup/teardown
    ├── mocks.go          # Mock implementations
    ├── db.go             # Database test helpers
    └── http.go           # HTTP test helpers
```

## Test Types

### Unit Tests (`tests/unit/`)

**Purpose**: Test individual functions, methods, and components in isolation.

**Characteristics**:
- Fast execution (< 100ms per test)
- No external dependencies (no DB, Redis, network)
- Use mocks for dependencies
- Test one thing at a time

**Example**:
```go
// tests/unit/security/jwt_test.go
func TestJWT_GenerateToken(t *testing.T) {
    // Test JWT token generation in isolation
}
```

**Run**: `make test-unit` or `go test ./tests/unit/...`

### Integration Tests (`tests/integration/`)

**Purpose**: Test component interactions with real services.

**Characteristics**:
- Require external services (PostgreSQL, Redis)
- Test component interactions
- Slower than unit tests
- Use test databases/instances

**Example**:
```go
// tests/integration/api/auth_test.go
func TestAuthFlow_RegisterAndLogin(t *testing.T) {
    // Test complete registration and login flow
}
```

**Run**: `make test-integration` or `go test ./tests/integration/...`

### End-to-End Tests (`tests/e2e/`)

**Purpose**: Test complete user flows from start to finish.

**Characteristics**:
- Test full system (all services running)
- Test real user scenarios
- Slowest tests
- Require full environment setup

**Example**:
```go
// tests/e2e/auth_flow.go
func TestE2E_UserRegistrationToChat(t *testing.T) {
    // Test: Register -> Verify Email -> Login -> Create API Key -> Send Chat Request
}
```

**Run**: `make test-e2e` or `go test ./tests/e2e/...`

## Running Tests

### All Tests
```bash
make test              # Run all tests
make test-coverage     # Run with coverage report
```

### By Type
```bash
make test-unit         # Unit tests only
make test-integration  # Integration tests only
make test-e2e          # E2E tests only
```

### Specific Package
```bash
go test ./tests/unit/security/... -v
go test ./tests/integration/api/... -v
```

### With Coverage
```bash
make test-coverage
# Opens coverage.html in browser
```

### With Race Detector
```bash
make test-race
```

## Test Utilities

### Database Helpers (`tests/testutil/db.go`)

```go
// SetupTestDB creates a test database connection
db := testutil.SetupTestDB(t)
defer testutil.CleanupTestDB(t, db)

// CreateTestUser creates a test user
user := testutil.CreateTestUser(t, db, "test@example.com")
```

### HTTP Helpers (`tests/testutil/http.go`)

```go
// CreateTestServer creates a test HTTP server
server := testutil.CreateTestServer(t)
defer server.Close()

// MakeRequest makes an authenticated request
resp := testutil.MakeRequest(t, server, "POST", "/auth/login", body, token)
```

### Mock Helpers (`tests/testutil/mocks.go`)

```go
// CreateMockEmailService creates a mock email service
emailSvc := testutil.CreateMockEmailService(t)

// CreateMockUserRepo creates a mock user repository
userRepo := testutil.CreateMockUserRepo(t)
```

## Test Data

### Fixtures (`tests/fixtures/`)

Store reusable test data in JSON files:

```json
// tests/fixtures/users.json
{
  "valid_user": {
    "email": "test@example.com",
    "password": "SecurePass123!",
    "name": "Test User"
  }
}
```

Load in tests:
```go
userData := testutil.LoadFixture(t, "users.json", "valid_user")
```

## Best Practices

1. **Naming**: Use descriptive test names
   - ✅ `TestJWT_GenerateToken_WithValidClaims_ReturnsToken`
   - ❌ `TestJWT1`

2. **Isolation**: Each test should be independent
   - Don't rely on test execution order
   - Clean up after each test

3. **Table-Driven Tests**: Use for multiple scenarios
   ```go
   tests := []struct{
       name string
       input string
       want string
   }{
       {"case1", "input1", "output1"},
       {"case2", "input2", "output2"},
   }
   ```

4. **Mocks**: Mock external dependencies
   - Don't make real HTTP calls in unit tests
   - Don't connect to real databases in unit tests

5. **Test Coverage**: Aim for >80% coverage
   - Focus on critical paths (auth, routing, security)
   - Don't obsess over 100% coverage

## CI/CD Integration

Tests run automatically in CI/CD:

```yaml
# .github/workflows/test.yml
- name: Run unit tests
  run: make test-unit

- name: Run integration tests
  run: make test-integration
  env:
    POSTGRES_URL: ${{ secrets.TEST_POSTGRES_URL }}
    REDIS_URL: ${{ secrets.TEST_REDIS_URL }}
```

## Troubleshooting

### Tests failing with "connection refused"
- Ensure test services are running (PostgreSQL, Redis)
- Check test configuration in `tests/testutil/setup.go`

### Tests timing out
- Integration/E2E tests may take longer
- Increase timeout: `go test -timeout 10m ./tests/integration/...`

### Coverage not generating
- Ensure tests are actually running
- Check `coverage.out` file is created
- Run `go tool cover -html=coverage.out`

## Migration from Old Structure

Old tests are co-located with source files (e.g., `internal/api/middleware/jwt_test.go`).

**Migration Plan**:
1. Keep existing tests working
2. Gradually move to new structure
3. New tests go in `tests/` directory
4. Eventually migrate all tests

**Current Status**:
- ✅ New structure created
- ✅ Test utilities available
- ⏳ Migration in progress

