package tunnel

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type RequestEvent struct {
	Time       time.Time
	Method     string
	Path       string
	StatusCode int
	StatusText string
	LatencyMs  int64 // Response time in milliseconds
}

type RequestEventHandler func(event RequestEvent)

type ConnectionStatusChangeHandler func(status string)

type TunnelClient struct {
	serverURL             string
	localURL              string
	protocol              string // http, tcp, tls, udp
	host                  string // Optional: specific host/subdomain
	wsConn                *websocket.Conn
	tunnel                *TunnelInfo
	httpClient            *http.Client
	logger                zerolog.Logger
	mu                    sync.RWMutex
	reconnectMu           sync.Mutex
	isConnected           bool
	isReconnecting        bool // True when attempting to reconnect
	shouldExit            bool // True when tunnel was disconnected from dashboard - should exit instead of reconnect
	token                 string
	requestQueue          []*HTTPRequest // Queue for requests during disconnection
	queueMu               sync.Mutex
	tcpConnections        map[string]net.Conn // Track active TCP/TLS connections
	tcpConnMu             sync.RWMutex
	udpConn               *net.UDPConn // UDP connection to local service
	udpConnMu             sync.RWMutex
	subdomain             string                        // Saved subdomain for resuming
	tunnelID              string                        // Saved tunnel ID for resuming
	forceNew              bool                          // If true, force creating a new tunnel (don't resume or auto-find)
	persistence           *TunnelPersistence            // For saving/loading tunnel state
	latencyMs             int64                         // Current latency in milliseconds
	latencyMu             sync.RWMutex                  // Mutex for latency updates
	lastPongTime          time.Time                     // Last time we received a pong
	pongMu                sync.RWMutex                  // Mutex for lastPongTime
	requestHandler        RequestEventHandler           // Callback for request events
	requestHandlerMu      sync.RWMutex                  // Mutex for request handler
	statusChangeHandler   ConnectionStatusChangeHandler // Callback for connection status changes
	statusChangeHandlerMu sync.RWMutex                  // Mutex for status change handler
	writeMu               sync.Mutex                    // Mutex to serialize WebSocket writes (WebSocket is not thread-safe for concurrent writes)
	lastNotifiedStatus    string                        // Last status that was notified to prevent duplicate notifications
}

func NewTunnelClient(serverURL, localURL string, logger zerolog.Logger) *TunnelClient {
	return NewTunnelClientWithOptions(serverURL, localURL, "http", "", logger)
}

func NewTunnelClientWithOptions(serverURL, localURL, protocol, host string, logger zerolog.Logger) *TunnelClient {
	if protocol == "" {
		protocol = ProtocolHTTP
	}

	client := &TunnelClient{
		serverURL: serverURL,
		localURL:  localURL,
		protocol:  protocol,
		host:      host,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		logger:         logger,
		requestQueue:   make([]*HTTPRequest, 0),
		tcpConnections: make(map[string]net.Conn),
		persistence:    NewTunnelPersistence(logger),
	}

	if state, err := client.persistence.Load(); err == nil && state != nil {
		savedServer := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(state.ServerURL, "http://"), "https://"), "ws://"), "wss://")
		currentServer := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(serverURL, "http://"), "https://"), "ws://"), "wss://")
		protocolMatches := state.Protocol == "" || state.Protocol == protocol

		if savedServer == currentServer && protocolMatches {
			logger.Debug().
				Str("subdomain", state.Subdomain).
				Str("tunnel_id", state.TunnelID).
				Str("public_url", state.PublicURL).
				Msg("Found saved tunnel state (not using for resume - server will auto-find from DB)")
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

func (tc *TunnelClient) SetResumeInfo(subdomain, tunnelID string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.subdomain = subdomain
	tc.tunnelID = tunnelID
}

func (tc *TunnelClient) SetToken(token string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.token = token
}

func (tc *TunnelClient) SetRequestHandler(handler RequestEventHandler) {
	tc.requestHandlerMu.Lock()
	defer tc.requestHandlerMu.Unlock()
	tc.requestHandler = handler
}

func (tc *TunnelClient) SetConnectionStatusChangeHandler(handler ConnectionStatusChangeHandler) {
	tc.statusChangeHandlerMu.Lock()
	defer tc.statusChangeHandlerMu.Unlock()
	tc.statusChangeHandler = handler
}

func (tc *TunnelClient) notifyStatusChange() {
	tc.statusChangeHandlerMu.RLock()
	handler := tc.statusChangeHandler
	tc.statusChangeHandlerMu.RUnlock()

	if handler != nil {
		status := tc.GetConnectionStatus()

		tc.mu.RLock()
		lastNotifiedStatus := tc.lastNotifiedStatus
		tc.mu.RUnlock()

		if status != lastNotifiedStatus {
			tc.mu.Lock()
			tc.lastNotifiedStatus = status
			tc.mu.Unlock()

			go handler(status)
		}
	}
}

func (tc *TunnelClient) ClearResumeInfo() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.subdomain = ""
	tc.tunnelID = ""
	tc.forceNew = true // Set forceNew flag to prevent server from auto-finding tunnels
}

