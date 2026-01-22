package tunnel

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// RequestEvent represents an HTTP request event for CLI display
type RequestEvent struct {
	Time       time.Time
	Method     string
	Path       string
	StatusCode int
	StatusText string
}

// RequestEventHandler is a callback function for HTTP request events
type RequestEventHandler func(event RequestEvent)

// TunnelClient connects local server to tunnel server
type TunnelClient struct {
	serverURL      string
	localURL       string
	protocol       string // http, tcp, tls, udp
	host           string // Optional: specific host/subdomain
	wsConn         *websocket.Conn
	tunnel         *TunnelInfo
	httpClient     *http.Client
	logger         zerolog.Logger
	mu             sync.RWMutex
	reconnectMu    sync.Mutex
	isConnected    bool
	isReconnecting bool // True when attempting to reconnect
	token          string
	requestQueue   []*HTTPRequest // Queue for requests during disconnection
	queueMu        sync.Mutex
	tcpConnections map[string]net.Conn // Track active TCP/TLS connections
	tcpConnMu      sync.RWMutex
	udpConn        *net.UDPConn   // UDP connection to local service
	udpConnMu      sync.RWMutex
	subdomain      string             // Saved subdomain for resuming
	tunnelID       string             // Saved tunnel ID for resuming
	persistence    *TunnelPersistence // For saving/loading tunnel state
	latencyMs      int64              // Current latency in milliseconds
	latencyMu      sync.RWMutex       // Mutex for latency updates
	lastPongTime   time.Time          // Last time we received a pong
	pongMu         sync.RWMutex       // Mutex for lastPongTime
	requestHandler RequestEventHandler // Callback for request events
	requestHandlerMu sync.RWMutex      // Mutex for request handler
}

// TunnelInfo is defined in types.go

// NewTunnelClient creates a new tunnel client
func NewTunnelClient(serverURL, localURL string, logger zerolog.Logger) *TunnelClient {
	return NewTunnelClientWithOptions(serverURL, localURL, "http", "", logger)
}

// NewTunnelClientWithOptions creates a new tunnel client with protocol and host options
func NewTunnelClientWithOptions(serverURL, localURL, protocol, host string, logger zerolog.Logger) *TunnelClient {
	// Default to http if protocol not specified
	if protocol == "" {
		protocol = ProtocolHTTP
	}

	client := &TunnelClient{
		serverURL: serverURL,
		localURL:  localURL,
		protocol:  protocol,
		host:      host,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:         logger,
		requestQueue:   make([]*HTTPRequest, 0),
		tcpConnections: make(map[string]net.Conn),
		persistence:    NewTunnelPersistence(logger),
	}

	// Try to load saved tunnel state
	if state, err := client.persistence.Load(); err == nil && state != nil {
		// Normalize server URLs for comparison (remove http://, https://, ws://, wss:// prefixes)
		savedServer := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(state.ServerURL, "http://"), "https://"), "ws://"), "wss://")
		currentServer := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(serverURL, "http://"), "https://"), "ws://"), "wss://")

		// Only use saved state if server URL matches (to avoid resuming wrong server)
		// Also check if protocol matches (if saved state has protocol)
		protocolMatches := state.Protocol == "" || state.Protocol == protocol

		if savedServer == currentServer && protocolMatches {
			client.subdomain = state.Subdomain
			client.tunnelID = state.TunnelID
			logger.Debug().
				Str("subdomain", state.Subdomain).
				Str("tunnel_id", state.TunnelID).
				Str("public_url", state.PublicURL).
				Msg("Loaded saved tunnel state - will attempt to resume existing tunnel")
		} else {
			logger.Debug().
				Str("saved_server", state.ServerURL).
				Str("current_server", serverURL).
				Str("saved_protocol", state.Protocol).
				Str("current_protocol", protocol).
				Msg("Saved tunnel state is for different server/protocol, will create new tunnel")
		}
	}

	return client
}

// SetResumeInfo sets subdomain and tunnel ID for resumption
func (tc *TunnelClient) SetResumeInfo(subdomain, tunnelID string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.subdomain = subdomain
	tc.tunnelID = tunnelID
}

