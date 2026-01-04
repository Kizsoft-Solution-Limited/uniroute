package tunnel

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

// TunnelServer manages tunnel connections
type TunnelServer struct {
	upgrader        websocket.Upgrader
	tunnels         map[string]*TunnelConnection
	tunnelsMu       sync.RWMutex
	httpServer      *http.Server
	port            int
	logger          zerolog.Logger
	subdomainPrefix string
	requestTracker  *RequestTracker
	tokenService    *TokenService
	repository      *TunnelRepository
	requestLogger   *RequestLogger
	rateLimiter     RateLimiterInterface
	statsCollector  *StatsCollector
	security        *SecurityMiddleware
	domainManager   *DomainManager
	requireAuth     bool
}

// TunnelConnection represents an active tunnel connection
type TunnelConnection struct {
	ID           string
	Subdomain    string
	LocalURL     string
	WSConn       *websocket.Conn
	CreatedAt    time.Time
	LastActive   time.Time
	RequestCount int64
	mu           sync.RWMutex
}

// TunnelConnection is used (Tunnel is in types.go for database model)

// NewTunnelServer creates a new tunnel server
func NewTunnelServer(port int, logger zerolog.Logger) *TunnelServer {
	return &TunnelServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now (TODO: add proper origin checking)
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		tunnels:         make(map[string]*TunnelConnection),
		port:            port,
		logger:          logger,
		subdomainPrefix: "tunnel",
		requestTracker:  NewRequestTracker(logger),
		tokenService:    NewTokenService(logger),
		rateLimiter:     NewTunnelRateLimiter(logger),
		statsCollector:  NewStatsCollector(logger),
		security:        NewSecurityMiddleware(logger),
		requireAuth:     false, // Phase 2: Can be enabled via config
	}
}

// SetRepository sets the tunnel repository
func (ts *TunnelServer) SetRepository(repo *TunnelRepository) {
	ts.repository = repo
	if repo != nil {
		ts.requestLogger = NewRequestLogger(repo, ts.logger)
	}
}

// SetRateLimiter sets the rate limiter (allows switching to Redis-based)
func (ts *TunnelServer) SetRateLimiter(limiter RateLimiterInterface) {
	ts.rateLimiter = limiter
}

// SetDomainManager sets the domain manager
func (ts *TunnelServer) SetDomainManager(manager *DomainManager) {
	ts.domainManager = manager
}

// SetRequireAuth enables/disables authentication requirement
func (ts *TunnelServer) SetRequireAuth(require bool) {
	ts.requireAuth = require
}

// Start starts the tunnel server
func (ts *TunnelServer) Start() error {
	mux := http.NewServeMux()

	// WebSocket endpoint for tunnel connections
	mux.HandleFunc("/tunnel", ts.handleTunnelConnection)

	// HTTP endpoint for forwarding requests
	mux.HandleFunc("/", ts.handleHTTPRequest)

	// Web interface endpoint
	mux.HandleFunc("/web", ts.handleWebInterface)
	
	// Phase 3: API endpoints
	mux.HandleFunc("/api/tunnels", ts.handleListTunnels)
	mux.HandleFunc("/api/tunnels/", ts.handleTunnelStats)

	ts.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", ts.port),
		Handler: mux,
	}

	ts.logger.Info().Int("port", ts.port).Msg("Tunnel server starting")
	return ts.httpServer.ListenAndServe()
}