func (tc *TunnelClient) SetForceNew(forceNew bool) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.forceNew = forceNew
}

func (tc *TunnelClient) websocketURL() (scheme, host string) {
	u := strings.TrimSpace(tc.serverURL)
	if idx := strings.Index(u, "/"); idx >= 0 {
		u = u[:idx]
	}
	host = u
	for _, prefix := range []string{"wss://", "ws://", "https://", "http://"} {
		if strings.HasPrefix(strings.ToLower(u), prefix) {
			host = u[len(prefix):]
			if idx := strings.Index(host, "/"); idx >= 0 {
				host = host[:idx]
			}
			if prefix == "https://" || prefix == "wss://" {
				return "wss", host
			}
			return "ws", host
		}
	}
	if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
		return "ws", host
	}
	return "wss", host
}

func (tc *TunnelClient) Connect() error {
	tc.reconnectMu.Lock()
	defer tc.reconnectMu.Unlock()

	scheme, host := tc.websocketURL()
	wsURL := fmt.Sprintf("%s://%s/tunnel", scheme, host)
	tc.logger.Debug().Str("url", wsURL).Msg("Connecting to tunnel server")

	connectStart := time.Now()
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to tunnel server: %w", err)
	}
	tc.wsConn = conn

	connectLatency := time.Since(connectStart)
	tc.latencyMu.Lock()
	tc.latencyMs = connectLatency.Milliseconds()
	tc.latencyMu.Unlock()

	tc.mu.RLock()
	authToken := tc.token
	forceNew := tc.forceNew
	resumeSubdomain := tc.subdomain
	resumeTunnelID := tc.tunnelID
	tc.mu.RUnlock()

	initMsg := InitMessage{
		Type:      MsgTypeInit,
		Version:   "1.0",
		Protocol:  tc.protocol,
		LocalURL:  tc.localURL,
		Host:      tc.host,
		Token:     authToken,
		ForceNew:  forceNew,
		Subdomain: resumeSubdomain,
		TunnelID:  resumeTunnelID,
	}

	if forceNew {
		tc.logger.Debug().
			Str("protocol", tc.protocol).
			Str("local_url", tc.localURL).
			Str("host", tc.host).
			Msg("Force creating new tunnel (--new flag set)")
	} else if tc.host != "" {
		tc.logger.Debug().
			Str("protocol", tc.protocol).
			Str("local_url", tc.localURL).
			Str("host", tc.host).
			Msg("Creating new tunnel with specified host name")
	} else {
		tc.logger.Debug().
			Str("protocol", tc.protocol).
			Str("local_url", tc.localURL).
			Msg("No tunnel specified - server will auto-find from database or create new")
	}

	tc.writeMu.Lock()
	err = conn.WriteJSON(initMsg)
	tc.writeMu.Unlock()

	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to send init message: %w", err)
	}

	var responseData map[string]interface{}
	if err := conn.ReadJSON(&responseData); err != nil {
		conn.Close()
		return fmt.Errorf("failed to read response: %w", err)
	}

	if errorMsg, hasError := responseData["error"].(string); hasError {
		conn.Close()
		errorMessage, _ := responseData["message"].(string)
		if errorMessage == "" {
			errorMessage = errorMsg
		}
		
		if strings.Contains(strings.ToLower(errorMsg), "token") || 
		   strings.Contains(strings.ToLower(errorMsg), "authentication") ||
		   strings.Contains(strings.ToLower(errorMsg), "expired") ||
		   strings.Contains(strings.ToLower(errorMsg), "invalid") {
			tc.mu.Lock()
			tc.token = ""
			tc.mu.Unlock()

			tc.logger.Info().
				Str("error", errorMsg).
				Msg("Authentication failed - token cleared. Please run 'uniroute auth login' to authenticate again")
		}
		
		return fmt.Errorf("%s: %s", errorMsg, errorMessage)
	}

	var response InitResponse
	response.Type, _ = responseData["type"].(string)
	response.TunnelID, _ = responseData["tunnel_id"].(string)
	response.Subdomain, _ = responseData["subdomain"].(string)
	response.PublicURL, _ = responseData["public_url"].(string)
	response.Status, _ = responseData["status"].(string)

	tc.mu.Lock()
	tc.tunnel = &TunnelInfo{
		ID:        response.TunnelID,
		Subdomain: response.Subdomain,
		PublicURL: response.PublicURL,
		Status:    response.Status,
	}
	tc.subdomain = response.Subdomain
	tc.tunnelID = response.TunnelID
	tc.isConnected = true
	tc.isReconnecting = false  // Clear reconnecting flag on successful connection
	tc.lastNotifiedStatus = "" // Reset last notified status to force notification
	tc.mu.Unlock()

	tc.notifyStatusChange()

	tc.pongMu.Lock()
	tc.lastPongTime = time.Now()
	tc.pongMu.Unlock()

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

	go tc.processQueuedRequests()

	tc.logger.Debug().
		Str("tunnel_id", response.TunnelID).
		Str("subdomain", response.Subdomain).
		Str("public_url", response.PublicURL).
		Str("local_url", tc.localURL).
		Msg("Tunnel connected")

	go tc.handleMessages()

	go tc.heartbeat()

	return nil
}

