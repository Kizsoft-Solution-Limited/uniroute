package tunnel

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// Note: uuid is already imported above

// TunnelServer manages tunnel connections
type TunnelServer struct {
	upgrader        websocket.Upgrader
	tunnels         map[string]*TunnelConnection
	tunnelsMu       sync.RWMutex
	tcpConnections  map[string]*TCPConnection // Track active TCP/TLS connections
	tcpConnMu       sync.RWMutex
	portMap         map[int]*TunnelConnection // Map TCP port -> tunnel (for TCP/TLS tunnels)
	portMapMu       sync.RWMutex
	nextTCPPort     int          // Next available TCP port for allocation
	tcpListener     net.Listener // TCP listener for accepting connections
	tcpListenerMu   sync.RWMutex
	httpServer      *http.Server
	port            int
	tcpPortBase     int // Base port for TCP tunnel allocation (default: 20000)
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
	jwtValidator    func(tokenString string) (userID string, err error) // JWT validator function (optional)
}

// TCPConnection represents an active TCP/TLS connection
type TCPConnection struct {
	ID        string
	TunnelID  string
	Conn      net.Conn
	CreatedAt time.Time
	mu        sync.RWMutex
}

// TunnelConnection represents an active tunnel connection
type TunnelConnection struct {
	ID           string
	Subdomain    string
	LocalURL     string
	Protocol     string // http, tcp, tls
	Host         string // Optional requested host
	WSConn       *websocket.Conn
	CreatedAt    time.Time
	LastActive   time.Time
	RequestCount int64
	handlerReady bool // Flag to indicate handler goroutine is ready
	mu           sync.RWMutex
}

// TunnelConnection is used (Tunnel is in types.go for database model)

// NewTunnelServer creates a new tunnel server
// allowedOrigins: List of allowed origin patterns (empty = use defaults)
func NewTunnelServer(port int, logger zerolog.Logger, allowedOrigins []string) *TunnelServer {
	// Default allowed origins if not provided
	defaultOrigins := []string{
		"http://localhost",
		"https://localhost",
		"http://127.0.0.1",
		"https://127.0.0.1",
		"tunnel.uniroute.co",
		".uniroute.co", // Allow subdomains
	}

	// Use provided origins if available, otherwise use defaults
	originPatterns := defaultOrigins
	if len(allowedOrigins) > 0 {
		originPatterns = allowedOrigins
	}

	return &TunnelServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Validate origin for security
				origin := r.Header.Get("Origin")
				if origin == "" {
					// No origin header - allow for direct connections (CLI, etc.)
					return true
				}
				// For tunnel connections, we validate against allowed origins
				for _, pattern := range originPatterns {
					if strings.Contains(origin, pattern) {
						return true
					}
				}
				// Log suspicious origin attempts
				logger.Warn().Str("origin", origin).Msg("Rejected WebSocket connection from unauthorized origin")
				return false
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		tunnels:         make(map[string]*TunnelConnection),
		tcpConnections:  make(map[string]*TCPConnection),
		portMap:         make(map[int]*TunnelConnection),
		nextTCPPort:     20000, // Start TCP port allocation from 20000
		tcpPortBase:     20000, // Base port for TCP tunnels
		port:            port,
		logger:          logger,
		subdomainPrefix: "tunnel",
		requestTracker:  NewRequestTracker(logger),
		tokenService:    NewTokenService(logger),
		rateLimiter:     NewTunnelRateLimiter(logger),
		statsCollector:  NewStatsCollector(logger),
		security:        NewSecurityMiddleware(logger),
		requireAuth:     false, // Can be enabled via config
	}
}

// SetTCPPortBase sets the base port for TCP tunnel allocation
func (ts *TunnelServer) SetTCPPortBase(basePort int) {
	ts.tcpPortBase = basePort
	ts.nextTCPPort = basePort
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

// SetJWTValidator sets a JWT validator function for extracting user ID from auth tokens
// This allows tunnels created via authenticated CLI to be automatically associated with users
func (ts *TunnelServer) SetJWTValidator(validator func(tokenString string) (userID string, err error)) {
	ts.jwtValidator = validator
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

	// API endpoints for tunnel management
	mux.HandleFunc("/api/tunnels", ts.handleListTunnels)
	mux.HandleFunc("/api/tunnels/", func(w http.ResponseWriter, r *http.Request) {
		// Route based on path pattern
		path := r.URL.Path
		if strings.HasSuffix(path, "/stats") {
			ts.handleTunnelStats(w, r)
		} else if strings.Contains(path, "/requests/") && strings.HasSuffix(path, "/replay") {
			ts.handleReplayTunnelRequest(w, r)
		} else if strings.Contains(path, "/requests/") {
			ts.handleGetTunnelRequest(w, r)
		} else if strings.HasSuffix(path, "/requests") {
			ts.handleListTunnelRequests(w, r)
		} else {
			ts.handleTunnelStats(w, r)
		}
	})

	ts.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", ts.port),
		Handler: mux,
	}

	ts.logger.Info().Int("port", ts.port).Msg("Tunnel server starting")

	// TCP listeners are started per-tunnel when TCP/TLS tunnels are created
	// See allocateTCPPort() which starts a listener for each allocated port

	// Start HTTP server (handles WebSocket and HTTP tunnels)
	// This blocks, so TCP listeners run in background
	return ts.httpServer.ListenAndServe()
}

