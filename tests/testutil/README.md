# Test Utilities

This package provides utilities and helpers for writing tests.

## Available Utilities

### Setup (`setup.go`)
- `GetTestConfig()` - Load test configuration
- `SkipIfShort()` - Skip test in short mode
- `SkipIfNoPostgres()` - Skip if PostgreSQL unavailable
- `SkipIfNoRedis()` - Skip if Redis unavailable

### Database (`db.go`)
- `SetupTestDB()` - Create test database connection
- `CleanupTestDB()` - Clean up test database
- `CreateTestUser()` - Create a test user
- `CreateTestAPIKey()` - Create a test API key

### HTTP (`http.go`)
- `CreateTestServer()` - Create test HTTP server
- `MakeRequest()` - Make HTTP request
- `ParseResponse()` - Parse JSON response
- `AssertStatusCode()` - Assert status code
- `CreateAuthToken()` - Create test JWT token

### Mocks (`mocks.go`)
- `CreateMockEmailService()` - Create mock email service
- `MockEmailService` - Mock email service implementation

## Usage Example

```go
package integration

import (
    "testing"
    "github.com/Kizsoft-Solution-Limited/uniroute/tests/testutil"
)

func TestAuthFlow(t *testing.T) {
    // Setup
    db := testutil.SetupTestDB(t)
    defer testutil.CleanupTestDB(t, db)
    
    // Create test user
    userID := testutil.CreateTestUser(t, db, "test@example.com", "password", "Test User")
    
    // Create test server
    server := testutil.CreateTestServer(t, func(r *gin.Engine) {
        // Setup your router
    })
    defer server.Close()
    
    // Make request
    resp := testutil.MakeRequest(t, "POST", server.URL+"/auth/login", 
        map[string]string{"email": "test@example.com", "password": "password"}, "")
    
    // Assert
    testutil.AssertStatusCode(t, resp, 200)
}
```