func (tc *TunnelClient) handleMessages() {
	readDeadline := 90 * time.Second

	for {
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

			if websocket.IsCloseError(err, websocket.ClosePolicyViolation) {
				// Close code 1008 (Policy Violation) - tunnel was disconnected from dashboard
				tc.logger.Info().Err(err).Msg("Tunnel was disconnected from dashboard (close 1008) - exiting (will not reconnect)")
				tc.mu.Lock()
				tc.isConnected = false
				tc.isReconnecting = false
				tc.shouldExit = true // Signal that we should exit instead of reconnecting
				tc.mu.Unlock()
				tc.notifyStatusChange()
				return // Exit without reconnecting
			}

			if strings.Contains(errStr, "websocket: close 1008") || 
			   strings.Contains(errStr, "close 1008") ||
			   strings.Contains(errStr, "1008") ||
			   strings.Contains(errStr, "Policy Violation") ||
			   strings.Contains(errStr, "policy violation") ||
			   strings.Contains(errStr, "disconnected from dashboard") {
				tc.logger.Info().Err(err).Msg("Tunnel was disconnected from dashboard (detected via error string) - exiting (will not reconnect)")
				tc.mu.Lock()
				tc.isConnected = false
				tc.isReconnecting = false
				tc.shouldExit = true // Signal that we should exit instead of reconnecting
				tc.mu.Unlock()
				tc.notifyStatusChange()
				return // Exit without reconnecting
			}

			if websocket.IsCloseError(err, websocket.CloseNormalClosure) ||
			   strings.Contains(errStr, "websocket: close 1000") || 
			   strings.Contains(errStr, "close 1000 (normal)") {
				// Normal shutdown - just return without logging
				return
			}

			if strings.Contains(errStr, "repeated read on failed websocket connection") {
				tc.logger.Debug().Err(err).Msg("WebSocket connection failed - attempting to reconnect")
				tc.mu.Lock()
				tc.isConnected = false
				tc.mu.Unlock()
				// Notify status change immediately
				tc.notifyStatusChange()
				// Attempt to reconnect
				tc.reconnect()
				return
			}

			isConnectionError := false

			// WebSocket close codes indicate actual connection closure (but not normal 1000 or 1008)
			if strings.Contains(errStr, "websocket: close") {
				if !strings.Contains(errStr, "close 1000") && !strings.Contains(errStr, "close 1008") && !strings.Contains(errStr, "1008") {
					isConnectionError = true
				}
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
			if isConnectionError {
				tc.logger.Debug().Err(err).Msg("Connection lost - attempting to reconnect")
				tc.mu.Lock()
				tc.isConnected = false
				tc.mu.Unlock()
				// Notify status change immediately
				tc.notifyStatusChange()
				// Attempt to reconnect
				tc.reconnect()
				return
			} else {
				// Log other errors (like unexpected EOF without close code) at debug level
				// These might be temporary or related to local server issues, not network disconnection
				tc.logger.Debug().Err(err).Msg("Read error (not a connection failure) - breaking loop to prevent panic")
				// Don't continue - connection might be in bad state, break to prevent panic
				return
			}
		}

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
			if len(msg.Data) > 0 || msg.RequestID != "" {
				go tc.handleTCPData(msg.RequestID, msg.Data, msg.Type == MsgTypeTLSData)
			}
		case MsgTypeTCPError, MsgTypeTLSError:
			if msg.Error != nil {
				tc.handleTCPError(msg.RequestID, msg.Error)
			}
		case MsgTypeUDPData:
			if len(msg.Data) > 0 && msg.RequestID != "" {
				go tc.handleUDPData(msg.RequestID, msg.Data)
			}
		case MsgTypeUDPError:
			if msg.Error != nil {
				tc.handleUDPError(msg.RequestID, msg.Error)
			}
		case MsgTypePong:
			tc.pongMu.Lock()
			tc.lastPongTime = time.Now()
			tc.pongMu.Unlock()
			tc.logger.Debug().Msg("Received pong")
		case MsgTypeTunnelStatus:
			tc.logger.Debug().Msg("Received tunnel status update")
		default:
			tc.logger.Warn().Str("type", msg.Type).Msg("Unknown message type")
		}
	}
}