// handleWebInterface handles web interface requests
func (ts *TunnelServer) handleWebInterface(w http.ResponseWriter, r *http.Request) {
	// Same as root request - redirect to root or show same page
	ts.handleRootRequest(w, r)
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
		ws.Close()
		return
	}

	// Validate and sanitize local URL
	if err := validateLocalURL(initMsg.LocalURL); err != nil {
		ts.logger.Error().Err(err).Str("local_url", initMsg.LocalURL).Msg("Invalid local URL")
		ws.WriteJSON(map[string]string{"error": "invalid local_url: " + err.Error()})
		ws.Close()
		return
	}

	// Validate subdomain if provided
	if initMsg.Subdomain != "" {
		if err := validateSubdomain(initMsg.Subdomain); err != nil {
			ts.logger.Error().Err(err).Str("subdomain", initMsg.Subdomain).Msg("Invalid subdomain")
			ws.WriteJSON(map[string]string{"error": "invalid subdomain: " + err.Error()})
			ws.Close()
			return
		}
	}

	// Default to HTTP if protocol not specified
	protocol := initMsg.Protocol
	if protocol == "" {
		protocol = ProtocolHTTP
	}

	// Validate protocol
	if protocol != ProtocolHTTP && protocol != ProtocolTCP && protocol != ProtocolTLS {
		ts.logger.Error().Str("protocol", protocol).Msg("Invalid protocol")
		ws.WriteJSON(map[string]string{"error": "invalid protocol, must be http, tcp, or tls"})
		ws.Close()
		return
	}

	// Check if client wants to resume existing tunnel
	var subdomain string
	var tunnelID string
	var isResume bool

	ts.logger.Info().
		Str("init_subdomain", initMsg.Subdomain).
		Str("init_tunnel_id", initMsg.TunnelID).
		Str("init_protocol", initMsg.Protocol).
		Msg("Received tunnel connection request - checking for resume")

	if initMsg.Subdomain != "" || initMsg.TunnelID != "" {
		// Try to resume existing tunnel
		// First check in-memory tunnels (active connections)
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
			// Resume existing in-memory tunnel
			// CRITICAL FIX: Don't update the existing tunnel's WebSocket here
			// Instead, let the unified resume path (below) create a fresh tunnel
			// This ensures consistent behavior and prevents race conditions
			subdomain = existingTunnel.Subdomain
			tunnelID = existingTunnel.ID
			isResume = true

			ts.logger.Info().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Str("existing_tunnel_id", existingTunnel.ID).
				Str("existing_local_url", existingTunnel.LocalURL).
				Msg("Resuming existing tunnel (from memory) - will create fresh tunnel connection")
		} else if ts.repository != nil {
			// Not in memory, check database for persisted tunnel
			var dbTunnel *Tunnel
			var err error

			ts.logger.Debug().
				Str("requested_subdomain", initMsg.Subdomain).
				Str("requested_tunnel_id", initMsg.TunnelID).
				Msg("Tunnel not in memory, checking database for resume")

			if initMsg.Subdomain != "" {
				ts.logger.Debug().
					Str("subdomain", initMsg.Subdomain).
					Msg("Looking up tunnel by subdomain in database")
				dbTunnel, err = ts.repository.GetTunnelBySubdomain(context.Background(), initMsg.Subdomain)
				if err != nil {
					ts.logger.Debug().
						Err(err).
						Str("subdomain", initMsg.Subdomain).
						Msg("Tunnel not found by subdomain in database")
				} else if dbTunnel != nil {
					ts.logger.Debug().
						Str("tunnel_id", dbTunnel.ID).
						Str("subdomain", dbTunnel.Subdomain).
						Str("status", dbTunnel.Status).
						Msg("Found tunnel in database by subdomain")
				}
			} else if initMsg.TunnelID != "" {
				// Parse tunnel ID string to UUID
				tunnelUUID, parseErr := uuid.Parse(initMsg.TunnelID)
				if parseErr == nil {
					ts.logger.Debug().
						Str("tunnel_id", initMsg.TunnelID).
						Msg("Looking up tunnel by ID in database")
					dbTunnel, err = ts.repository.GetTunnelByID(context.Background(), tunnelUUID)
					if err != nil {
						ts.logger.Debug().
							Err(err).
							Str("tunnel_id", initMsg.TunnelID).
							Msg("Tunnel not found by ID in database")
					} else if dbTunnel != nil {
						ts.logger.Debug().
							Str("tunnel_id", dbTunnel.ID).
							Str("subdomain", dbTunnel.Subdomain).
							Str("status", dbTunnel.Status).
							Msg("Found tunnel in database by ID")
					}
				} else {
					err = parseErr
					ts.logger.Debug().
						Err(parseErr).
						Str("tunnel_id", initMsg.TunnelID).
						Msg("Failed to parse tunnel ID as UUID")
				}
			}

			if err == nil && dbTunnel != nil {
				// Found tunnel in database, resume it (regardless of status - inactive tunnels can be resumed)
				// Use the subdomain from database (this is the authoritative source)
				// But verify it matches what client requested (for security/validation)
				if initMsg.Subdomain != "" && dbTunnel.Subdomain != initMsg.Subdomain {
					ts.logger.Warn().
						Str("requested_subdomain", initMsg.Subdomain).
						Str("database_subdomain", dbTunnel.Subdomain).
						Msg("Subdomain mismatch - using database subdomain (authoritative)")
				}
				subdomain = dbTunnel.Subdomain
				tunnelID = dbTunnel.ID
				isResume = true

				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("subdomain", subdomain).
					Str("previous_status", dbTunnel.Status).
					Str("requested_subdomain", initMsg.Subdomain).
					Str("requested_tunnel_id", initMsg.TunnelID).
					Str("database_subdomain", dbTunnel.Subdomain).
					Msg("Resuming existing tunnel (from database) - will mark as active")
			} else {
				// Tunnel not found in memory or database, create new one
				if err != nil {
					ts.logger.Warn().
						Err(err).
						Str("requested_subdomain", initMsg.Subdomain).
						Str("requested_tunnel_id", initMsg.TunnelID).
						Msg("Tunnel not found in database - will create new tunnel")
				} else {
					ts.logger.Info().
						Str("requested_subdomain", initMsg.Subdomain).
						Str("requested_tunnel_id", initMsg.TunnelID).
						Msg("Requested tunnel not found in memory or database, creating new tunnel")
				}
			}
		} else {
			// No database, tunnel not in memory, create new one
			ts.logger.Info().
				Str("requested_subdomain", initMsg.Subdomain).
				Str("requested_tunnel_id", initMsg.TunnelID).
				Msg("Requested tunnel not found in memory (no database), creating new tunnel")
		}
	}

	// Create new tunnel if not resuming
	if !isResume {
		// Allocate subdomain using domain manager
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
		// Always generate a new UUID for new tunnels
		tunnelID = generateID()
	} else {
		// For resume, validate that tunnelID is a valid UUID
		// If client sent an old hex-format ID, we need to use the one from database
		if tunnelID != "" {
			if _, err := uuid.Parse(tunnelID); err != nil {
				ts.logger.Warn().
					Err(err).
					Str("invalid_tunnel_id", tunnelID).
					Msg("Resume tunnel ID is not a valid UUID - using database ID instead")
				// If we found the tunnel in database, use its ID
				if ts.repository != nil {
					var dbTunnel *Tunnel
					if subdomain != "" {
						dbTunnel, err = ts.repository.GetTunnelBySubdomain(context.Background(), subdomain)
					}
					if err == nil && dbTunnel != nil {
						tunnelID = dbTunnel.ID
					} else {
						// Can't resume with invalid ID, create new tunnel
						ts.logger.Warn().Msg("Cannot resume tunnel with invalid ID - creating new tunnel")
						isResume = false
						tunnelID = generateID()
					}
				} else {
					// No database, can't validate - create new tunnel
					ts.logger.Warn().Msg("Cannot validate tunnel ID without database - creating new tunnel")
					isResume = false
					tunnelID = generateID()
				}
			}
		}
	}

	// Create or update tunnel connection
	var tunnel *TunnelConnection
	// Ensure subdomain is set (critical for registration)
	if subdomain == "" {
		ts.logger.Error().Msg("CRITICAL: subdomain is empty - cannot create tunnel")
		ws.WriteJSON(map[string]string{"error": "subdomain is required"})
		return
	}
	
	if isResume {
		// SIMPLIFIED: Always create a fresh tunnel connection for resume, just like new tunnel
		// This ensures resume works exactly the same way as new tunnel creation
		// CRITICAL: Create the new tunnel FIRST, then atomically replace the old one
		// This minimizes the window where no tunnel is registered (prevents 404 errors)
		
		// Double-check subdomain is not empty before proceeding
		if subdomain == "" {
			ts.logger.Error().
				Str("tunnel_id", tunnelID).
				Str("init_subdomain", initMsg.Subdomain).
				Msg("CRITICAL: Cannot resume tunnel - subdomain is empty")
			ws.WriteJSON(map[string]string{"error": "subdomain is required for resume"})
			return
		}
		
		// Create fresh tunnel connection
		// CRITICAL: Always use LocalURL from client's init message, not from database
		// This ensures the tunnel forwards to the port the user specified
		tunnel = &TunnelConnection{
			ID:           tunnelID,
			Subdomain:    subdomain,
			LocalURL:     initMsg.LocalURL,
			Protocol:     protocol,
			Host:         initMsg.Host,
			WSConn:       ws,
			CreatedAt:    time.Now(),
			LastActive:   time.Now(),
			RequestCount: 0,
			handlerReady: false, // Will be set to true when handler starts
		}

		// CRITICAL: Atomically replace old tunnel with new one
		// This ensures there's no gap where HTTP requests would get 404
		ts.tunnelsMu.Lock()
		existingTunnel := ts.tunnels[subdomain]
		
		// CRITICAL: Register new tunnel FIRST (with valid WSConn)
		// This ensures HTTP requests always see a valid tunnel
		ts.tunnels[subdomain] = tunnel
		registeredCount := len(ts.tunnels)
		ts.tunnelsMu.Unlock()
		
		ts.logger.Debug().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Int("total_tunnels", registeredCount).
			Bool("is_resume", isResume).
			Msg("Registered tunnel in memory")
		
		// NOW mark old tunnel as replaced and close its connection
		// This must happen AFTER new tunnel is registered to avoid 404s
		if existingTunnel != nil && existingTunnel != tunnel {
			// Get old WebSocket connection before marking as replaced
			existingTunnel.mu.RLock()
			oldWSConn := existingTunnel.WSConn
			oldTunnelID := existingTunnel.ID
			existingTunnel.mu.RUnlock()
			
			// Mark old tunnel's WebSocket as replaced
			// This signals the old handler to exit immediately when it checks
			existingTunnel.mu.Lock()
			existingTunnel.WSConn = nil
			existingTunnel.mu.Unlock()
			
			ts.logger.Info().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Str("old_tunnel_id", oldTunnelID).
				Msg("Marked old tunnel connection as replaced - old handler will exit")
			
			// Close old WebSocket connection after marking as replaced
			// This will cause the old handler's ReadJSON to fail and exit
			if oldWSConn != nil && oldWSConn != ws {
				go func() {
					// Give old handler a moment to detect the replacement
					time.Sleep(50 * time.Millisecond)
					oldWSConn.Close()
					ts.logger.Debug().
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Str("old_tunnel_id", oldTunnelID).
						Msg("Closed old WebSocket connection")
				}()
			}
		}
		
		if existingTunnel != nil && existingTunnel != tunnel {
			ts.logger.Info().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Str("old_tunnel_id", existingTunnel.ID).
				Msg("Atomically replaced old tunnel with fresh tunnel for resume")
		}
		
		// Verify registration succeeded immediately
		ts.tunnelsMu.RLock()
		verifyTunnel, verifyExists := ts.tunnels[subdomain]
		ts.tunnelsMu.RUnlock()
		
		if !verifyExists || verifyTunnel != tunnel {
			// Emergency re-registration
			ts.logger.Error().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Bool("verify_exists", verifyExists).
				Msg("CRITICAL: Tunnel registration verification failed - re-registering immediately")
			ts.tunnelsMu.Lock()
			ts.tunnels[subdomain] = tunnel
			ts.tunnelsMu.Unlock()
		} else {
			ts.logger.Info().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Str("local_url", initMsg.LocalURL).
				Int("total_tunnels", registeredCount).
				Msg("Successfully registered resumed tunnel in memory - ready to accept requests")
		}
		
		// Allocate TCP port for TCP/TLS tunnels if needed (resumed tunnels might need port allocation)
		if protocol == ProtocolTCP || protocol == ProtocolTLS {
			tcpPort := ts.allocateTCPPort(tunnel)
			if tcpPort > 0 {
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("protocol", protocol).
					Int("tcp_port", tcpPort).
					Msg("Allocated TCP port for resumed tunnel")
			}
		}

		// Update database: mark as active, update LocalURL, and update last_active_at when resuming
		if ts.repository != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				
				// Update LocalURL first (in case it changed, e.g., different port)
				if err := ts.repository.UpdateTunnelLocalURL(ctx, tunnelID, initMsg.LocalURL); err != nil {
					ts.logger.Error().
						Err(err).
						Str("tunnel_id", tunnelID).
						Str("local_url", initMsg.LocalURL).
						Msg("Failed to update tunnel LocalURL on resume")
				}
				
				// Update status to active
				if err := ts.repository.UpdateTunnelStatus(ctx, tunnelID, "active"); err != nil {
					ts.logger.Error().
						Err(err).
						Str("tunnel_id", tunnelID).
						Msg("Failed to update tunnel status to active on resume from database")
				} else {
					// Update last_active_at when resuming
					if err := ts.repository.UpdateTunnelActivity(ctx, tunnelID, 0); err != nil {
						ts.logger.Error().
							Err(err).
							Str("tunnel_id", tunnelID).
							Msg("Failed to update tunnel last_active_at on resume from database")
					}
				}
			}()
		}

		ts.logger.Info().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Msg("Created fresh tunnel connection for resume (same code path as new tunnel)")
	} else {
		// Create new tunnel connection
		tunnel = &TunnelConnection{
			ID:           tunnelID,
			Subdomain:    subdomain,
			LocalURL:     initMsg.LocalURL,
			Protocol:     protocol,
			Host:         initMsg.Host,
			WSConn:       ws,
			CreatedAt:    time.Now(),
			LastActive:   time.Now(),
			RequestCount: 0,
			handlerReady: false, // Will be set to true when handler starts
		}

		// Register tunnel
		ts.tunnelsMu.Lock()
		ts.tunnels[subdomain] = tunnel
		ts.tunnelsMu.Unlock()

		// Allocate TCP port for TCP/TLS tunnels
		if protocol == ProtocolTCP || protocol == ProtocolTLS {
			tcpPort := ts.allocateTCPPort(tunnel)
			if tcpPort > 0 {
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("protocol", protocol).
					Int("tcp_port", tcpPort).
					Msg("Allocated TCP port for tunnel")
			}
		}
	}

	// Safety check: ensure tunnel is set
	if tunnel == nil {
		ts.logger.Error().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Bool("is_resume", isResume).
			Msg("CRITICAL: tunnel is nil after creation/resume - this should never happen")
		ws.WriteJSON(map[string]string{"error": "internal error: tunnel not created"})
		return
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

	// Generate public URL using domain manager
	// For TCP/TLS tunnels, use the allocated TCP port
	var publicURL string
	tunnel.mu.RLock()
	tunnelProtocol := tunnel.Protocol
	tunnel.mu.RUnlock()

	if tunnelProtocol == ProtocolTCP || tunnelProtocol == ProtocolTLS {
		// Get allocated TCP port
		ts.portMapMu.RLock()
		var tcpPort int
		for port, t := range ts.portMap {
			if t.ID == tunnelID {
				tcpPort = port
				break
			}
		}
		ts.portMapMu.RUnlock()

		if tcpPort > 0 {
			// For TCP/TLS, public URL is host:port format
			if ts.domainManager != nil && ts.domainManager.baseDomain != "" {
				publicURL = fmt.Sprintf("%s:%d", ts.domainManager.GetPublicURL(subdomain, ts.port, false), tcpPort)
			} else {
				publicURL = fmt.Sprintf("%s.localhost:%d", subdomain, tcpPort)
			}
		} else {
			// Fallback if port allocation failed
			publicURL = fmt.Sprintf("%s.localhost:%d", subdomain, ts.port)
		}
	} else {
		// HTTP tunnel - use HTTP URL
		publicURL = fmt.Sprintf("http://%s.localhost:%d", subdomain, ts.port)
		if ts.domainManager != nil {
			// Use HTTPS if base domain is set (Coolify will handle SSL termination)
			useHTTPS := ts.domainManager.baseDomain != ""
			publicURL = ts.domainManager.GetPublicURL(subdomain, ts.port, useHTTPS)
		}
	}

	// Save tunnel to database if repository is available
	ts.logger.Debug().
		Bool("repository_available", ts.repository != nil).
		Bool("is_resume", isResume).
		Str("tunnel_id", tunnelID).
		Str("subdomain", subdomain).
		Msg("Checking if tunnel should be saved to database")

	if ts.repository != nil && !isResume {
		// Try to extract user ID from auth token if provided
		var userID string
		if initMsg.Token != "" {
			if ts.jwtValidator != nil {
				if extractedUserID, err := ts.jwtValidator(initMsg.Token); err == nil && extractedUserID != "" {
					userID = extractedUserID
					ts.logger.Info().
						Str("tunnel_id", tunnelID).
						Str("user_id", userID).
						Msg("Extracted user ID from auth token - tunnel will be associated with user")
				} else {
					ts.logger.Warn().
						Err(err).
						Str("tunnel_id", tunnelID).
						Msg("Failed to extract user ID from token (tunnel will be unassociated) - JWT validator may not be configured or token invalid")
				}
			} else {
				ts.logger.Warn().
					Str("tunnel_id", tunnelID).
					Msg("Auth token provided but JWT validator not configured - tunnel will be unassociated. Set JWT_SECRET in tunnel server config.")
			}
		} else {
			ts.logger.Debug().
				Str("tunnel_id", tunnelID).
				Msg("No auth token provided - tunnel will be unassociated")
		}

		dbTunnel := &Tunnel{
			ID:           tunnelID,
			UserID:       userID, // Set user_id if token was validated
			Subdomain:    subdomain,
			LocalURL:     initMsg.LocalURL,
			PublicURL:    publicURL,
			Status:       "active",
			RequestCount: 0,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastActive:   time.Now(),
		}

		ts.logger.Info().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Str("user_id", userID).
			Str("local_url", initMsg.LocalURL).
			Str("public_url", publicURL).
			Msg("Attempting to save tunnel to database")

		// Save original tunnel ID before CreateTunnel (which may modify it)
		originalTunnelID := tunnelID
		if err := ts.repository.CreateTunnel(context.Background(), dbTunnel); err != nil {
			ts.logger.Error().
				Err(err).
				Str("tunnel_id", tunnelID).
				Str("tunnel_id_length", fmt.Sprintf("%d", len(tunnelID))).
				Str("subdomain", subdomain).
				Str("user_id", userID).
				Str("local_url", initMsg.LocalURL).
				Str("public_url", publicURL).
				Msg("Failed to save tunnel to database - check database connection and tunnel ID format")
			// Continue even if database save fails - tunnel is still functional in memory
		} else {
			// CreateTunnel uses UPSERT - if subdomain already exists, it updates LocalURL and returns existing ID
			// Update the tunnel connection ID to match what's in the database
			if dbTunnel.ID != originalTunnelID {
				ts.logger.Info().
					Str("original_tunnel_id", originalTunnelID).
					Str("database_tunnel_id", dbTunnel.ID).
					Str("subdomain", subdomain).
					Str("local_url", initMsg.LocalURL).
					Msg("Tunnel with this subdomain already existed - updated LocalURL and using existing tunnel ID")
				tunnelID = dbTunnel.ID
				tunnel.ID = dbTunnel.ID
			} else {
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("subdomain", subdomain).
					Str("local_url", initMsg.LocalURL).
					Msg("Created new tunnel in database")
			}

			if userID != "" {
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("subdomain", subdomain).
					Str("user_id", userID).
					Str("local_url", initMsg.LocalURL).
					Msg("Saved tunnel to database with user association")
			} else {
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("subdomain", subdomain).
					Str("local_url", initMsg.LocalURL).
					Msg("Saved tunnel to database (unassociated - user can associate later)")
			}
		}
	}

	// Verify tunnel is properly registered before sending response
	ts.tunnelsMu.RLock()
	registeredTunnel, isRegistered := ts.tunnels[subdomain]
	ts.tunnelsMu.RUnlock()
	
	if !isRegistered || registeredTunnel != tunnel {
		ts.logger.Error().
			Str("subdomain", subdomain).
			Bool("is_registered", isRegistered).
			Msg("CRITICAL: Tunnel not properly registered before sending response - registering now")
		// Emergency registration
		ts.tunnelsMu.Lock()
		ts.tunnels[subdomain] = tunnel
		ts.tunnelsMu.Unlock()
	}
	
	// Verify WebSocket connection is set
	if tunnel.WSConn == nil {
		ts.logger.Error().
			Str("subdomain", subdomain).
			Str("tunnel_id", tunnelID).
			Msg("CRITICAL: WebSocket connection is nil - cannot start message handler")
		ws.WriteJSON(map[string]string{"error": "internal error: websocket connection not set"})
		return
	}

	// CRITICAL: Start handler BEFORE sending success response
	// This ensures the tunnel is fully ready when client receives "Tunnel Connected Successfully!"
	// Verify tunnel is registered and WebSocket is set before starting handler
	ts.tunnelsMu.RLock()
	finalTunnel, finalExists := ts.tunnels[subdomain]
	ts.tunnelsMu.RUnlock()
	
	if !finalExists || finalTunnel != tunnel {
		ts.logger.Error().
			Str("subdomain", subdomain).
			Str("tunnel_id", tunnelID).
			Bool("final_exists", finalExists).
			Msg("CRITICAL: Tunnel not in registry before starting handler - re-registering")
		ts.tunnelsMu.Lock()
		ts.tunnels[subdomain] = tunnel
		ts.tunnelsMu.Unlock()
	}
	
	tunnel.mu.RLock()
	wsConnCheck := tunnel.WSConn
	tunnel.mu.RUnlock()
	
	if wsConnCheck == nil {
		ts.logger.Error().
			Str("subdomain", subdomain).
			Str("tunnel_id", tunnelID).
			Msg("CRITICAL: WebSocket connection is nil before starting handler")
		return
	}
	
	ts.logger.Info().
		Str("tunnel_id", tunnelID).
		Str("subdomain", subdomain).
		Str("public_url", publicURL).
		Bool("is_resume", isResume).
		Bool("ws_conn_set", wsConnCheck != nil).
		Msg("Tunnel ready - starting message handler")

	// CRITICAL: Start handler immediately - it will be ready to process requests
	// The handler starts its message loop immediately, so it's ready as soon as goroutine starts
	go ts.handleTunnelMessages(tunnel)
	
	// CRITICAL: Wait for handler to be ready BEFORE sending success response
	// This ensures HTTP requests won't arrive before handler is initialized
	// For resume, this is especially important to prevent 404s
	maxWait := 100 // Wait up to 1 second (increased for resume reliability)
	for i := 0; i < maxWait; i++ {
		tunnel.mu.RLock()
		ready := tunnel.handlerReady
		wsConnStillValid := tunnel.WSConn != nil
		tunnel.mu.RUnlock()
		if ready && wsConnStillValid {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	
	// Final check
	tunnel.mu.RLock()
	handlerReady := tunnel.handlerReady
	tunnel.mu.RUnlock()
	
	if !handlerReady {
		ts.logger.Warn().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Bool("is_resume", isResume).
			Msg("Handler not ready after wait - sending response anyway (handler should be ready soon)")
	} else {
		ts.logger.Info().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Bool("is_resume", isResume).
			Msg("Handler confirmed ready - sending success response to client")
	}
	
	// Final verification: ensure tunnel is still registered and ready
	ts.tunnelsMu.RLock()
	finalVerify, finalVerifyExists := ts.tunnels[subdomain]
	tunnelReady := finalVerifyExists && finalVerify == tunnel
	ts.tunnelsMu.RUnlock()
	
	if !tunnelReady {
		ts.logger.Error().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Msg("CRITICAL: Tunnel disappeared before sending success response - emergency re-register")
		ts.tunnelsMu.Lock()
		ts.tunnels[subdomain] = tunnel
		ts.tunnelsMu.Unlock()
		tunnelReady = true // Update after re-registering
	}
	
	// CRITICAL: Final verification that tunnel is accessible via handleHTTPRequest
	// This ensures that when the client receives success, HTTP requests will definitely find the tunnel
	ts.tunnelsMu.RLock()
	verifyAccessible, accessibleExists := ts.tunnels[subdomain]
	accessible := accessibleExists && verifyAccessible == tunnel && verifyAccessible.WSConn != nil
	ts.tunnelsMu.RUnlock()
	
	if !accessible {
		ts.logger.Error().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Bool("exists", accessibleExists).
			Bool("same_tunnel", accessibleExists && verifyAccessible == tunnel).
			Msg("CRITICAL: Tunnel not accessible before sending success - re-registering")
		ts.tunnelsMu.Lock()
		ts.tunnels[subdomain] = tunnel
		ts.tunnelsMu.Unlock()
	}

	// Send confirmation - handler is now ready
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
	
	ts.logger.Info().
		Str("tunnel_id", tunnelID).
		Str("subdomain", subdomain).
		Bool("is_resume", isResume).
		Msg("Tunnel connected and ready")
	
	// Final verification after a brief moment to ensure tunnel stayed registered
	go func() {
		time.Sleep(200 * time.Millisecond)
		ts.tunnelsMu.RLock()
		stillThere, stillExists := ts.tunnels[subdomain]
		ts.tunnelsMu.RUnlock()
		
		if !stillExists || stillThere != tunnel {
			ts.logger.Error().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Bool("still_exists", stillExists).
				Msg("CRITICAL: Tunnel was removed from registry shortly after handler started - re-registering")
			ts.tunnelsMu.Lock()
			ts.tunnels[subdomain] = tunnel
			ts.tunnelsMu.Unlock()
			ts.logger.Info().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Msg("Re-registered tunnel after it was removed")
		} else {
			ts.logger.Debug().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Msg("Tunnel verification passed - still registered after handler started")
		}
	}()
}

// handleTunnelMessages handles messages from tunnel client
func (ts *TunnelServer) handleTunnelMessages(tunnel *TunnelConnection) {
	// Store reference to the WebSocket connection we're handling
	// This allows us to detect if the connection has been replaced (during resume)
	tunnel.mu.RLock()
	ourWSConn := tunnel.WSConn
	tunnelSubdomain := tunnel.Subdomain
	tunnelID := tunnel.ID
	tunnel.mu.RUnlock()
	
	if ourWSConn == nil {
		ts.logger.Error().
			Str("tunnel_id", tunnelID).
			Str("subdomain", tunnelSubdomain).
			Msg("CRITICAL: handleTunnelMessages started with nil WebSocket connection - exiting")
		return
	}
	
	ts.logger.Info().
		Str("tunnel_id", tunnelID).
		Str("subdomain", tunnelSubdomain).
		Msg("handleTunnelMessages started - beginning message loop and ready to process requests")
	
	// Mark handler as ready
	tunnel.mu.Lock()
	tunnel.handlerReady = true
	tunnel.mu.Unlock()
	
	// Verify tunnel is still registered after a brief moment (catch any immediate removal)
	// Minimal sleep to make handler ready as fast as possible
	time.Sleep(20 * time.Millisecond)
	ts.tunnelsMu.RLock()
	stillRegistered, stillExists := ts.tunnels[tunnelSubdomain]
	ts.tunnelsMu.RUnlock()
	
	if !stillExists || stillRegistered != tunnel {
		ts.logger.Error().
			Str("tunnel_id", tunnelID).
			Str("subdomain", tunnelSubdomain).
			Bool("still_exists", stillExists).
			Msg("CRITICAL: Tunnel was removed from registry immediately after handler started - re-registering")
		// Re-register the tunnel
		ts.tunnelsMu.Lock()
		ts.tunnels[tunnelSubdomain] = tunnel
		ts.tunnelsMu.Unlock()
		ts.logger.Info().
			Str("tunnel_id", tunnelID).
			Str("subdomain", tunnelSubdomain).
			Msg("Re-registered tunnel after immediate removal")
	}
	
	defer func() {
		// Only close if this is still our connection (hasn't been replaced)
		tunnel.mu.RLock()
		currentWSConn := tunnel.WSConn
		tunnel.mu.RUnlock()
		
		if currentWSConn == ourWSConn {
			ourWSConn.Close()
		} else {
			ts.logger.Debug().
				Str("tunnel_id", tunnelID).
				Str("subdomain", tunnelSubdomain).
				Msg("Skipping close - WebSocket connection has been replaced (tunnel resumed)")
		}
	}()

	// Set read deadline to detect dead connections (120 seconds - longer than client heartbeat interval)
	// This ensures we detect connections that are "half-open" (TCP exists but no data flows)
	heartbeatInterval := 30 * time.Second
	readDeadline := 4 * heartbeatInterval // 120 seconds - should receive at least one ping/pong in this time

	for {
		// Check if connection has been replaced (tunnel was resumed)
		tunnel.mu.RLock()
		currentWSConn := tunnel.WSConn
		tunnel.mu.RUnlock()
		
		// Exit if connection is nil (marked as replaced) or different (replaced with new connection)
		if currentWSConn == nil || currentWSConn != ourWSConn {
			// Connection has been replaced or marked as replaced - our goroutine is obsolete, exit silently
			ts.logger.Debug().
				Str("tunnel_id", tunnel.ID).
				Str("subdomain", tunnelSubdomain).
				Bool("ws_conn_nil", currentWSConn == nil).
				Bool("ws_conn_different", currentWSConn != nil && currentWSConn != ourWSConn).
				Msg("WebSocket connection replaced or marked as replaced - old handler exiting (tunnel resumed)")
			return // Don't call removeTunnel - new connection is already active
		}
		
		// Set read deadline before each read to detect stale connections
		ourWSConn.SetReadDeadline(time.Now().Add(readDeadline))
		
		var msg TunnelMessage
		if err := ourWSConn.ReadJSON(&msg); err != nil {
			// Check again if connection was replaced during ReadJSON
			tunnel.mu.RLock()
			currentWSConn = tunnel.WSConn
			tunnel.mu.RUnlock()
			
			if currentWSConn != ourWSConn {
				// Connection was replaced during read - don't remove tunnel
				ts.logger.Debug().
					Str("tunnel_id", tunnelID).
					Str("subdomain", tunnelSubdomain).
					Msg("WebSocket connection replaced during read - old handler exiting (tunnel resumed)")
				return
			}
			
			// Check if this is a normal close (client shutdown) vs. unexpected error
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("subdomain", tunnelSubdomain).
					Msg("Tunnel client closed connection normally")
			} else {
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("subdomain", tunnelSubdomain).
					Err(err).
					Msg("Tunnel disconnected (unexpected error)")
			}
			
			ts.removeTunnel(tunnelSubdomain)
			return
		}
		
		// Reset read deadline after successful read (connection is alive)
		tunnel.WSConn.SetReadDeadline(time.Time{}) // Clear deadline

		tunnel.mu.Lock()
		tunnel.LastActive = time.Now()
		tunnel.mu.Unlock()
		
		// Periodically update database last_active_at (every 30 seconds or on ping)
		// This ensures the database reflects that the tunnel is still alive
		// We do this on ping messages to avoid too frequent DB updates
		if msg.Type == MsgTypePing {
			if ts.repository != nil {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					tunnel.mu.RLock()
					tunnelID := tunnel.ID
					requestCount := tunnel.RequestCount
					tunnel.mu.RUnlock()
					
					// Update last_active_at in database (keep request count as-is)
					if err := ts.repository.UpdateTunnelActivity(ctx, tunnelID, requestCount); err != nil {
						ts.logger.Debug().
							Err(err).
							Str("tunnel_id", tunnelID).
							Msg("Failed to update tunnel last_active_at on ping (non-critical)")
					}
				}()
			}
		}

		// Handle different message types
		switch msg.Type {
		case MsgTypePing:
			tunnel.WSConn.WriteJSON(TunnelMessage{Type: MsgTypePong})
		case MsgTypeHTTPResponse:
			// Complete pending HTTP request with response
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
			// Fail pending HTTP request with error
			if msg.Error != nil {
				// Create error with both error type and message for better detection
				// Format: "error_type: error_message" so we can detect connection_refused
				err := fmt.Errorf("%s: %s", msg.Error.Error, msg.Error.Message)
				ts.logger.Info().
					Str("request_id", msg.RequestID).
					Str("error_type", msg.Error.Error).
					Str("error_message", msg.Error.Message).
					Str("formatted_error", err.Error()).
					Msg("Received HTTP error from tunnel client")

				if err2 := ts.requestTracker.FailRequest(msg.RequestID, err); err2 != nil {
					ts.logger.Error().
						Err(err2).
						Str("request_id", msg.RequestID).
						Msg("Failed to fail request in tracker")
				} else {
					ts.logger.Info().
						Str("request_id", msg.RequestID).
						Str("error_type", msg.Error.Error).
						Str("error_message", msg.Error.Message).
						Msg("Successfully failed HTTP request in tracker - waiting for forwardHTTPRequest to detect and show error page")
				}
			}
		case MsgTypeTCPData, MsgTypeTLSData:
			// Forward TCP/TLS data to pending connection
			// Empty data with RequestID indicates new connection request
			if msg.RequestID != "" {
				ts.forwardTCPData(tunnel, msg.RequestID, msg.Data)
			}
		case MsgTypeTCPError, MsgTypeTLSError:
			// Handle TCP/TLS errors
			if msg.Error != nil {
				ts.handleTCPError(tunnel, msg.RequestID, msg.Error)
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
	ts.logger.Debug().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("host", r.Host).
		Str("remote_addr", r.RemoteAddr).
		Msg("Received HTTP request")

	// Add security headers
	ts.security.AddSecurityHeaders(w, r)
	if r.Method == http.MethodOptions {
		return // Preflight handled
	}

	// Validate request
	if err := ts.security.ValidateRequest(r); err != nil {
		ts.logger.Warn().Err(err).Str("method", r.Method).Str("path", r.URL.Path).Msg("Invalid request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Sanitize path
	r.URL.Path = ts.security.SanitizePath(r.URL.Path)

	// Extract subdomain from Host header
	host := r.Host
	subdomain := extractSubdomain(host)

	ts.logger.Debug().
		Str("host", host).
		Str("extracted_subdomain", subdomain).
		Msg("Extracted subdomain from host header")

	if subdomain == "" {
		// No subdomain, return info page or 404
		ts.logger.Debug().Str("host", host).Msg("No subdomain found, showing root page")
		ts.handleRootRequest(w, r)
		return
	}

	ts.tunnelsMu.RLock()
	tunnel, exists := ts.tunnels[subdomain]
	// Log all available tunnels for debugging
	availableSubdomains := make([]string, 0, len(ts.tunnels))
	tunnelDetails := make(map[string]string)
	for sub, t := range ts.tunnels {
		availableSubdomains = append(availableSubdomains, sub)
		t.mu.RLock()
		tunnelDetails[sub] = fmt.Sprintf("id=%s,protocol=%s,local=%s", t.ID, t.Protocol, t.LocalURL)
		t.mu.RUnlock()
	}
	ts.tunnelsMu.RUnlock()

	ts.logger.Info().
		Str("subdomain", subdomain).
		Bool("exists_in_memory", exists).
		Strs("available_tunnels", availableSubdomains).
		Interface("tunnel_details", tunnelDetails).
		Str("request_path", r.URL.Path).
		Str("request_method", r.Method).
		Str("request_host", r.Host).
		Bool("has_repository", ts.repository != nil).
		Msg("HTTP request received - checking tunnel in memory")
	
	// CRITICAL DEBUG: Log if tunnel exists but we might still return 404
	if exists && tunnel != nil {
		tunnel.mu.RLock()
		tunnelID := tunnel.ID
		tunnelLocalURL := tunnel.LocalURL
		tunnelProtocol := tunnel.Protocol
		wsConn := tunnel.WSConn
		handlerReady := tunnel.handlerReady
		tunnel.mu.RUnlock()
		
		ts.logger.Info().
			Str("subdomain", subdomain).
			Str("tunnel_id", tunnelID).
			Str("tunnel_local_url", tunnelLocalURL).
			Str("tunnel_protocol", tunnelProtocol).
			Bool("ws_conn_nil", wsConn == nil).
			Bool("handler_ready", handlerReady).
			Msg("Tunnel found in memory - checking readiness")
	}
	
	if exists && tunnel != nil {
		tunnel.mu.RLock()
		wsConn := tunnel.WSConn
		tunnel.mu.RUnlock()
		
		ts.logger.Debug().
			Str("tunnel_id", tunnel.ID).
			Str("tunnel_subdomain", tunnel.Subdomain).
			Str("tunnel_protocol", tunnel.Protocol).
			Str("tunnel_local_url", tunnel.LocalURL).
			Bool("ws_conn_nil", wsConn == nil).
			Msg("Found tunnel in memory - details")
		
		// CRITICAL: Check if WebSocket connection is actually valid
		if wsConn == nil {
			ts.logger.Warn().
				Str("tunnel_id", tunnel.ID).
				Str("subdomain", subdomain).
				Msg("Tunnel exists in memory but WebSocket connection is nil - tunnel not ready")
			// Treat as if tunnel doesn't exist - it's not ready to handle requests
			exists = false
		} else {
			// Check if handler is ready (for resumed tunnels, handler might not be ready yet)
			tunnel.mu.RLock()
			handlerReady := tunnel.handlerReady
			tunnel.mu.RUnlock()
			
			if !handlerReady {
				ts.logger.Debug().
					Str("tunnel_id", tunnel.ID).
					Str("subdomain", subdomain).
					Msg("Tunnel exists but handler not ready yet - waiting briefly")
				// Give handler a moment to initialize (max 100ms)
				for i := 0; i < 10; i++ {
					time.Sleep(10 * time.Millisecond)
					tunnel.mu.RLock()
					handlerReady = tunnel.handlerReady
					tunnel.mu.RUnlock()
					if handlerReady {
						break
					}
				}
				if !handlerReady {
					ts.logger.Warn().
						Str("tunnel_id", tunnel.ID).
						Str("subdomain", subdomain).
						Msg("Tunnel handler still not ready after wait - proceeding anyway (handler should be ready soon)")
					// Continue anyway - handler should be ready very soon
					// The request will be queued in the request tracker and handled when response arrives
				} else {
					ts.logger.Debug().
						Str("tunnel_id", tunnel.ID).
						Str("subdomain", subdomain).
						Msg("Tunnel handler is ready - proceeding with request")
				}
			}
		}
	}

	if !exists {
		// Tunnel not in memory - check database to see if it exists but is disconnected
		// CRITICAL: Also check if a resume might be in progress (tunnel was just created)
		// Give it a brief moment in case resume is happening right now
		if ts.repository != nil {
			// First, wait longer in case resume is in progress (max 1 second)
			// This is important because resume can take a moment to register the tunnel
			for i := 0; i < 100; i++ {
				time.Sleep(10 * time.Millisecond)
				ts.tunnelsMu.RLock()
				tunnel, exists = ts.tunnels[subdomain]
				ts.tunnelsMu.RUnlock()
				if exists && tunnel != nil {
					ts.logger.Info().
						Str("subdomain", subdomain).
						Int("wait_iterations", i+1).
						Msg("Tunnel appeared in memory during wait - resume likely in progress")
					
					// CRITICAL: Also wait for handler to be ready
					tunnel.mu.RLock()
					handlerReady := tunnel.handlerReady
					wsConn := tunnel.WSConn
					tunnel.mu.RUnlock()
					
					if wsConn != nil && !handlerReady {
						ts.logger.Debug().
							Str("subdomain", subdomain).
							Msg("Tunnel found but handler not ready - waiting for handler")
						// Wait up to 500ms more for handler to be ready
						for j := 0; j < 50; j++ {
							time.Sleep(10 * time.Millisecond)
							tunnel.mu.RLock()
							handlerReady = tunnel.handlerReady
							tunnel.mu.RUnlock()
							if handlerReady {
								ts.logger.Info().
									Str("subdomain", subdomain).
									Int("handler_wait_iterations", j+1).
									Msg("Handler is now ready after wait")
								break
							}
						}
					}
					break
				}
			}
			
			// If still not in memory, check database
			if !exists {
				dbTunnel, err := ts.repository.GetTunnelBySubdomain(context.Background(), subdomain)
				if err == nil && dbTunnel != nil {
					// Tunnel exists in database but client is not connected
					// Return 503 (Service Unavailable) to indicate tunnel exists but is disconnected
					// This is different from 404 (tunnel doesn't exist) and 502 (local server error)
					ts.logger.Info().
						Str("subdomain", subdomain).
						Str("tunnel_id", dbTunnel.ID).
						Str("status", dbTunnel.Status).
						Str("local_url", dbTunnel.LocalURL).
						Msg("Tunnel exists in database but not in memory - showing disconnected page (503)")
					ts.writeErrorPage(w, r, nil, http.StatusServiceUnavailable, "Tunnel Disconnected", "The tunnel exists but is not currently connected.", fmt.Sprintf("The subdomain '%s' is associated with a tunnel, but the tunnel client is not connected. Please start the tunnel client to resume this tunnel.", html.EscapeString(subdomain)))
					return
				}
			}
		}
		
		// If still not exists after checking database and waiting, return 404
		// BUT: Only return 404 if we're sure the tunnel doesn't exist
		// If repository is nil or database check failed, we can't be sure, so log it
		if !exists {
			if ts.repository == nil {
				ts.logger.Warn().
					Str("subdomain", subdomain).
					Msg("Tunnel not found in memory and no database available - returning 404")
			} else {
				ts.logger.Info().
					Str("subdomain", subdomain).
					Msg("Tunnel not found in memory or database - returning 404")
			}
			ts.writeErrorPage(w, r, nil, http.StatusNotFound, "Tunnel Not Found", "The requested tunnel does not exist.", fmt.Sprintf("The subdomain '%s' is not associated with any tunnel. Please check the URL or create a new tunnel.", html.EscapeString(subdomain)))
			return
		}
	}

	// Final check: ensure tunnel is still valid before forwarding
	if tunnel == nil {
		ts.logger.Error().
			Str("subdomain", subdomain).
			Msg("CRITICAL: Tunnel is nil after all checks - returning 500")
		http.Error(w, "Internal server error: tunnel is nil", http.StatusInternalServerError)
		return
	}

	// Check tunnel protocol - only forward HTTP requests to HTTP tunnels
	tunnel.mu.RLock()
	protocol := tunnel.Protocol
	wsConn := tunnel.WSConn
	handlerReady := tunnel.handlerReady
	tunnel.mu.RUnlock()

	if protocol != ProtocolHTTP {
		ts.logger.Warn().
			Str("subdomain", subdomain).
			Str("protocol", protocol).
			Msg("Tunnel protocol mismatch - returning 400")
		http.Error(w, fmt.Sprintf("Tunnel protocol is %s, not HTTP", protocol), http.StatusBadRequest)
		return
	}

	if wsConn == nil {
		ts.logger.Error().
			Str("subdomain", subdomain).
			Str("tunnel_id", tunnel.ID).
			Msg("CRITICAL: WebSocket connection is nil when forwarding - returning 503")
		ts.writeErrorPage(w, r, tunnel, http.StatusServiceUnavailable, "Tunnel Connection Lost", "The tunnel connection was lost.", "The WebSocket connection is not available. Please reconnect the tunnel client.")
		return
	}

	ts.logger.Info().
		Str("subdomain", subdomain).
		Str("tunnel_id", tunnel.ID).
		Bool("handler_ready", handlerReady).
		Str("protocol", protocol).
		Str("local_url", tunnel.LocalURL).
		Msg("Forwarding HTTP request to tunnel client")

	// Forward request to tunnel client
	ts.forwardHTTPRequest(tunnel, w, r)
}

// handleRootRequest handles requests to root domain
func (ts *TunnelServer) handleRootRequest(w http.ResponseWriter, r *http.Request) {
	// Get active tunnel count from memory (tunnels currently connected)
	ts.tunnelsMu.RLock()
	activeTunnelsInMemory := len(ts.tunnels)
	ts.tunnelsMu.RUnlock()

	// Get total tunnel count from database if available
	activeTunnels := activeTunnelsInMemory
	totalTunnels := activeTunnelsInMemory
	
	if ts.repository != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		// Get ALL tunnels from database (not filtered by user)
		tunnels, err := ts.repository.ListAllTunnels(ctx)
		if err != nil {
			ts.logger.Debug().Err(err).Msg("Failed to fetch tunnels from database for root page")
			// Continue with in-memory counts if database query fails
		} else {
			totalTunnels = len(tunnels)
			// Count active tunnels from database (status='active')
			activeCount := 0
			for _, t := range tunnels {
				if t.Status == "active" {
					activeCount++
				}
			}
			// Use database active count (more accurate than in-memory)
			activeTunnels = activeCount
			
			ts.logger.Debug().
				Int("total_tunnels", totalTunnels).
				Int("active_tunnels", activeTunnels).
				Int("active_in_memory", activeTunnelsInMemory).
				Msg("Root page: tunnel counts from database")
		}
	} else {
		ts.logger.Debug().Msg("Root page: repository not available, using in-memory counts only")
	}

	// Add security headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'none'; img-src 'self' data:; font-src 'self' data:;")
	w.WriteHeader(http.StatusOK)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" class="dark">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>UniRoute Tunnel Server</title>
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
			background: linear-gradient(135deg, #0f172a 0%%, #1e3a8a 50%%, #312e81 100%%);
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
			color: #f1f5f9;
			padding: 20px;
		}
		.container {
			max-width: 700px;
			width: 100%%;
			background: rgba(15, 23, 42, 0.8);
			backdrop-filter: blur(12px);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 16px;
			padding: 48px;
			box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3), 0 10px 10px -5px rgba(0, 0, 0, 0.2);
		}
		.logo {
			display: flex;
			align-items: center;
			justify-content: center;
			margin-bottom: 40px;
		}
		.logo-icon {
			width: 64px;
			height: 64px;
			background: linear-gradient(135deg, #3b82f6 0%%, #6366f1 50%%, #a855f7 100%%);
			border-radius: 16px;
			display: flex;
			align-items: center;
			justify-content: center;
			box-shadow: 0 10px 15px -3px rgba(59, 130, 246, 0.3);
			margin-right: 16px;
		}
		.logo-icon span {
			color: white;
			font-weight: bold;
			font-size: 32px;
		}
		.logo-text {
			font-size: 36px;
			font-weight: bold;
			color: white;
		}
		.status-badge {
			display: inline-flex;
			align-items: center;
			background: rgba(34, 197, 94, 0.1);
			border: 1px solid rgba(34, 197, 94, 0.3);
			border-radius: 20px;
			padding: 8px 16px;
			margin-bottom: 32px;
		}
		.status-dot {
			width: 8px;
			height: 8px;
			background: #22c55e;
			border-radius: 50%%;
			margin-right: 8px;
			animation: pulse 2s infinite;
		}
		@keyframes pulse {
			0%%, 100%% { opacity: 1; }
			50%% { opacity: 0.5; }
		}
		.status-text {
			color: #22c55e;
			font-size: 14px;
			font-weight: 600;
		}
		h1 {
			font-size: 32px;
			font-weight: bold;
			text-align: center;
			margin-bottom: 8px;
			color: white;
		}
		.subtitle {
			text-align: center;
			color: #cbd5e1;
			margin-bottom: 40px;
			font-size: 16px;
		}
		.stats-grid {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
			gap: 20px;
			margin-bottom: 32px;
		}
		.stat-card {
			background: rgba(30, 41, 59, 0.6);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 12px;
			padding: 24px;
			text-align: center;
		}
		.stat-label {
			color: #94a3b8;
			font-size: 14px;
			font-weight: 500;
			margin-bottom: 8px;
		}
		.stat-value {
			color: white;
			font-size: 36px;
			font-weight: bold;
			font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
		}
		.info-section {
			background: rgba(30, 41, 59, 0.6);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 12px;
			padding: 24px;
			margin-top: 32px;
		}
		.info-title {
			font-size: 18px;
			font-weight: 600;
			margin-bottom: 16px;
			color: white;
		}
		.info-item {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: 12px 0;
			border-bottom: 1px solid rgba(148, 163, 184, 0.1);
		}
		.info-item:last-child {
			border-bottom: none;
		}
		.info-label {
			color: #94a3b8;
			font-size: 14px;
		}
		.info-value {
			color: white;
			font-size: 14px;
			font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
		}
		.footer {
			text-align: center;
			margin-top: 32px;
			color: #64748b;
			font-size: 13px;
		}
		.footer a {
			color: #60a5fa;
			text-decoration: none;
		}
		.footer a:hover {
			text-decoration: underline;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="logo">
			<div class="logo-icon">
				<span>U</span>
			</div>
			<span class="logo-text">UniRoute</span>
		</div>
		
		<div style="text-align: center; margin-bottom: 24px;">
			<div class="status-badge">
				<div class="status-dot"></div>
				<span class="status-text">Running</span>
			</div>
		</div>
		
		<h1>Tunnel Server</h1>
		<p class="subtitle">Secure tunneling service for exposing local services</p>
		
		<div class="stats-grid">
			<div class="stat-card">
				<div class="stat-label">Active Tunnels</div>
				<div class="stat-value">%d</div>
			</div>
			<div class="stat-card">
				<div class="stat-label">Total Tunnels</div>
				<div class="stat-value">%d</div>
			</div>
		</div>
		
		<div class="info-section">
			<div class="info-title">Server Information</div>
			<div class="info-item">
				<span class="info-label">Status</span>
				<span class="info-value">Operational</span>
			</div>
			<div class="info-item">
				<span class="info-label">Health Check</span>
				<span class="info-value"><a href="/health" style="color: #60a5fa;">/health</a></span>
			</div>
		</div>
		
		<div class="footer">
			Powered by <a href="https://uniroute.co" target="_blank">UniRoute</a> | 
			<a href="/" style="color: #60a5fa; text-decoration: none;">Refresh</a>
		</div>
	</div>
	<script>
		// Auto-refresh every 5 seconds to show updated tunnel counts
		setTimeout(function() {
			window.location.reload();
		}, 5000);
	</script>
</body>
</html>`, activeTunnels, totalTunnels)

	w.Write([]byte(html))
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

	// Check rate limit
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
	
	ts.logger.Info().
		Str("request_id", requestID).
		Str("method", reqData.Method).
		Str("path", reqData.Path).
		Str("query", reqData.Query).
		Str("local_url", tunnel.LocalURL).
		Int("body_size", len(reqData.Body)).
		Msg("Serialized request - will forward to tunnel client")

	// Record request for rate limiting
	ts.rateLimiter.RecordRequest(r.Context(), tunnel.ID)

	// Send request through WebSocket
	msg := TunnelMessage{
		Type:      MsgTypeHTTPRequest,
		RequestID: requestID,
		Request:   reqData,
	}

	// Check if WebSocket connection is still valid before sending
	tunnel.mu.RLock()
	wsConn := tunnel.WSConn
	tunnel.mu.RUnlock()

	if wsConn == nil {
		ts.requestTracker.FailRequest(requestID, fmt.Errorf("tunnel connection lost"))
		ts.writeErrorPage(w, r, tunnel, http.StatusBadGateway, "Tunnel Connection Lost", "The tunnel connection was lost while processing your request.", "The tunnel client disconnected. Please reconnect the tunnel client.")
		return
	}

	if err := wsConn.WriteJSON(msg); err != nil {
		ts.requestTracker.FailRequest(requestID, err)
		ts.logger.Error().Err(err).Str("request_id", requestID).Str("tunnel_id", tunnel.ID).Msg("Failed to send request to tunnel client - connection may be broken")
		ts.writeErrorPage(w, r, tunnel, http.StatusBadGateway, "Tunnel Connection Error", "Failed to forward request to tunnel client.", html.EscapeString(err.Error()))
		return
	}

	// Wait for response using request tracker
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := pendingReq.WaitForResponse(ctx)
	if err != nil {
		ts.logger.Error().
			Err(err).
			Str("request_id", requestID).
			Str("tunnel_id", tunnel.ID).
			Str("subdomain", tunnel.Subdomain).
			Str("error_string", err.Error()).
			Msg("Failed to receive response from tunnel client")

		// Check if this is a connection refused error and return a nice error page
		// Check for various forms of connection refused errors (case-insensitive)
		// The error format from HTTPError is: "error_type: error_message"
		// e.g., "connection_refused: dial tcp [::1]:8890: connect: connection refused"
		errMsg := err.Error()
		errMsgLower := strings.ToLower(errMsg)

		ts.logger.Info().
			Str("request_id", requestID).
			Str("error", errMsg).
			Str("error_lowercase", errMsgLower).
			Msg("Error received from tunnel client - checking type for custom error page")

		// More comprehensive connection refused detection
		// Error format from client: "connection_refused: dial tcp [::1]:8084: connect: connection refused"
		// Check for error type "connection_refused" first (most reliable)
		isConnectionRefused := strings.HasPrefix(errMsgLower, "connection_refused:") ||
			strings.Contains(errMsgLower, "connection_refused") ||
			strings.Contains(errMsgLower, "connection refused") ||
			(strings.Contains(errMsgLower, "dial tcp") && strings.Contains(errMsgLower, "connect: connection refused")) ||
			strings.Contains(errMsgLower, "connect: connection refused") ||
			(strings.Contains(errMsgLower, "dial") && strings.Contains(errMsgLower, "refused"))

		if isConnectionRefused {
			ts.logger.Info().
				Str("request_id", requestID).
				Str("error", errMsg).
				Str("subdomain", tunnel.Subdomain).
				Str("local_url", tunnel.LocalURL).
				Msg("Detected connection refused error - writing custom error page")
			ts.writeConnectionRefusedError(w, r, tunnel, errMsg)
			ts.logger.Debug().
				Str("request_id", requestID).
				Msg("Custom connection refused error page written")
			return
		}

		// Return generic error page for other errors
		ts.logger.Info().
			Str("request_id", requestID).
			Str("error", errMsg).
			Str("subdomain", tunnel.Subdomain).
			Int("status_code", http.StatusBadGateway).
			Msg("Writing generic 502 error page")
		ts.writeErrorPage(w, r, tunnel, http.StatusBadGateway, "Bad Gateway", "The tunnel server encountered an error while processing your request.", html.EscapeString(errMsg))
		ts.logger.Debug().
			Str("request_id", requestID).
			Msg("Generic 502 error page written")
		return
	}

	// Write response
	ts.logger.Info().
		Str("request_id", requestID).
		Str("tunnel_id", tunnel.ID).
		Str("subdomain", tunnel.Subdomain).
		Int("status_code", response.Status).
		Int("response_size", len(response.Body)).
		Int("header_count", len(response.Headers)).
		Msg("Writing HTTP response to client")
	
	// Validate response before writing
	if response == nil {
		ts.logger.Error().
			Str("request_id", requestID).
			Str("tunnel_id", tunnel.ID).
			Msg("CRITICAL: Response is nil - cannot write")
		http.Error(w, "Internal server error: response is nil", http.StatusInternalServerError)
		return
	}
	
	ts.writeResponse(w, response)
	
	ts.logger.Debug().
		Str("request_id", requestID).
		Str("tunnel_id", tunnel.ID).
		Int("status_code", response.Status).
		Msg("HTTP response written successfully")

	latency := time.Since(start)
	tunnel.mu.Lock()
	tunnel.RequestCount++
	count := tunnel.RequestCount
	tunnel.mu.Unlock()

	// Log request to database
	if ts.requestLogger != nil {
		reqLog := &TunnelRequestLog{
			TunnelID:        tunnel.ID,
			RequestID:       requestID,
			Method:          r.Method,
			Path:            r.URL.Path,
			QueryString:     r.URL.RawQuery,
			RequestHeaders:  reqData.Headers,
			RequestBody:     reqData.Body,
			StatusCode:      response.Status,
			ResponseHeaders: response.Headers,
			ResponseBody:    response.Body,
			LatencyMs:       int(latency.Milliseconds()),
			RequestSize:     len(reqData.Body),
			ResponseSize:    len(response.Body),
			RemoteAddr:      r.RemoteAddr,
			UserAgent:       r.UserAgent(),
			CreatedAt:       time.Now(),
		}
		ts.requestLogger.LogRequest(r.Context(), reqLog)
	}

	// Record statistics
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

	ts.logger.Info().
		Str("request_id", requestID).
		Str("tunnel_id", tunnel.ID).
		Str("subdomain", tunnel.Subdomain).
		Int("status_code", response.Status).
		Dur("latency", latency).
		Msg("Request forwarded and response written successfully")
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
	// CRITICAL: Headers must be set BEFORE WriteHeader
	// Once WriteHeader is called, headers cannot be modified
	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}

	// Write status code (this commits headers)
	w.WriteHeader(resp.Status)

	// Write body
	if len(resp.Body) > 0 {
		if n, err := w.Write(resp.Body); err != nil {
			ts.logger.Error().
				Err(err).
				Int("bytes_written", n).
				Int("body_size", len(resp.Body)).
				Msg("Failed to write response body")
		}
	}
}

// writeConnectionRefusedError writes a styled error page when local server is not running
func (ts *TunnelServer) writeConnectionRefusedError(w http.ResponseWriter, r *http.Request, tunnel *TunnelConnection, errorMsg string) {
	tunnel.mu.RLock()
	localURL := tunnel.LocalURL
	publicURL := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	tunnel.mu.RUnlock()

	// HTML escape all user inputs to prevent XSS
	publicURL = html.EscapeString(publicURL)
	localURL = html.EscapeString(localURL)

	// Extract port from local URL if available
	localPort := "unknown"
	if strings.HasPrefix(localURL, "http://") {
		parts := strings.Split(localURL[7:], ":")
		if len(parts) > 1 {
			localPort = strings.Split(parts[1], "/")[0]
		}
	} else if strings.Contains(localURL, ":") {
		parts := strings.Split(localURL, ":")
		if len(parts) > 1 {
			localPort = strings.Split(parts[1], "/")[0]
		}
	}
	localPort = html.EscapeString(localPort)

	// Add security headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'none'; img-src 'self' data:; font-src 'self' data:;")

	ts.logger.Debug().
		Str("public_url", publicURL).
		Str("local_url", localURL).
		Str("local_port", localPort).
		Msg("Writing connection refused error page")

	w.WriteHeader(http.StatusBadGateway)

	ts.logger.Debug().
		Str("public_url", publicURL).
		Str("local_url", localURL).
		Int("html_length", 0). // Will update after building HTML
		Msg("Building connection refused error page HTML")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" class="dark">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Connection Refused - UniRoute Tunnel</title>
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
			background: linear-gradient(135deg, #0f172a 0%%, #1e3a8a 50%%, #312e81 100%%);
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
			color: #f1f5f9;
			padding: 20px;
		}
		.container {
			max-width: 600px;
			width: 100%%;
			background: rgba(15, 23, 42, 0.8);
			backdrop-filter: blur(12px);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 16px;
			padding: 40px;
			box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3), 0 10px 10px -5px rgba(0, 0, 0, 0.2);
		}
		.logo {
			display: flex;
			align-items: center;
			justify-content: center;
			margin-bottom: 32px;
		}
		.logo-icon {
			width: 48px;
			height: 48px;
			background: linear-gradient(135deg, #3b82f6 0%%, #6366f1 50%%, #a855f7 100%%);
			border-radius: 12px;
			display: flex;
			align-items: center;
			justify-content: center;
			box-shadow: 0 10px 15px -3px rgba(59, 130, 246, 0.3);
			margin-right: 12px;
		}
		.logo-icon span {
			color: white;
			font-weight: bold;
			font-size: 24px;
		}
		.logo-text {
			font-size: 28px;
			font-weight: bold;
			color: white;
		}
		.error-icon {
			width: 80px;
			height: 80px;
			margin: 0 auto 24px;
			background: rgba(239, 68, 68, 0.1);
			border-radius: 50%%;
			display: flex;
			align-items: center;
			justify-content: center;
			border: 2px solid rgba(239, 68, 68, 0.3);
		}
		.error-icon svg {
			width: 48px;
			height: 48px;
			color: #ef4444;
		}
		h1 {
			font-size: 28px;
			font-weight: bold;
			text-align: center;
			margin-bottom: 12px;
			color: white;
		}
		.subtitle {
			text-align: center;
			color: #cbd5e1;
			margin-bottom: 32px;
			font-size: 16px;
		}
		.info-box {
			background: rgba(30, 41, 59, 0.6);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 12px;
			padding: 24px;
			margin-bottom: 24px;
		}
		.info-row {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: 12px 0;
			border-bottom: 1px solid rgba(148, 163, 184, 0.1);
		}
		.info-row:last-child {
			border-bottom: none;
		}
		.info-label {
			color: #94a3b8;
			font-size: 14px;
			font-weight: 500;
		}
		.info-value {
			color: white;
			font-size: 14px;
			font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
			word-break: break-all;
			text-align: right;
		}
		.steps {
			background: rgba(30, 41, 59, 0.6);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 12px;
			padding: 24px;
			margin-top: 24px;
		}
		.steps h2 {
			font-size: 18px;
			font-weight: 600;
			margin-bottom: 16px;
			color: white;
		}
		.step {
			display: flex;
			align-items: flex-start;
			margin-bottom: 16px;
		}
		.step:last-child {
			margin-bottom: 0;
		}
		.step-number {
			width: 28px;
			height: 28px;
			background: linear-gradient(135deg, #3b82f6 0%%, #6366f1 100%%);
			border-radius: 50%%;
			display: flex;
			align-items: center;
			justify-content: center;
			font-weight: bold;
			font-size: 14px;
			color: white;
			flex-shrink: 0;
			margin-right: 12px;
		}
		.step-content {
			flex: 1;
			color: #cbd5e1;
			font-size: 14px;
			line-height: 1.6;
		}
		.step-content code {
			background: rgba(15, 23, 42, 0.8);
			padding: 2px 6px;
			border-radius: 4px;
			font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
			font-size: 13px;
			color: #60a5fa;
		}
		.footer {
			text-align: center;
			margin-top: 32px;
			color: #64748b;
			font-size: 13px;
		}
		.footer a {
			color: #60a5fa;
			text-decoration: none;
		}
		.footer a:hover {
			text-decoration: underline;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="logo">
			<div class="logo-icon">
				<span>U</span>
			</div>
			<span class="logo-text">UniRoute</span>
		</div>
		
		<div class="error-icon">
			<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
			</svg>
		</div>
		
		<h1>Connection Refused</h1>
		<p class="subtitle">The tunnel is connected, but your local server is not running</p>
		
		<div class="info-box">
			<div class="info-row">
				<span class="info-label">Public URL:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Local URL:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Local Port:</span>
				<span class="info-value">%s</span>
			</div>
		</div>
		
		<div class="steps">
			<h2>How to fix this:</h2>
			<div class="step">
				<div class="step-number">1</div>
				<div class="step-content">Make sure your local server is running on <code>%s</code></div>
			</div>
			<div class="step">
				<div class="step-number">2</div>
				<div class="step-content">Verify the port number matches your application's configuration</div>
			</div>
			<div class="step">
				<div class="step-number">3</div>
				<div class="step-content">Check that your firewall isn't blocking the connection</div>
			</div>
			<div class="step">
				<div class="step-number">4</div>
				<div class="step-content">Once your server is running, refresh this page</div>
			</div>
		</div>
		
		<div class="footer">
			Powered by <a href="https://uniroute.co" target="_blank">UniRoute</a>
		</div>
	</div>
</body>
</html>`, publicURL, localURL, localPort, localURL)

	ts.logger.Debug().
		Int("html_length", len(html)).
		Msg("Writing connection refused error page HTML to response")

	if n, err := w.Write([]byte(html)); err != nil {
		ts.logger.Error().
			Err(err).
			Int("bytes_written", n).
			Msg("Failed to write connection refused error page")
	} else {
		ts.logger.Debug().
			Int("bytes_written", n).
			Msg("Successfully wrote connection refused error page")
	}
}

// writeErrorPage writes a styled error page for various HTTP errors
func (ts *TunnelServer) writeErrorPage(w http.ResponseWriter, r *http.Request, tunnel *TunnelConnection, statusCode int, title, subtitle, details string) {
	var publicURL, localURL string
	if tunnel != nil {
		tunnel.mu.RLock()
		localURL = tunnel.LocalURL
		publicURL = fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
		tunnel.mu.RUnlock()
	} else {
		publicURL = fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
		localURL = "N/A"
	}

	// HTML escape all user inputs to prevent XSS
	publicURL = html.EscapeString(publicURL)
	localURL = html.EscapeString(localURL)
	title = html.EscapeString(title)
	subtitle = html.EscapeString(subtitle)
	details = html.EscapeString(details)

	// Determine icon and color based on status code
	var iconSVG, iconColor string
	switch statusCode {
	case http.StatusNotFound:
		iconSVG = `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>`
		iconColor = "#f59e0b" // amber
	case http.StatusBadGateway:
		iconSVG = `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>`
		iconColor = "#ef4444" // red
	case http.StatusServiceUnavailable:
		iconSVG = `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z"></path>`
		iconColor = "#f59e0b" // amber
	default:
		iconSVG = `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>`
		iconColor = "#ef4444" // red
	}

	// Add security headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'none'; img-src 'self' data:; font-src 'self' data:;")

	ts.logger.Debug().
		Int("status_code", statusCode).
		Str("title", title).
		Str("public_url", publicURL).
		Str("local_url", localURL).
		Msg("Writing error page")

	w.WriteHeader(statusCode)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" class="dark">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s - UniRoute Tunnel</title>
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
			background: linear-gradient(135deg, #0f172a 0%%, #1e3a8a 50%%, #312e81 100%%);
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
			color: #f1f5f9;
			padding: 20px;
		}
		.container {
			max-width: 600px;
			width: 100%%;
			background: rgba(15, 23, 42, 0.8);
			backdrop-filter: blur(12px);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 16px;
			padding: 40px;
			box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3), 0 10px 10px -5px rgba(0, 0, 0, 0.2);
		}
		.logo {
			display: flex;
			align-items: center;
			justify-content: center;
			margin-bottom: 32px;
		}
		.logo-icon {
			width: 48px;
			height: 48px;
			background: linear-gradient(135deg, #3b82f6 0%%, #6366f1 50%%, #a855f7 100%%);
			border-radius: 12px;
			display: flex;
			align-items: center;
			justify-content: center;
			box-shadow: 0 10px 15px -3px rgba(59, 130, 246, 0.3);
			margin-right: 12px;
		}
		.logo-icon span {
			color: white;
			font-weight: bold;
			font-size: 24px;
		}
		.logo-text {
			font-size: 28px;
			font-weight: bold;
			color: white;
		}
		.error-icon {
			width: 80px;
			height: 80px;
			margin: 0 auto 24px;
			background: rgba(239, 68, 68, 0.1);
			border-radius: 50%%;
			display: flex;
			align-items: center;
			justify-content: center;
			border: 2px solid rgba(239, 68, 68, 0.3);
		}
		.error-icon svg {
			width: 48px;
			height: 48px;
			color: %s;
		}
		h1 {
			font-size: 28px;
			font-weight: bold;
			text-align: center;
			margin-bottom: 12px;
			color: white;
		}
		.subtitle {
			text-align: center;
			color: #cbd5e1;
			margin-bottom: 32px;
			font-size: 16px;
		}
		.info-box {
			background: rgba(30, 41, 59, 0.6);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 12px;
			padding: 24px;
			margin-bottom: 24px;
		}
		.info-row {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: 12px 0;
			border-bottom: 1px solid rgba(148, 163, 184, 0.1);
		}
		.info-row:last-child {
			border-bottom: none;
		}
		.info-label {
			color: #94a3b8;
			font-size: 14px;
			font-weight: 500;
		}
		.info-value {
			color: white;
			font-size: 14px;
			font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
			word-break: break-all;
			text-align: right;
		}
		.details {
			background: rgba(30, 41, 59, 0.6);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 12px;
			padding: 16px;
			margin-top: 24px;
		}
		.details-text {
			color: #cbd5e1;
			font-size: 13px;
			font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
			word-break: break-all;
		}
		.footer {
			text-align: center;
			margin-top: 32px;
			color: #64748b;
			font-size: 13px;
		}
		.footer a {
			color: #60a5fa;
			text-decoration: none;
		}
		.footer a:hover {
			text-decoration: underline;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="logo">
			<div class="logo-icon">
				<span>U</span>
			</div>
			<span class="logo-text">UniRoute</span>
		</div>
		
		<div class="error-icon">
			<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
				%s
			</svg>
		</div>
		
		<h1>%s</h1>
		<p class="subtitle">%s</p>
		
		<div class="info-box">
			<div class="info-row">
				<span class="info-label">Public URL:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Local URL:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Status Code:</span>
				<span class="info-value">%d</span>
			</div>
		</div>
		
		<div class="details">
			<div class="details-text">%s</div>
		</div>
		
		<div class="footer">
			Powered by <a href="https://uniroute.co" target="_blank">UniRoute</a>
		</div>
	</div>
</body>
</html>`, title, iconColor, iconSVG, title, subtitle, publicURL, localURL, statusCode, details)

	w.Write([]byte(html))
}

// generateSubdomain generates a random subdomain
func (ts *TunnelServer) generateSubdomain() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:12] // 12 character subdomain
}

// removeTunnel removes a tunnel and updates database status to inactive
// It checks if the tunnel's WebSocket connection matches before removing
// to avoid removing a tunnel that was just resumed with a new connection
func (ts *TunnelServer) removeTunnel(subdomain string) {
	ts.tunnelsMu.Lock()
	tunnel, exists := ts.tunnels[subdomain]
	if !exists {
		ts.tunnelsMu.Unlock()
		return // Tunnel already removed
	}
	
	// Get tunnel details while holding the lock
	tunnel.mu.RLock()
	tunnelID := tunnel.ID
	currentWSConn := tunnel.WSConn
	tunnel.mu.RUnlock()
	
	// Double-check: tunnel might have been re-registered with a new connection
	// Only remove if it's still the same tunnel object
	if ts.tunnels[subdomain] != tunnel {
		ts.logger.Debug().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Msg("Tunnel was re-registered - skipping removal (tunnel was resumed)")
		ts.tunnelsMu.Unlock()
		return // Tunnel was re-registered, don't remove it
	}
	
	// Check if WebSocket connection is still active (not nil)
	// If it's nil, the tunnel might be in a bad state, but we should still remove it
	// If it's not nil, check if it's the same connection we're tracking
	tunnel.mu.RLock()
	wsConnStillActive := tunnel.WSConn != nil && tunnel.WSConn == currentWSConn
	tunnel.mu.RUnlock()
	
	if wsConnStillActive {
		// WebSocket is still active - this might be a premature removal
		// Log a warning but proceed with removal (the connection might be dead)
		ts.logger.Warn().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Msg("Removing tunnel with active WebSocket connection - connection may be dead")
	}
	
	// Remove from map
	delete(ts.tunnels, subdomain)
	ts.tunnelsMu.Unlock()
	
	ts.logger.Debug().
		Str("tunnel_id", tunnelID).
		Str("subdomain", subdomain).
		Bool("ws_conn_nil", currentWSConn == nil).
		Msg("Removed tunnel from memory")

	if exists && tunnel != nil {
		// Remove from port map if it's a TCP/TLS tunnel
		tunnel.mu.RLock()
		protocol := tunnel.Protocol
		tunnelID := tunnel.ID
		tunnel.mu.RUnlock()

		if protocol == ProtocolTCP || protocol == ProtocolTLS {
			ts.portMapMu.Lock()
			for port, t := range ts.portMap {
				if t.ID == tunnelID {
					delete(ts.portMap, port)
					ts.logger.Info().
						Int("port", port).
						Str("tunnel_id", tunnelID).
						Msg("Released TCP port")
					break
				}
			}
			ts.portMapMu.Unlock()
		}

		// Update database status to inactive
		if ts.repository != nil && tunnelID != "" {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := ts.repository.UpdateTunnelStatus(ctx, tunnelID, "inactive"); err != nil {
					ts.logger.Error().
						Err(err).
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Msg("Failed to update tunnel status to inactive")
				} else {
					ts.logger.Info().
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Msg("Updated tunnel status to inactive in database")
				}
			}()
		}

		ts.logger.Info().Str("subdomain", subdomain).Msg("Tunnel removed")
	}
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
	// Generate a proper UUID for tunnel IDs (database expects UUID format)
	return uuid.New().String()
}

// forwardTCPData forwards TCP/TLS data to the appropriate connection
func (ts *TunnelServer) forwardTCPData(tunnel *TunnelConnection, connectionID string, data []byte) {
	ts.tcpConnMu.RLock()
	tcpConn, exists := ts.tcpConnections[connectionID]
	ts.tcpConnMu.RUnlock()

	if !exists {
		ts.logger.Warn().
			Str("connection_id", connectionID).
			Str("tunnel_id", tunnel.ID).
			Msg("TCP connection not found")
		return
	}

	// Write data to the TCP connection
	tcpConn.mu.Lock()
	_, err := tcpConn.Conn.Write(data)
	tcpConn.mu.Unlock()

	if err != nil {
		ts.logger.Error().
			Err(err).
			Str("connection_id", connectionID).
			Msg("Failed to write TCP data")
		// Close connection and remove from map
		ts.closeTCPConnection(connectionID)
	}
}

// handleTCPError handles TCP/TLS errors from tunnel client
func (ts *TunnelServer) handleTCPError(tunnel *TunnelConnection, connectionID string, err *HTTPError) {
	ts.logger.Error().
		Str("connection_id", connectionID).
		Str("error", err.Error).
		Str("message", err.Message).
		Msg("TCP connection error from tunnel")

	// Close the TCP connection
	ts.closeTCPConnection(connectionID)
}

// closeTCPConnection closes and removes a TCP connection
func (ts *TunnelServer) closeTCPConnection(connectionID string) {
	ts.tcpConnMu.Lock()
	defer ts.tcpConnMu.Unlock()

	tcpConn, exists := ts.tcpConnections[connectionID]
	if exists {
		if tcpConn.Conn != nil {
			tcpConn.Conn.Close()
		}
		delete(ts.tcpConnections, connectionID)
		ts.logger.Debug().
			Str("connection_id", connectionID).
			Msg("TCP connection closed")
	}
}

// allocateTCPPort allocates a TCP port for a tunnel and starts listening on it
func (ts *TunnelServer) allocateTCPPort(tunnel *TunnelConnection) int {
	ts.portMapMu.Lock()
	defer ts.portMapMu.Unlock()

	// Find next available port
	maxPort := ts.tcpPortBase + 10000 // Allow up to 10000 TCP tunnels
	startPort := ts.nextTCPPort

	for i := 0; i < 10000; i++ {
		port := (startPort + i) % maxPort
		if port < ts.tcpPortBase {
			port = ts.tcpPortBase + (port % 10000)
		}

		// Check if port is already allocated
		if _, exists := ts.portMap[port]; !exists {
			// Try to listen on this port to ensure it's available
			testListener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				testListener.Close()

				// Allocate port
				ts.portMap[port] = tunnel
				ts.nextTCPPort = port + 1

				// Start listener for this port in background
				go ts.startPortListener(port, tunnel)

				return port
			}
		}
	}

	ts.logger.Error().Msg("Failed to allocate TCP port - no available ports")
	return 0
}

// startPortListener starts a listener on a specific port for a tunnel
func (ts *TunnelServer) startPortListener(port int, tunnel *TunnelConnection) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		ts.logger.Error().
			Err(err).
			Int("port", port).
			Str("tunnel_id", tunnel.ID).
			Msg("Failed to start TCP listener for port")
		return
	}

	ts.logger.Info().
		Int("port", port).
		Str("tunnel_id", tunnel.ID).
		Str("protocol", tunnel.Protocol).
		Msg("TCP listener started for tunnel port")

	// Accept connections on this port
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Check if tunnel still exists
			ts.portMapMu.RLock()
			_, exists := ts.portMap[port]
			ts.portMapMu.RUnlock()

			if !exists {
				ts.logger.Debug().
					Int("port", port).
					Msg("TCP listener closed - tunnel removed")
				listener.Close()
				return
			}

			ts.logger.Error().
				Err(err).
				Int("port", port).
				Msg("Failed to accept TCP connection")
			continue
		}

		// Handle connection
		go ts.handleTCPConnection(tunnel, conn)
	}
}

// handleTCPConnection handles incoming TCP/TLS connections for TCP/TLS tunnels
// This is called when a TCP connection is established to the tunnel server
func (ts *TunnelServer) handleTCPConnection(tunnel *TunnelConnection, conn net.Conn) {
	connectionID := generateID()

	// Store connection
	ts.tcpConnMu.Lock()
	ts.tcpConnections[connectionID] = &TCPConnection{
		ID:        connectionID,
		TunnelID:  tunnel.ID,
		Conn:      conn,
		CreatedAt: time.Now(),
	}
	ts.tcpConnMu.Unlock()

	// Determine message type based on protocol
	tunnel.mu.RLock()
	protocol := tunnel.Protocol
	wsConn := tunnel.WSConn
	tunnel.mu.RUnlock()

	msgType := MsgTypeTCPData
	if protocol == ProtocolTLS {
		msgType = MsgTypeTLSData
	}

	// Send connection request to tunnel client (empty data = new connection)
	msg := TunnelMessage{
		Type:      msgType,
		RequestID: connectionID,
		Data:      []byte{}, // Empty data indicates new connection
	}

	if wsConn == nil {
		ts.logger.Error().
			Str("tunnel_id", tunnel.ID).
			Msg("Tunnel WebSocket connection not available")
		conn.Close()
		ts.closeTCPConnection(connectionID)
		return
	}

	if err := wsConn.WriteJSON(msg); err != nil {
		ts.logger.Error().
			Err(err).
			Str("connection_id", connectionID).
			Msg("Failed to send TCP connection request to tunnel")
		conn.Close()
		ts.closeTCPConnection(connectionID)
		return
	}

	ts.logger.Debug().
		Str("connection_id", connectionID).
		Str("tunnel_id", tunnel.ID).
		Str("protocol", protocol).
		Msg("TCP connection established, forwarding to tunnel client")

	// Read data from TCP connection and forward to tunnel
	go func() {
		defer ts.closeTCPConnection(connectionID)
		buffer := make([]byte, 4096)
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				if err != io.EOF {
					ts.logger.Debug().
						Err(err).
						Str("connection_id", connectionID).
						Msg("TCP connection read error")
				}
				// Send close message to tunnel
				errorType := MsgTypeTCPError
				if protocol == ProtocolTLS {
					errorType = MsgTypeTLSError
				}
				closeMsg := TunnelMessage{
					Type:      errorType,
					RequestID: connectionID,
					Error: &HTTPError{
						RequestID: connectionID,
						Error:     "connection_closed",
						Message:   "TCP connection closed",
					},
				}
				wsConn.WriteJSON(closeMsg)
				return
			}

			if n > 0 {
				// Forward data to tunnel
				dataMsg := TunnelMessage{
					Type:      msgType,
					RequestID: connectionID,
					Data:      buffer[:n],
				}
				if err := wsConn.WriteJSON(dataMsg); err != nil {
					ts.logger.Error().
						Err(err).
						Str("connection_id", connectionID).
						Msg("Failed to forward TCP data to tunnel")
					return
				}
			}
		}
	}()
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
		subdomain := parts[0]
		// Validate subdomain format (alphanumeric and hyphens only, max 63 chars)
		if len(subdomain) > 63 {
			return ""
		}
		matched, _ := regexp.MatchString("^[a-zA-Z0-9-]+$", subdomain)
		if !matched {
			return ""
		}
		return subdomain
	}

	// For domain format like "abc123.uniroute.co"
	// Return first part as subdomain
	subdomain := parts[0]
	// Validate subdomain format (alphanumeric and hyphens only, max 63 chars)
	if len(subdomain) > 63 {
		return ""
	}
	matched, _ := regexp.MatchString("^[a-zA-Z0-9-]+$", subdomain)
	if !matched {
		return ""
	}
	return subdomain
}

// validateLocalURL validates a local URL format
func validateLocalURL(url string) error {
	if len(url) > 2048 {
		return fmt.Errorf("URL too long (max 2048 characters)")
	}

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"..",
		"\x00",
		"javascript:",
		"data:",
		"file:",
	}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(url, pattern) {
			return fmt.Errorf("URL contains invalid pattern: %s", pattern)
		}
	}

	// Validate URL format (must start with http://, https://, or be host:port)
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		// Valid HTTP(S) URL
		return nil
	}

	// Validate host:port format
	if strings.Contains(url, ":") {
		parts := strings.Split(url, ":")
		if len(parts) == 2 {
			// Basic validation - should be host:port
			return nil
		}
	}

	return fmt.Errorf("invalid URL format, must be http://..., https://..., or host:port")
}

// validateSubdomain validates subdomain format
func validateSubdomain(subdomain string) error {
	if len(subdomain) == 0 {
		return fmt.Errorf("subdomain cannot be empty")
	}

	if len(subdomain) > 63 {
		return fmt.Errorf("subdomain too long (max 63 characters)")
	}

	// Must be alphanumeric and hyphens only, cannot start or end with hyphen
	matched, _ := regexp.MatchString("^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$", subdomain)
	if !matched {
		return fmt.Errorf("subdomain must contain only alphanumeric characters and hyphens, and cannot start or end with hyphen")
	}

	return nil
}

// Types are defined in types.go
