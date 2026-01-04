package tunnel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// TunnelClient connects local server to tunnel server
type TunnelClient struct {
	serverURL    string
	localURL     string
	wsConn       *websocket.Conn
	tunnel       *TunnelInfo
	httpClient   *http.Client
	logger       zerolog.Logger
	mu           sync.RWMutex
	reconnectMu  sync.Mutex
	isConnected  bool
	token        string
	requestQueue []*HTTPRequest // Queue for requests during disconnection
	queueMu      sync.Mutex
	subdomain    string             // Saved subdomain for resuming
	tunnelID     string             // Saved tunnel ID for resuming
	persistence  *TunnelPersistence // For saving/loading tunnel state
	latencyMs    int64              // Current latency in milliseconds
	latencyMu    sync.RWMutex       // Mutex for latency updates
}

// TunnelInfo is defined in types.go

// NewTunnelClient creates a new tunnel client
func NewTunnelClient(serverURL, localURL string, logger zerolog.Logger) *TunnelClient {
	client := &TunnelClient{
		serverURL: serverURL,
		localURL:  localURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:       logger,
		requestQueue: make([]*HTTPRequest, 0),
		persistence:  NewTunnelPersistence(logger),
	}

	// Try to load saved tunnel state
	if state, err := client.persistence.Load(); err == nil && state != nil {
		// Only use saved state if server URL matches (to avoid resuming wrong server)
		if state.ServerURL == serverURL {
			client.subdomain = state.Subdomain
			client.tunnelID = state.TunnelID
			logger.Info().
				Str("subdomain", state.Subdomain).
				Str("public_url", state.PublicURL).
				Msg("Loaded saved tunnel state - will attempt to resume")
		} else {
			logger.Debug().
				Str("saved_server", state.ServerURL).
				Str("current_server", serverURL).
				Msg("Saved tunnel state is for different server, will create new tunnel")
		}
	}

	return client
}

// Connect connects to the tunnel server
func (tc *TunnelClient) Connect() error {
	tc.reconnectMu.Lock()
	defer tc.reconnectMu.Unlock()

	// Connect to WebSocket and measure initial connection latency
	wsURL := fmt.Sprintf("ws://%s/tunnel", tc.serverURL)
	tc.logger.Info().Str("url", wsURL).Msg("Connecting to tunnel server")

	connectStart := time.Now()
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to tunnel server: %w", err)
	}
	tc.wsConn = conn

	// Measure initial connection latency
	connectLatency := time.Since(connectStart)
	tc.latencyMu.Lock()
	tc.latencyMs = connectLatency.Milliseconds()
	tc.latencyMu.Unlock()

	// Send initialization message
	// If we have a saved subdomain/tunnelID, try to resume
	initMsg := InitMessage{
		Type:     MsgTypeInit,
		Version:  "1.0",
		LocalURL: tc.localURL,
	}

	// Try to resume existing tunnel if we have subdomain/tunnelID
	if tc.subdomain != "" || tc.tunnelID != "" {
		initMsg.Subdomain = tc.subdomain
		initMsg.TunnelID = tc.tunnelID
		tc.logger.Info().
			Str("subdomain", tc.subdomain).
			Str("tunnel_id", tc.tunnelID).
			Msg("Attempting to resume existing tunnel")
	}

	if err := conn.WriteJSON(initMsg); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send init message: %w", err)
	}

	// Read response
	var response InitResponse
	if err := conn.ReadJSON(&response); err != nil {
		conn.Close()
		return fmt.Errorf("failed to read response: %w", err)
	}

	tc.mu.Lock()
	tc.tunnel = &TunnelInfo{
		ID:        response.TunnelID,
		Subdomain: response.Subdomain,
		PublicURL: response.PublicURL,
		Status:    response.Status,
	}
	// Save subdomain and tunnel ID for future reconnections
	tc.subdomain = response.Subdomain
	tc.tunnelID = response.TunnelID
	tc.isConnected = true
	tc.mu.Unlock()

	// Save tunnel state to file for persistence across restarts
	if tc.persistence != nil {
		state := &TunnelState{
			TunnelID:  response.TunnelID,
			Subdomain: response.Subdomain,
			PublicURL: response.PublicURL,
			LocalURL:  tc.localURL,
			ServerURL: tc.serverURL,
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
		}
		if err := tc.persistence.Save(state); err != nil {
			tc.logger.Warn().Err(err).Msg("Failed to save tunnel state")
		}
	}

	// Process any queued requests after reconnection
	go tc.processQueuedRequests()

	tc.logger.Info().
		Str("tunnel_id", response.TunnelID).
		Str("subdomain", response.Subdomain).
		Str("public_url", response.PublicURL).
		Str("local_url", tc.localURL).
		Msg("Tunnel connected")

	// Start handling messages
	go tc.handleMessages()

	// Start heartbeat
	go tc.heartbeat()

	return nil
}