func (tc *TunnelClient) forwardToLocal(req *HTTPRequest) {
	if req == nil {
		tc.logger.Error().Msg("Received nil request")
		return
	}

	startTime := time.Now()

	path := req.Path
	if path == "" {
		path = "/"
	}

	baseURL, err := url.Parse(tc.localURL)
	if err != nil {
		tc.logger.Error().Err(err).Str("local_url", tc.localURL).Msg("Failed to parse base localURL")
		tc.sendError(req.RequestID, "invalid local URL", err.Error())
		return
	}

	fullURL := baseURL.ResolveReference(&url.URL{Path: path, RawQuery: req.Query})
	localURL := fullURL.String()

	tc.logger.Debug().
		Str("request_id", req.RequestID).
		Str("method", req.Method).
		Str("path", req.Path).
		Str("base_local_url", tc.localURL).
		Str("full_local_url", localURL).
		Str("url_host", baseURL.Host).
		Msg("Forwarding request to local server")

	httpReq, err := http.NewRequest(req.Method, localURL, bytes.NewReader(req.Body))
	if err != nil {
		tc.logger.Error().Err(err).Str("request_id", req.RequestID).Msg("Failed to create request")
		tc.sendError(req.RequestID, "failed to create request", err.Error())
		return
	}

	// Extract host:port from base URL (e.g., "localhost:3002")
	// Rewrites Host header to match upstream server
	localHost := baseURL.Host
	if localHost == "" {
		// Fallback: extract from URL string
		if strings.HasPrefix(tc.localURL, "http://") {
			localHost = strings.TrimPrefix(tc.localURL, "http://")
		} else if strings.HasPrefix(tc.localURL, "https://") {
			localHost = strings.TrimPrefix(tc.localURL, "https://")
		}
		// Remove path if present
		if idx := strings.Index(localHost, "/"); idx != -1 {
			localHost = localHost[:idx]
		}
	}

	// Copy headers, but rewrite Host header to match local server
	originalHost := ""
	for k, v := range req.Headers {
		if strings.EqualFold(k, "Host") {
			originalHost = v
		} else {
			if !strings.EqualFold(k, "Connection") &&
				!strings.EqualFold(k, "Content-Length") &&
				!strings.EqualFold(k, "Transfer-Encoding") {
				httpReq.Header.Set(k, v)
			}
		}
	}

	// Rewrite Host header to match local server
	if localHost != "" {
		httpReq.Header.Set("Host", localHost)
		httpReq.Host = localHost
		tc.logger.Debug().
			Str("rewritten_host", localHost).
			Str("original_host", originalHost).
			Str("base_local_url", tc.localURL).
			Str("full_request_url", localURL).
			Str("url_host", httpReq.URL.Host).
			Msg("Rewrote Host header for local server")
	} else {
		tc.logger.Error().
			Str("local_url", tc.localURL).
			Msg("Cannot rewrite Host header - empty host after parsing")
	}

	// Preserve original host in X-Forwarded-Host header for reference
	// Some servers use this for validation or logging
	if originalHost != "" {
		httpReq.Header.Set("X-Forwarded-Host", originalHost)
	}

	// Forward to local server
	// Log detailed request information for debugging
	tc.logger.Debug().
		Str("request_id", req.RequestID).
		Str("base_local_url", tc.localURL).
		Str("full_local_url", localURL).
		Str("host_header", httpReq.Header.Get("Host")).
		Str("request_host_field", httpReq.Host).
		Str("url_host", httpReq.URL.Host).
		Str("url_scheme", httpReq.URL.Scheme).
		Str("url_path", httpReq.URL.Path).
		Str("url_raw_query", httpReq.URL.RawQuery).
		Str("request_url", httpReq.URL.String()).
		Str("method", req.Method).
		Str("original_host", originalHost).
		Msg("Sending HTTP request to local server")

	// Make the request with a reasonable timeout
	resp, err := tc.httpClient.Do(httpReq)
	if err != nil {
		errStr := err.Error()
		tc.logger.Warn().
			Err(err).
			Str("request_id", req.RequestID).
			Str("local_url", localURL).
			Str("tunnel_local_url", tc.localURL).
			Str("method", req.Method).
			Str("path", req.Path).
			Str("host_header", httpReq.Header.Get("Host")).
			Str("request_host_field", httpReq.Host).
			Str("url_host", httpReq.URL.Host).
			Str("url_scheme", httpReq.URL.Scheme).
			Str("request_url", httpReq.URL.String()).
			Str("error_type", func() string {
				if strings.Contains(errStr, "connection refused") {
					return "connection_refused"
				} else if strings.Contains(errStr, "timeout") {
					return "timeout"
				} else if strings.Contains(errStr, "no such host") {
					return "no_such_host"
				}
				return "unknown"
			}()).
			Msg("Failed to forward request to local server - check if local server is running on correct port")

		// Notify request handler of error (502 Bad Gateway)
		latency := time.Since(startTime)
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
				LatencyMs:  latency.Milliseconds(),
			})
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

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := ""
		for k, v := range response.Headers {
			if strings.EqualFold(k, "Location") && v != "" {
				location = v
				break
			}
		}

		if location != "" {
			tc.logger.Debug().
				Str("redirect_location", location).
				Int("status_code", resp.StatusCode).
				Msg("Processing redirect Location header")

			tc.mu.RLock()
			tunnelURL := ""
			if tc.tunnel != nil && tc.tunnel.PublicURL != "" {
				tunnelURL = tc.tunnel.PublicURL
			}
			tc.mu.RUnlock()

			if tunnelURL != "" {
				locationURL, err := url.Parse(location)
				if err == nil {
					host := locationURL.Hostname()
					if host == "localhost" || host == "127.0.0.1" {
						tunnelURLParsed, err := url.Parse(tunnelURL)
						if err == nil {
							tunnelURLParsed.Path = locationURL.Path
							tunnelURLParsed.RawQuery = locationURL.RawQuery
							tunnelURLParsed.Fragment = locationURL.Fragment
							response.Headers["Location"] = tunnelURLParsed.String()

							tc.logger.Debug().
								Str("original_location", location).
								Str("rewritten_location", tunnelURLParsed.String()).
								Msg("Rewrote redirect Location header to use tunnel URL")
						}
					} else if host != "" {
						tc.logger.Debug().
							Str("external_redirect", location).
							Msg("Passing through external redirect unchanged")
					} else {
						// Relative redirect (like /path) - pass through unchanged
						tc.logger.Debug().
							Str("relative_redirect", location).
							Msg("Passing through relative redirect unchanged")
					}
				} else {
					tc.logger.Warn().
						Err(err).
						Str("location", location).
						Msg("Failed to parse Location URL, passing through unchanged")
				}
			}
		} else {
			tc.logger.Debug().
				Int("status_code", resp.StatusCode).
				Msg("Redirect response but no Location header found")
		}
	}

	// Send response back through tunnel
	tc.sendResponse(response)

	latency := time.Since(startTime)

	// Notify request handler if set (show all requests including static assets)
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
			LatencyMs:  latency.Milliseconds(),
		})
	}

	tc.logger.Debug().
		Str("request_id", req.RequestID).
		Int("status", resp.StatusCode).
		Int("body_size", len(body)).
		Msg("Request completed")
}

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

	// Serialize writes to prevent concurrent write panics
	tc.writeMu.Lock()
	defer tc.writeMu.Unlock()

	// Re-check connection after acquiring lock (it might have been closed)
	tc.mu.RLock()
	conn = tc.wsConn
	tc.mu.RUnlock()

	if conn == nil {
		tc.logger.Error().Str("request_id", resp.RequestID).Msg("Connection closed before write")
		return
	}

	if err := conn.WriteJSON(msg); err != nil {
		tc.logger.Error().Err(err).Str("request_id", resp.RequestID).Msg("Failed to send response")
	}
}

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

	// Serialize writes to prevent concurrent write panics
	tc.writeMu.Lock()
	defer tc.writeMu.Unlock()

	// Re-check connection after acquiring lock (it might have been closed)
	tc.mu.RLock()
	conn = tc.wsConn
	tc.mu.RUnlock()

	if conn == nil {
		tc.logger.Error().Str("request_id", requestID).Msg("Connection closed before write")
		return
	}

	if err := conn.WriteJSON(msg); err != nil {
		tc.logger.Error().Err(err).Str("request_id", requestID).Msg("Failed to send error")
	}
}