// handleWebInterface handles web interface requests
func (ts *TunnelServer) handleWebInterface(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
		<html>
		<head><title>UniRoute Tunnel Server</title></head>
		<body>
			<h1>UniRoute Tunnel Server</h1>
			<p>Status: Running</p>
			<p>Active Tunnels: %d</p>
			<p>Web Interface (Phase 3)</p>
		</body>
		</html>
	`, len(ts.tunnels))
}

// Stop stops the tunnel server
func (ts *TunnelServer) Stop(ctx context.Context) error {
	return ts.httpServer.Shutdown(ctx)
}

// handleTunnelConnection handles new tunnel connections
func (ts *TunnelServer) handleTunnelConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := ts.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ts.logger.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}
	// Don't close here - let handleTunnelMessages manage the connection lifecycle

	// Read initial connection message
	var initMsg InitMessage
	if err := ws.ReadJSON(&initMsg); err != nil {
		ts.logger.Error().Err(err).Msg("Failed to read init message")
		return
	}

	// Validate init message
	if initMsg.LocalURL == "" {
		ts.logger.Error().Msg("Local URL is required")
		ws.WriteJSON(map[string]string{"error": "local_url is required"})
		return
	}

	// Check if client wants to resume existing tunnel
	var subdomain string
	var tunnelID string
	var isResume bool

	if initMsg.Subdomain != "" || initMsg.TunnelID != "" {
		// Try to resume existing tunnel
		ts.tunnelsMu.RLock()
		var existingTunnel *TunnelConnection
		
		if initMsg.Subdomain != "" {
			existingTunnel = ts.tunnels[initMsg.Subdomain]
		} else if initMsg.TunnelID != "" {
			// Find by tunnel ID
			for _, t := range ts.tunnels {
				if t.ID == initMsg.TunnelID {
					existingTunnel = t
					break
				}
			}
		}
		ts.tunnelsMu.RUnlock()

		if existingTunnel != nil {
			// Resume existing tunnel
			subdomain = existingTunnel.Subdomain
			tunnelID = existingTunnel.ID
			isResume = true
			
			// Update connection
			existingTunnel.mu.Lock()
			existingTunnel.WSConn = ws
			existingTunnel.LastActive = time.Now()
			existingTunnel.mu.Unlock()
			
			ts.logger.Info().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Msg("Resuming existing tunnel")
		} else {
			// Tunnel not found, create new one
			ts.logger.Warn().
				Str("requested_subdomain", initMsg.Subdomain).
				Str("requested_tunnel_id", initMsg.TunnelID).
				Msg("Requested tunnel not found, creating new tunnel")
		}
	}

	// Create new tunnel if not resuming
	if !isResume {
		// Phase 5: Allocate subdomain using domain manager
		var err error
		if ts.domainManager != nil {
			subdomain, err = ts.domainManager.AllocateSubdomain(context.Background(), ts.repository)
			if err != nil {
				ts.logger.Error().Err(err).Msg("Failed to allocate subdomain")
				ws.WriteJSON(map[string]string{"error": "failed to allocate subdomain"})
				return
			}
		} else {
			// Fallback to old method
			subdomain = ts.generateSubdomain()
		}
		tunnelID = generateID()
	}

	// Create or update tunnel connection
	var tunnel *TunnelConnection
	if isResume {
		// Get existing tunnel (already updated WSConn above)
		ts.tunnelsMu.RLock()
		tunnel = ts.tunnels[subdomain]
		ts.tunnelsMu.RUnlock()
		
		// Update local URL in case it changed
		tunnel.mu.Lock()
		tunnel.LocalURL = initMsg.LocalURL
		tunnel.mu.Unlock()
	} else {
		// Create new tunnel connection
		tunnel = &TunnelConnection{
			ID:           tunnelID,
			Subdomain:    subdomain,
			LocalURL:     initMsg.LocalURL,
			WSConn:       ws,
			CreatedAt:    time.Now(),
			LastActive:   time.Now(),
			RequestCount: 0,
		}

		// Register tunnel
		ts.tunnelsMu.Lock()
		ts.tunnels[subdomain] = tunnel
		ts.tunnelsMu.Unlock()
	}

	if isResume {
		ts.logger.Info().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Str("local_url", initMsg.LocalURL).
			Msg("Tunnel resumed")
	} else {
		ts.logger.Info().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Str("local_url", initMsg.LocalURL).
			Msg("Tunnel created")
	}

	// Phase 5: Generate public URL using domain manager
	// Note: SSL/HTTPS is handled by Coolify or reverse proxy, so we use HTTPS when base domain is set
	publicURL := fmt.Sprintf("http://%s.localhost:%d", subdomain, ts.port)
	if ts.domainManager != nil {
		// Use HTTPS if base domain is set (Coolify will handle SSL termination)
		useHTTPS := ts.domainManager.baseDomain != ""
		publicURL = ts.domainManager.GetPublicURL(subdomain, ts.port, useHTTPS)
	}
	
	// Send confirmation
	response := InitResponse{
		Type:      MsgTypeTunnelCreated,
		TunnelID:  tunnelID,
		Subdomain: subdomain,
		PublicURL: publicURL,
		Status:    "active",
	}
	if err := ws.WriteJSON(response); err != nil {
		ts.logger.Error().Err(err).Msg("Failed to send tunnel creation response")
		return
	}

	// Handle tunnel messages
	go ts.handleTunnelMessages(tunnel)
}

// handleTunnelMessages handles messages from tunnel client
func (ts *TunnelServer) handleTunnelMessages(tunnel *TunnelConnection) {
	defer tunnel.WSConn.Close() // Close connection when handler exits
	
	for {
		var msg TunnelMessage
		if err := tunnel.WSConn.ReadJSON(&msg); err != nil {
			ts.logger.Info().
				Str("tunnel_id", tunnel.ID).
				Str("subdomain", tunnel.Subdomain).
				Err(err).
				Msg("Tunnel disconnected")
			ts.removeTunnel(tunnel.Subdomain)
			return
		}

		tunnel.mu.Lock()
		tunnel.LastActive = time.Now()
		tunnel.mu.Unlock()

		// Handle different message types
		switch msg.Type {
		case MsgTypePing:
			tunnel.WSConn.WriteJSON(TunnelMessage{Type: MsgTypePong})
		case MsgTypeHTTPResponse:
			// Complete pending request with response
			if msg.Response != nil {
				if err := ts.requestTracker.CompleteRequest(msg.RequestID, msg.Response); err != nil {
					ts.logger.Error().
						Err(err).
						Str("request_id", msg.RequestID).
						Msg("Failed to complete request")
				} else {
					ts.logger.Debug().
						Str("request_id", msg.RequestID).
						Msg("Completed HTTP response from tunnel")
				}
			}
		case MsgTypeHTTPError:
			// Fail pending request with error
			if msg.Error != nil {
				err := fmt.Errorf("%s: %s", msg.Error.Error, msg.Error.Message)
				if err2 := ts.requestTracker.FailRequest(msg.RequestID, err); err2 != nil {
					ts.logger.Error().
						Err(err2).
						Str("request_id", msg.RequestID).
						Msg("Failed to fail request")
				} else {
					ts.logger.Debug().
						Str("request_id", msg.RequestID).
						Str("error", msg.Error.Error).
						Msg("Failed HTTP request from tunnel")
				}
			}
		default:
			ts.logger.Warn().
				Str("type", msg.Type).
				Msg("Unknown message type")
		}
	}
}

// handleHTTPRequest handles incoming HTTP requests and forwards to tunnel
func (ts *TunnelServer) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	// Phase 4: Add security headers
	ts.security.AddSecurityHeaders(w, r)
	if r.Method == http.MethodOptions {
		return // Preflight handled
	}

	// Phase 4: Validate request
	if err := ts.security.ValidateRequest(r); err != nil {
		ts.logger.Warn().Err(err).Str("method", r.Method).Str("path", r.URL.Path).Msg("Invalid request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Phase 4: Sanitize path
	r.URL.Path = ts.security.SanitizePath(r.URL.Path)

	// Extract subdomain from Host header
	host := r.Host
	subdomain := extractSubdomain(host)

	if subdomain == "" {
		// No subdomain, return info page or 404
		ts.handleRootRequest(w, r)
		return
	}

	ts.tunnelsMu.RLock()
	tunnel, exists := ts.tunnels[subdomain]
	ts.tunnelsMu.RUnlock()

	if !exists {
		http.NotFound(w, r)
		return
	}

	// Forward request to tunnel client
	ts.forwardHTTPRequest(tunnel, w, r)
}

// handleRootRequest handles requests to root domain
func (ts *TunnelServer) handleRootRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
		<html>
		<head><title>UniRoute Tunnel Server</title></head>
		<body>
			<h1>UniRoute Tunnel Server</h1>
			<p>Status: Running</p>
			<p>Active Tunnels: %d</p>
		</body>
		</html>
	`, len(ts.tunnels))
}

