package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/gin-gonic/gin"
)

// TestServer wraps a test HTTP server
type TestServer struct {
	Server *httptest.Server
	Router *gin.Engine
}

// CreateTestServer creates a test HTTP server
func CreateTestServer(t *testing.T, setupRouter func(*gin.Engine)) *TestServer {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	if setupRouter != nil {
		setupRouter(router)
	}

	server := httptest.NewServer(router)

	return &TestServer{
		Server: server,
		Router: router,
	}
}

// Close closes the test server
func (ts *TestServer) Close() {
	if ts.Server != nil {
		ts.Server.Close()
	}
}

// MakeRequest makes an HTTP request to the test server
func MakeRequest(t *testing.T, method, url string, body interface{}, token string) *http.Response {
	var bodyReader io.Reader

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	// Use timeout for test client to prevent hanging tests
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	return resp
}

// ParseResponse parses JSON response body
func ParseResponse(t *testing.T, resp *http.Response, v interface{}) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("Failed to parse response: %v\nBody: %s", err, string(body))
	}
}

// AssertStatusCode asserts the response status code
func AssertStatusCode(t *testing.T, resp *http.Response, expected int) {
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status %d, got %d\nBody: %s", expected, resp.StatusCode, string(body))
	}
}

// CreateAuthToken creates a test JWT token
func CreateAuthToken(t *testing.T, userID, email string, roles []string) string {
	// Get JWT secret from config (or use default for tests)
	config := GetTestConfig(t)
	jwtService := security.NewJWTService(config.JWTSecret)

	// Ensure roles is not empty
	if roles == nil || len(roles) == 0 {
		roles = []string{"user"}
	}

	token, err := jwtService.GenerateToken(userID, email, roles, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	return token
}