func (tc *TunnelClient) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

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

			start := time.Now()

			tc.writeMu.Lock()
			tc.mu.RLock()
			conn = tc.wsConn
			isConnected = tc.isConnected
			tc.mu.RUnlock()

			if !isConnected || conn == nil {
				tc.writeMu.Unlock()
				return
			}

			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := conn.WriteJSON(TunnelMessage{Type: MsgTypePing})
			tc.writeMu.Unlock()

			if err != nil {
				if !strings.Contains(err.Error(), "close sent") {
					tc.logger.Error().Err(err).Msg("Heartbeat failed - connection lost")
				}
				tc.mu.Lock()
				tc.isConnected = false
				tc.mu.Unlock()
				tc.notifyStatusChange()
				go tc.reconnect()
				return
			}

			conn.SetWriteDeadline(time.Time{})

			latency := time.Since(start)
			tc.latencyMu.Lock()
			tc.latencyMs = latency.Milliseconds()
			tc.latencyMu.Unlock()

		case <-pongCheckTicker.C:
			tc.mu.RLock()
			shouldExit := tc.shouldExit
			tc.mu.RUnlock()
			if shouldExit {
				tc.logger.Info().Msg("Tunnel was disconnected from dashboard - exiting heartbeat loop")
				return
			}

			tc.pongMu.RLock()
			lastPong := tc.lastPongTime
			tc.pongMu.RUnlock()

			timeSinceLastPong := time.Since(lastPong)
			if timeSinceLastPong > 150*time.Second {
				tc.mu.RLock()
				isConnected := tc.isConnected
				conn := tc.wsConn
				tc.mu.RUnlock()

				if isConnected && conn != nil {
					tc.logger.Warn().
						Dur("time_since_last_pong", timeSinceLastPong).
						Msg("No pong received in 150 seconds - connection may be dead, will attempt reconnection")
					tc.mu.Lock()
					tc.isConnected = false
					tc.mu.Unlock()
					tc.notifyStatusChange()
					go tc.reconnect()
					return
				}
			}
		}
	}
}

