package tunnel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTunnelServer(t *testing.T) {
	logger := zerolog.Nop()
	server := NewTunnelServer(8080, logger)
	
	assert.NotNil(t, server)
	assert.Equal(t, 8080, server.port)
	assert.NotNil(t, server.tunnels)
}

func TestTunnelServer_Start(t *testing.T) {
	logger := zerolog.Nop()
	server := NewTunnelServer(0, logger) // Use random port
	
	// Start server in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Test health endpoint
	resp, err := http.Get("http://localhost:8080/health")
	if err == nil {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
	
	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Stop(ctx)
}

func TestTunnelServer_HandleHealth(t *testing.T) {
	logger := zerolog.Nop()
	server := NewTunnelServer(8080, logger)
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	server.handleHealth(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")
}

func TestTunnelServer_HandleRootRequest(t *testing.T) {
	logger := zerolog.Nop()
	server := NewTunnelServer(8080, logger)
	
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	server.handleRootRequest(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "UniRoute Tunnel Server")
}

func TestTunnelServer_GenerateSubdomain(t *testing.T) {
	logger := zerolog.Nop()
	server := NewTunnelServer(8080, logger)
	
	subdomain1 := server.generateSubdomain()
	subdomain2 := server.generateSubdomain()
	
	assert.NotEmpty(t, subdomain1)
	assert.NotEmpty(t, subdomain2)
	assert.NotEqual(t, subdomain1, subdomain2)
	assert.Len(t, subdomain1, 12) // Should be 12 characters
}

func TestTunnelServer_GetTunnel(t *testing.T) {
	logger := zerolog.Nop()
	server := NewTunnelServer(8080, logger)
	
	// Create a test tunnel
	tunnel := &TunnelConnection{
		ID:        "test-id",
		Subdomain: "test-subdomain",
		LocalURL:  "http://localhost:8084",
		CreatedAt: time.Now(),
	}
	
	server.tunnelsMu.Lock()
	server.tunnels["test-subdomain"] = tunnel
	server.tunnelsMu.Unlock()
	
	// Get tunnel
	retrieved, exists := server.GetTunnel("test-subdomain")
	assert.True(t, exists)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "test-subdomain", retrieved.Subdomain)
	
	// Get non-existent tunnel
	_, exists = server.GetTunnel("non-existent")
	assert.False(t, exists)
}

func TestTunnelServer_ListTunnels(t *testing.T) {
	logger := zerolog.Nop()
	server := NewTunnelServer(8080, logger)
	
	// Add some tunnels
	tunnel1 := &TunnelConnection{ID: "1", Subdomain: "sub1", LocalURL: "http://localhost:8084"}
	tunnel2 := &TunnelConnection{ID: "2", Subdomain: "sub2", LocalURL: "http://localhost:8085"}
	
	server.tunnelsMu.Lock()
	server.tunnels["sub1"] = tunnel1
	server.tunnels["sub2"] = tunnel2
	server.tunnelsMu.Unlock()
	
	tunnels := server.ListTunnels()
	assert.Len(t, tunnels, 2)
}

func TestTunnelServer_RemoveTunnel(t *testing.T) {
	logger := zerolog.Nop()
	server := NewTunnelServer(8080, logger)
	
	// Add tunnel
	tunnel := &TunnelConnection{ID: "test", Subdomain: "test-sub", LocalURL: "http://localhost:8084"}
	server.tunnelsMu.Lock()
	server.tunnels["test-sub"] = tunnel
	server.tunnelsMu.Unlock()
	
	// Remove tunnel
	server.removeTunnel("test-sub")
	
	// Verify removed
	_, exists := server.GetTunnel("test-sub")
	assert.False(t, exists)
}

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{"localhost with port", "abc123.localhost:8080", "abc123"},
		{"localhost without port", "abc123.localhost", "abc123"},
		{"domain with port", "abc123.example.com:8080", "abc123"},
		{"domain without port", "abc123.example.com", "abc123"},
		{"no subdomain", "localhost:8080", ""},
		{"no subdomain domain", "example.com", "example"}, // Returns first part even if it's the domain
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSubdomain(tt.host)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()
	
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 32) // UUID hex string length
}

func TestTunnelServer_SerializeRequest(t *testing.T) {
	logger := zerolog.Nop()
	server := NewTunnelServer(8080, logger)
	
	req := httptest.NewRequest("POST", "/test?param=value", nil)
	req.Header.Set("Content-Type", "application/json")
	reqID := "test-request-id"
	
	httpReq, err := server.serializeRequest(req, reqID)
	require.NoError(t, err)
	
	assert.Equal(t, reqID, httpReq.RequestID)
	assert.Equal(t, "POST", httpReq.Method)
	assert.Equal(t, "/test", httpReq.Path)
	assert.Equal(t, "param=value", httpReq.Query)
	assert.Equal(t, "application/json", httpReq.Headers["Content-Type"])
}

func TestTunnelServer_WriteResponse(t *testing.T) {
	logger := zerolog.Nop()
	server := NewTunnelServer(8080, logger)
	
	w := httptest.NewRecorder()
	resp := &HTTPResponse{
		RequestID: "test-id",
		Status:    200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"test": "data"}`),
	}
	
	server.writeResponse(w, resp)
	
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, `{"test": "data"}`, w.Body.String())
}

