package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
)

// Integration tests require a running server
// These tests can be run with: go test -tags=integration

func TestTunnelServer_ClientConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	logger := zerolog.Nop()
	server := tunnel.NewTunnelServer(0, logger) // Random port

	// Start server
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Create a test HTTP server to simulate local server
	localServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from local server"))
	}))
	defer localServer.Close()

	// Create tunnel client
	client := tunnel.NewTunnelClient("localhost:8080", localServer.URL, logger)

	// Connect (this will fail without proper WebSocket setup)
	// For now, we'll test the connection logic separately
	_ = client

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Stop(ctx)
}

func TestTunnelServer_WebSocketUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Note: handleTunnelConnection is unexported, so we can't access it directly
	// For integration tests, we would need to start the server and connect via WebSocket
	// This test would require a full server setup
	t.Skip("Requires access to unexported handleTunnelConnection - test through Start() method")
}

func TestTunnelServer_RequestForwarding(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This test requires a full setup with:
	// 1. Tunnel server running
	// 2. Tunnel client connected
	// 3. Local HTTP server running
	// 4. HTTP request sent to tunnel server

	// Will be implemented in Phase 2
	t.Skip("Request forwarding test - Phase 2")
}