func (tc *TunnelClient) reconnect() {
	tc.reconnectMu.Lock()
	defer tc.reconnectMu.Unlock()

	tc.mu.RLock()
	isConnected := tc.isConnected
	isReconnecting := tc.isReconnecting
	shouldExit := tc.shouldExit
	tc.mu.RUnlock()

	// If tunnel was disconnected from dashboard, don't reconnect - exit instead
	if shouldExit {
		tc.logger.Info().Msg("Tunnel was disconnected from dashboard - not reconnecting")
		return
	}

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

		tc.mu.RLock()
		shouldExit := tc.shouldExit
		tc.mu.RUnlock()
		if shouldExit {
			tc.logger.Info().Msg("Tunnel was disconnected from dashboard - stopping reconnection attempts")
			return
		}

		err := tc.Connect()
		if err == nil {
			tc.logger.Info().
				Str("subdomain", tc.subdomain).
				Str("public_url", tc.tunnel.PublicURL).
				Msg("Reconnected successfully - resumed same tunnel")
			return
		}

		errStr := err.Error()
		if strings.Contains(strings.ToLower(errStr), "disconnected from dashboard") ||
		   strings.Contains(strings.ToLower(errStr), "tunnel disconnected") ||
		   strings.Contains(strings.ToLower(errStr), "inactive") {
			tc.logger.Info().Err(err).Msg("Tunnel was disconnected from dashboard - not reconnecting")
			tc.mu.Lock()
			tc.shouldExit = true
			tc.isReconnecting = false
			tc.mu.Unlock()
			tc.notifyStatusChange()
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

func (tc *TunnelClient) Close() error {
	tc.mu.Lock()

	// Mark as disconnected immediately to stop goroutines
	tc.isConnected = false
	tc.isReconnecting = false

	// Notify status change immediately
	tc.notifyStatusChange()

	wsConn := tc.wsConn
	tc.wsConn = nil // Clear connection reference
	tc.mu.Unlock()

	if wsConn != nil {
		// Try to send close frame with timeout to prevent hanging
		// Use a channel to make the write non-blocking
		done := make(chan struct{}, 1)
		go func() {
			defer func() {
				// Recover from any panic during close
				if r := recover(); r != nil {
					tc.logger.Debug().Interface("panic", r).Msg("Recovered from panic during close")
				}
				done <- struct{}{}
			}()

			// Set a short write deadline to prevent hanging
			wsConn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
			closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Client shutting down")
			wsConn.WriteMessage(websocket.CloseMessage, closeMsg)
		}()

		// Wait for write with timeout (very short - don't block shutdown)
		select {
		case <-done:
			// Write completed (success or error - doesn't matter)
		case <-time.After(1 * time.Second):
			// Timeout - just close the connection immediately
			tc.logger.Debug().Msg("Close message write timed out, closing connection immediately")
		}

		// Close the connection immediately regardless of write result
		// Don't wait - just close it
		wsConn.Close()
	}
	return nil
}

func (tc *TunnelClient) GetTunnelInfo() *TunnelInfo {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.tunnel
}

func (tc *TunnelClient) IsConnected() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.isConnected
}

func (tc *TunnelClient) IsReconnecting() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.isReconnecting
}