// handleHealth handles health check requests
func (ts *TunnelServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","tunnels":%d}`, len(ts.tunnels))
}

// forwardHTTPRequest forwards HTTP request through tunnel
func (ts *TunnelServer) forwardHTTPRequest(tunnel *TunnelConnection, w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := generateID()
	
	// Phase 3: Check rate limit
	allowed, err := ts.rateLimiter.CheckRateLimit(r.Context(), tunnel.ID)
	if err != nil {
		ts.logger.Error().Err(err).Str("tunnel_id", tunnel.ID).Msg("Rate limit check failed")
		// Allow request if rate limit check fails (fail open)
	} else if !allowed {
		ts.logger.Warn().Str("tunnel_id", tunnel.ID).Msg("Rate limit exceeded")
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	
	// Register pending request for tracking
	pendingReq, err := ts.requestTracker.RegisterRequest(requestID)
	if err != nil {
		ts.logger.Error().Err(err).Str("request_id", requestID).Msg("Failed to register request")
		http.Error(w, "Failed to register request", http.StatusInternalServerError)
		return
	}
	
	// Serialize request
	reqData, err := ts.serializeRequest(r, requestID)
	if err != nil {
		ts.requestTracker.FailRequest(requestID, err)
		http.Error(w, "Failed to serialize request", http.StatusInternalServerError)
		return
	}
	
	// Record request for rate limiting
	ts.rateLimiter.RecordRequest(r.Context(), tunnel.ID)

	// Send request through WebSocket
	msg := TunnelMessage{
		Type:      MsgTypeHTTPRequest,
		RequestID: requestID,
		Request:   reqData,
	}

	if err := tunnel.WSConn.WriteJSON(msg); err != nil {
		ts.requestTracker.FailRequest(requestID, err)
		http.Error(w, "Tunnel connection error", http.StatusBadGateway)
		return
	}

	// Wait for response using request tracker
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := pendingReq.WaitForResponse(ctx)
	if err != nil {
		ts.logger.Error().Err(err).Str("request_id", requestID).Msg("Failed to receive response")
		http.Error(w, "Request failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Write response
	ts.writeResponse(w, response)

	latency := time.Since(start)
	tunnel.mu.Lock()
	tunnel.RequestCount++
	count := tunnel.RequestCount
	tunnel.mu.Unlock()

	// Phase 3: Log request to database
	if ts.requestLogger != nil {
		reqLog := &TunnelRequestLog{
			TunnelID:     tunnel.ID,
			RequestID:    requestID,
			Method:       r.Method,
			Path:         r.URL.Path,
			StatusCode:   response.Status,
			LatencyMs:    int(latency.Milliseconds()),
			RequestSize:  len(reqData.Body),
			ResponseSize: len(response.Body),
			CreatedAt:    time.Now(),
		}
		ts.requestLogger.LogRequest(r.Context(), reqLog)
	}

	// Phase 3: Record statistics
	isError := response.Status >= 400
	ts.statsCollector.RecordRequest(tunnel.ID, int(latency.Milliseconds()), len(reqData.Body), len(response.Body), isError)

	// Update database if repository is available
	if ts.repository != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := ts.repository.UpdateTunnelActivity(ctx, tunnel.ID, count); err != nil {
				ts.logger.Error().Err(err).Str("tunnel_id", tunnel.ID).Msg("Failed to update tunnel activity")
			}
		}()
	}

	ts.logger.Debug().
		Str("request_id", requestID).
		Str("tunnel_id", tunnel.ID).
		Dur("latency", latency).
		Msg("Request forwarded")
}

// serializeRequest serializes HTTP request to TunnelMessage format
func (ts *TunnelServer) serializeRequest(r *http.Request, requestID string) (*HTTPRequest, error) {
	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	r.Body.Close()

	// Copy headers
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &HTTPRequest{
		RequestID: requestID,
		Method:    r.Method,
		Path:      r.URL.Path,
		Query:     r.URL.RawQuery,
		Headers:   headers,
		Body:      body,
	}, nil
}

// writeResponse writes HTTP response from tunnel
func (ts *TunnelServer) writeResponse(w http.ResponseWriter, resp *HTTPResponse) {
	// Write status code
	w.WriteHeader(resp.Status)

	// Write headers
	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}

	// Write body
	if len(resp.Body) > 0 {
		w.Write(resp.Body)
	}
}

