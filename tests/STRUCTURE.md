# Test Structure Overview

## Directory Tree

```
tests/
├── README.md                 # Main testing guide
├── MIGRATION.md              # Migration guide from old structure
├── STRUCTURE.md              # This file
├── .gitignore                # Test artifacts to ignore
│
├── unit/                     # Unit tests (fast, isolated)
│   ├── api/                  # API tests
│   ├── middleware/           # Middleware tests
│   ├── security/             # Security tests
│   ├── providers/            # Provider tests
│   ├── gateway/              # Gateway/router tests
│   └── storage/              # Storage layer tests
│
├── integration/              # Integration tests (require services)
│   ├── api/                  # API integration tests
│   ├── auth/                 # Authentication flow tests
│   ├── database/             # Database integration tests
│   └── redis/                # Redis integration tests
│
├── e2e/                      # End-to-end tests (full system)
│   ├── auth_flow.go         # Complete auth flow
│   ├── chat_flow.go         # Complete chat flow
│   └── admin_flow.go         # Admin operations flow
│
├── fixtures/                  # Test data and fixtures
│   ├── users.json           # Sample user data
│   ├── requests.json        # Sample API requests
│   └── responses.json       # Sample API responses
│
└── testutil/                  # Test utilities and helpers
    ├── README.md            # Utility documentation
    ├── setup.go             # Test setup/configuration
    ├── db.go                # Database test helpers
    ├── http.go              # HTTP test helpers
    ├── mocks.go             # Mock implementations
    └── fixtures.go          # Fixture loading helpers
```

## Test Type Comparison

| Type | Location | Speed | Dependencies | Use Case |
|------|----------|-------|--------------|----------|
| **Unit** | `tests/unit/` | Fast (<100ms) | None (mocks) | Test individual functions |
| **Integration** | `tests/integration/` | Medium (1-5s) | DB, Redis | Test component interactions |
| **E2E** | `tests/e2e/` | Slow (5-30s) | Full system | Test complete user flows |

## Quick Start

### Run All Tests
```bash
make test
```

### Run by Type
```bash
make test-unit         # Fast unit tests
make test-integration  # Integration tests
make test-e2e          # End-to-end tests
```

### Run with Coverage
```bash
make test-coverage
# Opens coverage.html
```

## File Naming Conventions

- Test files: `*_test.go`
- Package name: `package_name_test` (for external testing)
- Test functions: `TestFunctionName_Scenario_Expected`
- Example: `TestJWT_GenerateToken_WithValidClaims_ReturnsToken`

## Current Status

✅ **Structure Created**: All directories and utilities in place
✅ **Documentation**: Complete guides available
✅ **Makefile**: Updated with organized test commands
⏳ **Migration**: In progress (old tests still work)

## Next Steps

1. Start writing new tests in `tests/` directory
2. Gradually migrate existing tests
3. Use test utilities for common operations
4. Organize by test type (unit/integration/e2e)