// SetToken sets the authentication token for the tunnel client
func (tc *TunnelClient) SetToken(token string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.token = token
}

// SetRequestHandler sets a callback function to be called for each HTTP request
func (tc *TunnelClient) SetRequestHandler(handler RequestEventHandler) {
	tc.requestHandlerMu.Lock()
	defer tc.requestHandlerMu.Unlock()
	tc.requestHandler = handler
}

// ClearResumeInfo clears saved subdomain and tunnel ID to force new tunnel creation
func (tc *TunnelClient) ClearResumeInfo() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.subdomain = ""
	tc.tunnelID = ""
}

// Connect connects to the tunnel server
func (tc *TunnelClient) Connect() error {
	tc.reconnectMu.Lock()
	defer tc.reconnectMu.Unlock()

	// Connect to WebSocket and measure initial connection latency
	wsURL := fmt.Sprintf("ws://%s/tunnel", tc.serverURL)
	tc.logger.Debug().Str("url", wsURL).Msg("Connecting to tunnel server")

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
	tc.mu.RLock()
	authToken := tc.token
	tc.mu.RUnlock()

	initMsg := InitMessage{
		Type:     MsgTypeInit,
		Version:  "1.0",
		Protocol: tc.protocol,
		LocalURL: tc.localURL,
		Host:     tc.host,
		Token:    authToken, // Include auth token if available
	}

	// Try to resume existing tunnel if we have subdomain/tunnelID
	if tc.subdomain != "" || tc.tunnelID != "" {
		initMsg.Subdomain = tc.subdomain
		initMsg.TunnelID = tc.tunnelID
		tc.logger.Debug().
			Str("subdomain", tc.subdomain).
			Str("tunnel_id", tc.tunnelID).
			Str("protocol", tc.protocol).
			Msg("Attempting to resume existing tunnel")
	} else {
		tc.logger.Debug().
			Str("protocol", tc.protocol).
			Str("local_url", tc.localURL).
			Str("host", tc.host).
			Msg("Creating new tunnel")
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
	tc.isReconnecting = false // Clear reconnecting flag on successful connection
	tc.mu.Unlock()

	// Initialize last pong time for heartbeat monitoring
	tc.pongMu.Lock()
	tc.lastPongTime = time.Now()
	tc.pongMu.Unlock()

	// Save tunnel state to file for persistence across restarts
	// This allows the tunnel to be automatically resumed on next run
	if tc.persistence != nil {
		state := &TunnelState{
			TunnelID:  response.TunnelID,
			Subdomain: response.Subdomain,
			PublicURL: response.PublicURL,
			LocalURL:  tc.localURL,
			ServerURL: tc.serverURL,
			Protocol:  tc.protocol,
			Host:      tc.host,
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
		}
		if err := tc.persistence.Save(state); err != nil {
			tc.logger.Warn().Err(err).Msg("Failed to save tunnel state")
		} else {
			tc.logger.Debug().
				Str("subdomain", response.Subdomain).
				Str("tunnel_id", response.TunnelID).
				Str("server_url", tc.serverURL).
				Msg("Saved tunnel state for auto-resume on next run")
		}
	}

	// Process any queued requests after reconnection
	go tc.processQueuedRequests()

	tc.logger.Debug().
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
	// Set read deadline to detect network disconnection faster
	// Use 40 seconds (just over one heartbeat interval of 30s) to detect dead connections quickly
	readDeadline := 40 * time.Second
	
	for {
		// Set read deadline before each read to detect network disconnection
		tc.mu.RLock()
		conn := tc.wsConn
		tc.mu.RUnlock()
		
		if conn == nil {
			return
		}
		
		conn.SetReadDeadline(time.Now().Add(readDeadline))
		
		var msg TunnelMessage
		if err := conn.ReadJSON(&msg); err != nil {
			errStr := err.Error()
			
			// Check if this is a normal shutdown (close 1000) - don't log or reconnect
			if strings.Contains(errStr, "websocket: close 1000") || strings.Contains(errStr, "close 1000 (normal)") {
				// Normal shutdown - just return without logging
				return
			}
			
			// Check if this is a real connection error that requires reconnection
			// "close 1006" (abnormal closure) with "unexpected EOF" indicates the server closed the connection
			// This should only happen on actual network disconnections, not when local server is down
			isConnectionError := false
			
			// WebSocket close codes indicate actual connection closure (but not normal 1000)
			if strings.Contains(errStr, "websocket: close") {
				// All WebSocket close codes indicate connection was closed (except 1000 which we already handled)
				isConnectionError = true
			}
			
			// Network-level errors that indicate connection is broken
			if strings.Contains(errStr, "connection reset") ||
				strings.Contains(errStr, "broken pipe") ||
				strings.Contains(errStr, "use of closed network connection") ||
				strings.Contains(errStr, "connection refused") && !strings.Contains(errStr, "localhost") {
				// "connection refused" for localhost is from local server being down, not network issue
				isConnectionError = true
			}
			
			// Timeout errors after read deadline - connection is likely dead
			if strings.Contains(errStr, "i/o timeout") || strings.Contains(errStr, "deadline exceeded") {
				isConnectionError = true
			}
			
			// Only reconnect on actual connection errors (network disconnection)
			// Errors from local server being down should NOT trigger reconnection
			if isConnectionError {
				tc.logger.Debug().Err(err).Msg("Connection lost - attempting to reconnect")
				tc.mu.Lock()
				tc.isConnected = false
				tc.mu.Unlock()
				// Attempt to reconnect
				tc.reconnect()
				return
			} else {
				// Log other errors (like unexpected EOF without close code) at debug level
				// These might be temporary or related to local server issues, not network disconnection
				tc.logger.Debug().Err(err).Msg("Read error (not a connection failure) - continuing")
				// Continue reading - connection might still be alive
				continue
			}
		}
		
		// Reset read deadline after successful read (connection is alive)
		conn.SetReadDeadline(time.Time{}) // Clear deadline

		switch msg.Type {
		case MsgTypeHTTPRequest:
			// Forward HTTP request to local server
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
		case MsgTypeTCPData, MsgTypeTLSData:
			// Handle TCP/TLS data
			if len(msg.Data) > 0 || msg.RequestID != "" {
				go tc.handleTCPData(msg.RequestID, msg.Data, msg.Type == MsgTypeTLSData)
			}
		case MsgTypeTCPError, MsgTypeTLSError:
			// Handle TCP/TLS errors
			if msg.Error != nil {
				tc.handleTCPError(msg.RequestID, msg.Error)
			}
		case MsgTypeUDPData:
			// Handle UDP data
			if len(msg.Data) > 0 && msg.RequestID != "" {
				go tc.handleUDPData(msg.RequestID, msg.Data)
			}
		case MsgTypeUDPError:
			// Handle UDP errors
			if msg.Error != nil {
				tc.handleUDPError(msg.RequestID, msg.Error)
			}
		case MsgTypePong:
			// Heartbeat response - update last pong time
			tc.pongMu.Lock()
			tc.lastPongTime = time.Now()
			tc.pongMu.Unlock()
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

	// Filter out static asset requests - don't notify handler for these
	staticAssetPaths := []string{
		"/favicon.ico",
		"/favicon.png",
		"/robots.txt",
		"/.well-known/",
	}
	
	path := req.Path
	if path == "" {
		path = "/"
	}
	
	isStaticAsset := false
	for _, staticPath := range staticAssetPaths {
		if path == staticPath || strings.HasPrefix(path, staticPath) {
			isStaticAsset = true
			break
		}
	}
	
	// Also check for image/icon file extensions
	if !isStaticAsset {
		lowerPath := strings.ToLower(path)
		staticExtensions := []string{".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg", ".webp", ".woff", ".woff2", ".ttf", ".eot"}
		for _, ext := range staticExtensions {
			if strings.HasSuffix(lowerPath, ext) {
				isStaticAsset = true
				break
			}
		}
	}

	// Build local URL
	localURL := tc.localURL + path
	if req.Query != "" {
		localURL += "?" + req.Query
	}

	tc.logger.Debug().
		Str("request_id", req.RequestID).
		Str("method", req.Method).
		Str("path", req.Path).
		Str("local_url", localURL).
		Bool("is_static_asset", isStaticAsset).
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
	tc.logger.Debug().
		Str("request_id", req.RequestID).
		Str("local_url", localURL).
		Msg("Sending HTTP request to local server")
	
	resp, err := tc.httpClient.Do(httpReq)
	if err != nil {
		// Log at debug level to avoid cluttering CLI output
		tc.logger.Debug().
			Err(err).
			Str("request_id", req.RequestID).
			Str("local_url", localURL).
			Msg("Failed to forward request to local server - connection error")
		
		// Notify request handler of error (502 Bad Gateway) - skip static assets
		if !isStaticAsset {
			tc.requestHandlerMu.RLock()
			handler := tc.requestHandler
			tc.requestHandlerMu.RUnlock()
			
			if handler != nil {
				handler(RequestEvent{
					Time:       time.Now(),
					Method:     req.Method,
					Path:       req.Path,
					StatusCode: 502,
					StatusText: "Bad Gateway",
				})
			}
		}
		
		tc.sendError(req.RequestID, "connection_refused", err.Error())
		return
	}
	defer resp.Body.Close()

	// Log at debug level to avoid cluttering CLI output
	tc.logger.Debug().
		Str("request_id", req.RequestID).
		Str("local_url", localURL).
		Int("status_code", resp.StatusCode).
		Str("status", resp.Status).
		Int("header_count", len(resp.Header)).
		Msg("Received response from local server")

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

	// Notify request handler if set (skip static assets)
	if !isStaticAsset {
		tc.requestHandlerMu.RLock()
		handler := tc.requestHandler
		tc.requestHandlerMu.RUnlock()
		
		if handler != nil {
			handler(RequestEvent{
				Time:       time.Now(),
				Method:     req.Method,
				Path:       req.Path,
				StatusCode: resp.StatusCode,
				StatusText: resp.Status,
			})
		}
	}

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

// heartbeat sends periodic ping to keep connection alive and detects connection loss
func (tc *TunnelClient) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

		// Check for missing pong responses (connection loss detection)
		// Check every 5 seconds to detect disconnection faster
		pongCheckTicker := time.NewTicker(5 * time.Second)
		defer pongCheckTicker.Stop()

	// Initialize last pong time
	tc.pongMu.Lock()
	tc.lastPongTime = time.Now()
	tc.pongMu.Unlock()

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

			// Set write deadline to detect network disconnection on write
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			
			// Measure latency by timing ping
			start := time.Now()
			if err := conn.WriteJSON(TunnelMessage{Type: MsgTypePing}); err != nil {
				tc.logger.Error().Err(err).Msg("Heartbeat failed - connection lost")
				tc.mu.Lock()
				tc.isConnected = false
				tc.mu.Unlock()
				// Trigger reconnection
				go tc.reconnect()
				return
			}
			
			// Clear write deadline after successful write
			conn.SetWriteDeadline(time.Time{})

			// Update latency (rough estimate - actual RTT would require pong response timing)
			// For now, use a simple measurement of ping send time
			latency := time.Since(start)
			tc.latencyMu.Lock()
			tc.latencyMs = latency.Milliseconds()
			tc.latencyMu.Unlock()

		case <-pongCheckTicker.C:
			// Check if we haven't received a pong in too long (connection might be dead)
			tc.pongMu.RLock()
			lastPong := tc.lastPongTime
			tc.pongMu.RUnlock()

			// If no pong received in 40 seconds (just over one heartbeat interval), consider connection dead
			// This detects network disconnection faster (within ~40 seconds instead of 90)
			// But only if we're actually connected - don't trigger reconnection if already disconnected
			timeSinceLastPong := time.Since(lastPong)
			if timeSinceLastPong > 40*time.Second {
				tc.mu.RLock()
				isConnected := tc.isConnected
				tc.mu.RUnlock()

				if isConnected {
					// Only mark as disconnected if we haven't received a pong in a long time
					// This means either the server isn't responding to pings, or the connection is dead
					tc.logger.Warn().
						Dur("time_since_last_pong", timeSinceLastPong).
						Msg("No pong received in 40 seconds - connection may be dead, will attempt reconnection")
					tc.mu.Lock()
					tc.isConnected = false
					tc.mu.Unlock()
					// Trigger reconnection
					go tc.reconnect()
					return
				}
			}
		}
	}
}

// reconnect attempts to reconnect to tunnel server
func (tc *TunnelClient) reconnect() {
	tc.reconnectMu.Lock()
	defer tc.reconnectMu.Unlock()

	// Check if already reconnecting
	tc.mu.RLock()
	isConnected := tc.isConnected
	isReconnecting := tc.isReconnecting
	tc.mu.RUnlock()

	if isConnected {
		return // Already connected
	}

	if isReconnecting {
		return // Already reconnecting in another goroutine
	}

	// Mark as reconnecting
	tc.mu.Lock()
	tc.isReconnecting = true
	tc.mu.Unlock()

	defer func() {
		tc.mu.Lock()
		tc.isReconnecting = false
		tc.mu.Unlock()
	}()

	backoff := 5 * time.Second
	maxBackoff := 60 * time.Second

	for {
		tc.logger.Debug().
			Dur("backoff", backoff).
			Str("subdomain", tc.subdomain).
			Msg("Attempting to reconnect...")

		time.Sleep(backoff)

		err := tc.Connect()
		if err == nil {
			tc.logger.Info().
				Str("subdomain", tc.subdomain).
				Str("public_url", tc.tunnel.PublicURL).
				Msg("Reconnected successfully - resumed same tunnel")
			return
		}

		tc.logger.Warn().
			Err(err).
			Dur("next_attempt", backoff).
			Msg("Reconnection failed, will retry")

		// Exponential backoff
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
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

// Close closes the tunnel connection gracefully
// This sends a close frame to the server before closing, allowing the server to detect the disconnection immediately
func (tc *TunnelClient) Close() error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.isConnected = false
	if tc.wsConn != nil {
		// Send close frame to notify server of graceful shutdown
		// This allows the server to immediately detect the disconnection
		closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Client shutting down")
		tc.wsConn.WriteMessage(websocket.CloseMessage, closeMsg)
		
		// Give server a moment to process the close message
		time.Sleep(100 * time.Millisecond)
		
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

// IsReconnecting returns whether the client is currently attempting to reconnect
func (tc *TunnelClient) IsReconnecting() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.isReconnecting
}

// GetConnectionStatus returns a string describing the connection status
func (tc *TunnelClient) GetConnectionStatus() string {
	tc.mu.RLock()
	isConnected := tc.isConnected
	isReconnecting := tc.isReconnecting
	tc.mu.RUnlock()

	if isConnected {
		return "online"
	} else if isReconnecting {
		return "reconnecting"
	} else {
		return "offline"
	}
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

// handleTCPData handles TCP/TLS data from tunnel server
func (tc *TunnelClient) handleTCPData(connectionID string, data []byte, isTLS bool) {
	// Check if connection exists
	tc.tcpConnMu.RLock()
	conn, exists := tc.tcpConnections[connectionID]
	tc.tcpConnMu.RUnlock()

	// If connection doesn't exist and this is the first message (empty data), create new connection
	if !exists {
		if len(data) == 0 {
			// New connection request - establish connection to local service
			if err := tc.establishTCPConnection(connectionID, isTLS); err != nil {
				tc.logger.Error().
					Err(err).
					Str("connection_id", connectionID).
					Msg("Failed to establish TCP connection")
				tc.sendTCPError(connectionID, "connection_failed", err.Error())
				return
			}
			// Get the connection we just created
			tc.tcpConnMu.RLock()
			conn = tc.tcpConnections[connectionID]
			tc.tcpConnMu.RUnlock()
		} else {
			tc.logger.Warn().
				Str("connection_id", connectionID).
				Msg("Received TCP data for non-existent connection")
			return
		}
	}

	if conn == nil {
		return
	}

	// If data is empty, this was just a connection establishment
	if len(data) == 0 {
		return
	}

	// Write data to local TCP connection
	_, err := conn.Write(data)
	if err != nil {
		tc.logger.Error().
			Err(err).
			Str("connection_id", connectionID).
			Msg("Failed to write TCP data to local connection")
		tc.sendTCPError(connectionID, "write_error", err.Error())
		tc.closeTCPConnection(connectionID)
	}
}

// establishTCPConnection establishes a TCP/TLS connection to local service
func (tc *TunnelClient) establishTCPConnection(connectionID string, useTLS bool) error {
	// Parse local URL (format: host:port for TCP/TLS)
	localAddr := tc.localURL
	if strings.HasPrefix(localAddr, "http://") {
		localAddr = strings.TrimPrefix(localAddr, "http://")
	} else if strings.HasPrefix(localAddr, "https://") {
		localAddr = strings.TrimPrefix(localAddr, "https://")
	}

	// Connect to local service
	var conn net.Conn
	var err error

	if useTLS {
		// TLS connection
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // Allow self-signed certificates
		}
		conn, err = tls.Dial("tcp", localAddr, tlsConfig)
	} else {
		// Plain TCP connection
		conn, err = net.Dial("tcp", localAddr)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to local service %s: %w", localAddr, err)
	}

	// Store connection
	tc.tcpConnMu.Lock()
	tc.tcpConnections[connectionID] = conn
	tc.tcpConnMu.Unlock()

	tc.logger.Info().
		Str("connection_id", connectionID).
		Str("local_addr", localAddr).
		Bool("tls", useTLS).
		Msg("Established TCP connection to local service")

	// Start reading from local connection and forwarding to tunnel
	go tc.forwardTCPToTunnel(connectionID, conn, useTLS)

	return nil
}

// forwardTCPToTunnel reads from local TCP connection and forwards to tunnel
func (tc *TunnelClient) forwardTCPToTunnel(connectionID string, conn net.Conn, isTLS bool) {
	defer tc.closeTCPConnection(connectionID)

	buffer := make([]byte, 4096)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				tc.logger.Debug().
					Err(err).
					Str("connection_id", connectionID).
					Msg("TCP connection read error")
			}
			// Send close message to tunnel
			tc.sendTCPError(connectionID, "connection_closed", "Local connection closed")
			return
		}

		if n > 0 {
			// Forward data to tunnel
			msgType := MsgTypeTCPData
			if isTLS {
				msgType = MsgTypeTLSData
			}

			msg := TunnelMessage{
				Type:      msgType,
				RequestID: connectionID,
				Data:      buffer[:n],
			}

			tc.mu.RLock()
			wsConn := tc.wsConn
			tc.mu.RUnlock()

			if wsConn == nil {
				tc.logger.Error().
					Str("connection_id", connectionID).
					Msg("WebSocket connection not available")
				return
			}

			if err := wsConn.WriteJSON(msg); err != nil {
				tc.logger.Error().
					Err(err).
					Str("connection_id", connectionID).
					Msg("Failed to forward TCP data")
				return
			}
		}
	}
}

