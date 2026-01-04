package tunnel

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewTunnelClient(t *testing.T) {
	logger := zerolog.Nop()
	client := NewTunnelClient("localhost:8080", "http://localhost:8084", logger)
	
	assert.NotNil(t, client)
	assert.Equal(t, "localhost:8080", client.serverURL)
	assert.Equal(t, "http://localhost:8084", client.localURL)
	assert.NotNil(t, client.httpClient)
}

func TestTunnelClient_GetTunnelInfo(t *testing.T) {
	logger := zerolog.Nop()
	client := NewTunnelClient("localhost:8080", "http://localhost:8084", logger)
	
	// Initially nil
	info := client.GetTunnelInfo()
	assert.Nil(t, info)
	
	// Set tunnel info
	client.mu.Lock()
	client.tunnel = &TunnelInfo{
		ID:        "test-id",
		Subdomain: "test-sub",
		PublicURL: "http://test-sub.localhost:8080",
		Status:    "active",
	}
	client.mu.Unlock()
	
	info = client.GetTunnelInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "test-sub", info.Subdomain)
}

func TestTunnelClient_IsConnected(t *testing.T) {
	logger := zerolog.Nop()
	client := NewTunnelClient("localhost:8080", "http://localhost:8084", logger)
	
	// Initially false
	assert.False(t, client.IsConnected())
	
	// Set connected
	client.mu.Lock()
	client.isConnected = true
	client.mu.Unlock()
	
	assert.True(t, client.IsConnected())
}

func TestTunnelClient_ForwardToLocal(t *testing.T) {
	// Create a test HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "test response"}`))
	}))
	defer testServer.Close()
	
	logger := zerolog.Nop()
	client := NewTunnelClient("localhost:8080", testServer.URL, logger)
	
	// This test requires a real WebSocket connection
	// Will be tested in integration tests
	_ = client
}

func TestTunnelClient_SendResponse(t *testing.T) {
	// This requires a real WebSocket connection
	// Will be tested in integration tests
}

func TestTunnelClient_SendError(t *testing.T) {
	// This requires a real WebSocket connection
	// Will be tested in integration tests
}

func TestTunnelClient_Close(t *testing.T) {
	logger := zerolog.Nop()
	client := NewTunnelClient("localhost:8080", "http://localhost:8084", logger)
	
	// Close when not connected
	err := client.Close()
	assert.NoError(t, err)
	
	// Close when connected (no connection, so should still work)
	client.mu.Lock()
	client.isConnected = true
	client.mu.Unlock()
	
	err = client.Close()
	assert.NoError(t, err)
	assert.False(t, client.IsConnected())
}