// handleMessages handles messages from tunnel server
func (tc *TunnelClient) handleMessages() {
	for {
		var msg TunnelMessage
		if err := tc.wsConn.ReadJSON(&msg); err != nil {
			tc.logger.Error().Err(err).Msg("Connection lost")
			tc.mu.Lock()
			tc.isConnected = false
			tc.mu.Unlock()
			// Attempt to reconnect
			tc.reconnect()
			return
		}

		switch msg.Type {
		case MsgTypeHTTPRequest:
			// Forward request to local server
			if msg.Request != nil {
				if tc.IsConnected() {
					go tc.forwardToLocal(msg.Request)
				} else {
					// Queue request if disconnected
					tc.queueMu.Lock()
					tc.requestQueue = append(tc.requestQueue, msg.Request)
					tc.queueMu.Unlock()
					tc.logger.Warn().Str("request_id", msg.Request.RequestID).Msg("Request queued - not connected")
				}
			}
		case MsgTypePong:
			// Heartbeat response - latency is measured in heartbeat function
			tc.logger.Debug().Msg("Received pong")
		case MsgTypeTunnelStatus:
			// Status update
			tc.logger.Debug().Msg("Received tunnel status update")
		default:
			tc.logger.Warn().Str("type", msg.Type).Msg("Unknown message type")
		}
	}
}

// forwardToLocal forwards request to local server
func (tc *TunnelClient) forwardToLocal(req *HTTPRequest) {
	if req == nil {
		tc.logger.Error().Msg("Received nil request")
		return
	}

	// Build local URL
	localURL := tc.localURL + req.Path
	if req.Query != "" {
		localURL += "?" + req.Query
	}

	tc.logger.Debug().
		Str("request_id", req.RequestID).
		Str("method", req.Method).
		Str("path", req.Path).
		Str("local_url", localURL).
		Msg("Forwarding request to local server")

	// Reconstruct HTTP request
	httpReq, err := http.NewRequest(req.Method, localURL, bytes.NewReader(req.Body))
	if err != nil {
		tc.logger.Error().Err(err).Str("request_id", req.RequestID).Msg("Failed to create request")
		tc.sendError(req.RequestID, "failed to create request", err.Error())
		return
	}

	// Copy headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Forward to local server
	resp, err := tc.httpClient.Do(httpReq)
	if err != nil {
		tc.logger.Error().Err(err).Str("request_id", req.RequestID).Msg("Failed to forward request")
		tc.sendError(req.RequestID, "connection_refused", err.Error())
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		tc.logger.Error().Err(err).Str("request_id", req.RequestID).Msg("Failed to read response body")
		tc.sendError(req.RequestID, "read_error", err.Error())
		return
	}

	// Build response
	response := &HTTPResponse{
		RequestID: req.RequestID,
		Status:    resp.StatusCode,
		Headers:   make(map[string]string),
		Body:      body,
	}

	// Copy headers
	for k, v := range resp.Header {
		if len(v) > 0 {
			response.Headers[k] = v[0]
		}
	}

	// Send response back through tunnel
	tc.sendResponse(response)

	tc.logger.Debug().
		Str("request_id", req.RequestID).
		Int("status", resp.StatusCode).
		Int("body_size", len(body)).
		Msg("Request completed")
}

// sendResponse sends response through tunnel
func (tc *TunnelClient) sendResponse(resp *HTTPResponse) {
	if resp == nil {
		return
	}

	msg := TunnelMessage{
		Type:      MsgTypeHTTPResponse,
		RequestID: resp.RequestID,
		Response:  resp,
	}

	tc.mu.RLock()
	conn := tc.wsConn
	tc.mu.RUnlock()

	if conn == nil {
		tc.logger.Error().Str("request_id", resp.RequestID).Msg("No connection to send response")
		return
	}

	if err := conn.WriteJSON(msg); err != nil {
		tc.logger.Error().Err(err).Str("request_id", resp.RequestID).Msg("Failed to send response")
	}
}