// handleTCPError handles TCP/TLS errors from tunnel server
func (tc *TunnelClient) handleTCPError(connectionID string, err *HTTPError) {
	tc.logger.Error().
		Str("connection_id", connectionID).
		Str("error", err.Error).
		Str("message", err.Message).
		Msg("TCP connection error from tunnel")

	tc.closeTCPConnection(connectionID)
}

// sendTCPError sends a TCP/TLS error to tunnel server
func (tc *TunnelClient) sendTCPError(connectionID, errorType, message string) {
	msg := TunnelMessage{
		Type:      MsgTypeTCPError,
		RequestID: connectionID,
		Error: &HTTPError{
			RequestID: connectionID,
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
		tc.logger.Error().
			Err(err).
			Str("connection_id", connectionID).
			Msg("Failed to send TCP error")
	}
}

// closeTCPConnection closes and removes a TCP connection
func (tc *TunnelClient) closeTCPConnection(connectionID string) {
	tc.tcpConnMu.Lock()
	defer tc.tcpConnMu.Unlock()

	conn, exists := tc.tcpConnections[connectionID]
	if exists {
		if conn != nil {
			conn.Close()
		}
		delete(tc.tcpConnections, connectionID)
		tc.logger.Debug().
			Str("connection_id", connectionID).
			Msg("TCP connection closed")
	}
}

// handleUDPData handles UDP data from tunnel server
func (tc *TunnelClient) handleUDPData(connectionID string, data []byte) {
	// Ensure UDP connection is established
	tc.udpConnMu.Lock()
	if tc.udpConn == nil {
		// Parse local URL (format: host:port for UDP)
		localAddr := tc.localURL
		if strings.HasPrefix(localAddr, "udp://") {
			localAddr = strings.TrimPrefix(localAddr, "udp://")
		} else if strings.HasPrefix(localAddr, "http://") {
			localAddr = strings.TrimPrefix(localAddr, "http://")
		} else if strings.HasPrefix(localAddr, "https://") {
			localAddr = strings.TrimPrefix(localAddr, "https://")
		}

		// Resolve UDP address
		addr, err := net.ResolveUDPAddr("udp", localAddr)
		if err != nil {
			tc.logger.Error().
				Err(err).
				Str("local_addr", localAddr).
				Msg("Failed to resolve UDP address")
			tc.udpConnMu.Unlock()
			tc.sendUDPError(connectionID, "resolve_error", err.Error())
			return
		}

		// Create UDP connection (we'll use a single connection for all packets)
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			tc.logger.Error().
				Err(err).
				Str("local_addr", localAddr).
				Msg("Failed to create UDP connection")
			tc.udpConnMu.Unlock()
			tc.sendUDPError(connectionID, "connection_failed", err.Error())
			return
		}

		tc.udpConn = conn
		tc.logger.Info().
			Str("local_addr", localAddr).
			Msg("Established UDP connection to local service")

		// Start reading from local UDP connection and forwarding to tunnel
		go tc.forwardUDPToTunnel()
	}
	udpConn := tc.udpConn
	tc.udpConnMu.Unlock()

	// Write data to local UDP connection
	_, err := udpConn.Write(data)
	if err != nil {
		tc.logger.Error().
			Err(err).
			Str("connection_id", connectionID).
			Msg("Failed to write UDP data to local connection")
		tc.sendUDPError(connectionID, "write_error", err.Error())
	}
}