// generateSubdomain generates a random subdomain
func (ts *TunnelServer) generateSubdomain() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:12] // 12 character subdomain
}

// removeTunnel removes a tunnel
func (ts *TunnelServer) removeTunnel(subdomain string) {
	ts.tunnelsMu.Lock()
	defer ts.tunnelsMu.Unlock()
	delete(ts.tunnels, subdomain)
	ts.logger.Info().Str("subdomain", subdomain).Msg("Tunnel removed")
}

// GetTunnel returns a tunnel by subdomain
func (ts *TunnelServer) GetTunnel(subdomain string) (*TunnelConnection, bool) {
	ts.tunnelsMu.RLock()
	defer ts.tunnelsMu.RUnlock()
	tunnel, exists := ts.tunnels[subdomain]
	return tunnel, exists
}

// ListTunnels returns all active tunnels
func (ts *TunnelServer) ListTunnels() []*TunnelConnection {
	ts.tunnelsMu.RLock()
	defer ts.tunnelsMu.RUnlock()

	tunnels := make([]*TunnelConnection, 0, len(ts.tunnels))
	for _, tunnel := range ts.tunnels {
		tunnels = append(tunnels, tunnel)
	}
	return tunnels
}

// Helper functions
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func extractSubdomain(host string) string {
	// Remove port if present
	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		hostname = host
	}

	// Split by dots
	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return ""
	}

	// For localhost:port format like "abc123.localhost:8080"
	if parts[1] == "localhost" {
		return parts[0]
	}

	// For domain format like "abc123.uniroute.dev"
	// Return first part as subdomain
	return parts[0]
}

// Types are defined in types.go