// sendError sends error response through tunnel
func (tc *TunnelClient) sendError(requestID, errorType, message string) {
	msg := TunnelMessage{
		Type:      MsgTypeHTTPError,
		RequestID: requestID,
		Error: &HTTPError{
			RequestID: requestID,
			Error:     errorType,
			Message:   message,
		},
	}

	tc.mu.RLock()
	conn := tc.wsConn
	tc.mu.RUnlock()

	if conn == nil {
		return
	}

	if err := conn.WriteJSON(msg); err != nil {
		tc.logger.Error().Err(err).Str("request_id", requestID).Msg("Failed to send error")
	}
}

// heartbeat sends periodic ping to keep connection alive
func (tc *TunnelClient) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tc.mu.RLock()
			conn := tc.wsConn
			isConnected := tc.isConnected
			tc.mu.RUnlock()

			if !isConnected || conn == nil {
				return
			}

			// Measure latency by timing ping
			start := time.Now()
			if err := conn.WriteJSON(TunnelMessage{Type: MsgTypePing}); err != nil {
				tc.logger.Error().Err(err).Msg("Heartbeat failed")
				return
			}

			// Update latency (rough estimate - actual RTT would require pong response timing)
			// For now, use a simple measurement of ping send time
			latency := time.Since(start)
			tc.latencyMu.Lock()
			tc.latencyMs = latency.Milliseconds()
			tc.latencyMu.Unlock()
		}
	}
}

// reconnect attempts to reconnect to tunnel server
func (tc *TunnelClient) reconnect() {
	tc.reconnectMu.Lock()
	defer tc.reconnectMu.Unlock()

	// Check if already reconnecting
	if !tc.isConnected {
		backoff := 5 * time.Second
		maxBackoff := 60 * time.Second

		for {
			tc.logger.Info().
				Dur("backoff", backoff).
				Str("subdomain", tc.subdomain).
				Msg("Attempting to reconnect...")
			time.Sleep(backoff)

			if err := tc.Connect(); err == nil {
				tc.logger.Info().
					Str("subdomain", tc.subdomain).
					Str("public_url", tc.tunnel.PublicURL).
					Msg("Reconnected successfully - resumed same tunnel")
				return
			}

			// Exponential backoff
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

// processQueuedRequests processes requests that were queued during disconnection
func (tc *TunnelClient) processQueuedRequests() {
	tc.queueMu.Lock()
	queued := make([]*HTTPRequest, len(tc.requestQueue))
	copy(queued, tc.requestQueue)
	tc.requestQueue = tc.requestQueue[:0] // Clear queue
	tc.queueMu.Unlock()

	if len(queued) > 0 {
		tc.logger.Info().Int("count", len(queued)).Msg("Processing queued requests after reconnection")
		for _, req := range queued {
			go tc.forwardToLocal(req)
		}
	}
}

// Close closes the tunnel connection
func (tc *TunnelClient) Close() error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.isConnected = false
	if tc.wsConn != nil {
		err := tc.wsConn.Close()
		tc.wsConn = nil
		return err
	}
	return nil
}

// GetTunnelInfo returns tunnel information
func (tc *TunnelClient) GetTunnelInfo() *TunnelInfo {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.tunnel
}

// IsConnected returns connection status
func (tc *TunnelClient) IsConnected() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.isConnected
}

// GetLatency returns the current latency in milliseconds
func (tc *TunnelClient) GetLatency() int64 {
	tc.latencyMu.RLock()
	defer tc.latencyMu.RUnlock()
	return tc.latencyMs
}

// GetConnectionStats fetches connection statistics from the tunnel server
func (tc *TunnelClient) GetConnectionStats(serverURL string, tunnelID string) (*ConnectionStats, error) {
	tc.mu.RLock()
	token := tc.token
	tc.mu.RUnlock()

	// Make HTTP request to stats endpoint
	url := fmt.Sprintf("http://%s/api/tunnels/%s/stats", serverURL, tunnelID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get stats: %d", resp.StatusCode)
	}

	var result struct {
		Connections ConnectionStats `json:"connections"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Connections, nil
}