// forwardUDPToTunnel reads from local UDP connection and forwards to tunnel
func (tc *TunnelClient) forwardUDPToTunnel() {
	tc.udpConnMu.RLock()
	conn := tc.udpConn
	tc.udpConnMu.RUnlock()

	if conn == nil {
		return
	}

	buffer := make([]byte, 4096)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			tc.logger.Debug().
				Err(err).
				Msg("UDP connection read error")
			// UDP is connectionless, so we don't close on read error
			// Just log and continue
			continue
		}

		if n > 0 {
			// Generate connection ID for this packet
			connectionID := fmt.Sprintf("udp-%d", time.Now().UnixNano())

			// Forward data to tunnel
			msg := TunnelMessage{
				Type:      MsgTypeUDPData,
				RequestID: connectionID,
				Data:      buffer[:n],
			}

			tc.mu.RLock()
			wsConn := tc.wsConn
			tc.mu.RUnlock()

			if wsConn != nil {
				if err := wsConn.WriteJSON(msg); err != nil {
					tc.logger.Error().
						Err(err).
						Str("connection_id", connectionID).
						Msg("Failed to forward UDP data to tunnel")
				}
			}
		}
	}
}

// handleUDPError handles UDP errors from tunnel server
func (tc *TunnelClient) handleUDPError(connectionID string, err *HTTPError) {
	tc.logger.Error().
		Str("connection_id", connectionID).
		Str("error", err.Error).
		Str("message", err.Message).
		Msg("UDP connection error from tunnel")
}

// sendUDPError sends a UDP error to tunnel server
func (tc *TunnelClient) sendUDPError(connectionID, errorType, message string) {
	msg := TunnelMessage{
		Type:      MsgTypeUDPError,
		RequestID: connectionID,
		Error: &HTTPError{
			RequestID: connectionID,
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
		tc.logger.Error().
			Err(err).
			Str("connection_id", connectionID).
			Msg("Failed to send UDP error")
	}
}