func (tc *TunnelClient) ShouldExit() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.shouldExit
}

func (tc *TunnelClient) GetConnectionStatus() string {
	tc.mu.RLock()
	isConnected := tc.isConnected
	isReconnecting := tc.isReconnecting
	wsConn := tc.wsConn
	tc.mu.RUnlock()

	// If we have a WebSocket connection and are marked as connected, we're online
	// This prevents false "reconnecting" status when connection is actually stable
	if isConnected && wsConn != nil {
		return "online"
	} else if isReconnecting {
		return "reconnecting"
	} else {
		return "offline"
	}
}

func (tc *TunnelClient) GetLatency() int64 {
	tc.latencyMu.RLock()
	defer tc.latencyMu.RUnlock()
	return tc.latencyMs
}

func (tc *TunnelClient) GetProtocol() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.protocol
}

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

func (tc *TunnelClient) handleTCPData(connectionID string, data []byte, isTLS bool) {
	// Check if connection exists
	tc.tcpConnMu.RLock()
	conn, exists := tc.tcpConnections[connectionID]
	tc.tcpConnMu.RUnlock()

	// If connection doesn't exist and this is the first message (empty data), create new connection
	if !exists {
		if len(data) == 0 {
			if err := tc.establishTCPConnection(connectionID, isTLS); err != nil {
				tc.logger.Error().
					Err(err).
					Str("connection_id", connectionID).
					Msg("Failed to establish TCP connection")
				tc.sendTCPError(connectionID, "connection_failed", err.Error())
				return
			}
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

			// Serialize writes to prevent concurrent write panics
			tc.writeMu.Lock()
			// Re-check connection after acquiring lock
			tc.mu.RLock()
			wsConn = tc.wsConn
			tc.mu.RUnlock()

			if wsConn == nil {
				tc.writeMu.Unlock()
				return
			}

			err := wsConn.WriteJSON(msg)
			tc.writeMu.Unlock()

			if err != nil {
				tc.logger.Error().
					Err(err).
					Str("connection_id", connectionID).
					Msg("Failed to forward TCP data")
				return
			}
		}
	}
}

func (tc *TunnelClient) handleTCPError(connectionID string, err *HTTPError) {
	tc.logger.Error().
		Str("connection_id", connectionID).
		Str("error", err.Error).
		Str("message", err.Message).
		Msg("TCP connection error from tunnel")

	tc.closeTCPConnection(connectionID)
}

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

	// Serialize writes to prevent concurrent write panics
	tc.writeMu.Lock()
	defer tc.writeMu.Unlock()

	// Re-check connection after acquiring lock (it might have been closed)
	tc.mu.RLock()
	conn = tc.wsConn
	tc.mu.RUnlock()

	if conn == nil {
		tc.logger.Error().Str("connection_id", connectionID).Msg("Connection closed before write")
		return
	}

	if err := conn.WriteJSON(msg); err != nil {
		tc.logger.Error().
			Err(err).
			Str("connection_id", connectionID).
			Msg("Failed to send TCP error")
	}
}

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
				// Serialize writes to prevent concurrent write panics
				tc.writeMu.Lock()
				// Re-check connection after acquiring lock
				tc.mu.RLock()
				wsConn = tc.wsConn
				tc.mu.RUnlock()

				if wsConn != nil {
					err := wsConn.WriteJSON(msg)
					tc.writeMu.Unlock()

					if err != nil {
						tc.logger.Error().
							Err(err).
							Str("connection_id", connectionID).
							Msg("Failed to forward UDP data to tunnel")
					}
				} else {
					tc.writeMu.Unlock()
				}
			}
		}
	}
}

func (tc *TunnelClient) handleUDPError(connectionID string, err *HTTPError) {
	tc.logger.Error().
		Str("connection_id", connectionID).
		Str("error", err.Error).
		Str("message", err.Message).
		Msg("UDP connection error from tunnel")
}

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

	// Serialize writes to prevent concurrent write panics
	tc.writeMu.Lock()
	defer tc.writeMu.Unlock()

	// Re-check connection after acquiring lock (it might have been closed)
	tc.mu.RLock()
	conn = tc.wsConn
	tc.mu.RUnlock()

	if conn == nil {
		tc.logger.Error().Str("connection_id", connectionID).Msg("Connection closed before write")
		return
	}

	if err := conn.WriteJSON(msg); err != nil {
		tc.logger.Error().
			Err(err).
			Str("connection_id", connectionID).
			Msg("Failed to send UDP error")
	}
}
