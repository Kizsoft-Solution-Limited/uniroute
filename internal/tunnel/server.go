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
	"net/url"
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
	udpConnections  map[string]*UDPConnection // Track active UDP connections
	udpConnMu       sync.RWMutex
	portMap         map[int]*TunnelConnection // Map TCP/UDP port -> tunnel (for TCP/TLS/UDP tunnels)
	portMapMu       sync.RWMutex
	nextTCPPort     int          // Next available TCP port for allocation
	tcpListener     net.Listener // TCP listener for accepting connections
	tcpListenerMu   sync.RWMutex
	udpListeners    map[int]net.PacketConn // UDP listeners by port
	udpListenersMu  sync.RWMutex
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
	jwtValidator    func(tokenString string) (userID string, err error)                 // JWT validator function (optional)
	apiKeyValidator func(ctx context.Context, apiKey string) (userID string, err error) // API key validator function (optional) - DEPRECATED: use apiKeyValidatorWithLimits
	// API key validator that returns rate limits - preferred when available
	apiKeyValidatorWithLimits func(ctx context.Context, apiKey string) (userID string, rateLimitPerMinute, rateLimitPerDay int, err error)
}

// TCPConnection represents an active TCP/TLS connection
type TCPConnection struct {
	ID        string
	TunnelID  string
	Conn      net.Conn
	CreatedAt time.Time
	mu        sync.RWMutex
}

// UDPConnection represents an active UDP connection
type UDPConnection struct {
	ID        string
	TunnelID  string
	Addr      net.Addr // Remote address for UDP
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
	writeMu      sync.Mutex // Mutex to serialize WebSocket writes (WebSocket is not thread-safe for concurrent writes)
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
		udpConnections:  make(map[string]*UDPConnection),
		udpListeners:    make(map[int]net.PacketConn),
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

// SetAPIKeyValidator sets an API key validator function for extracting user ID from API keys
// This allows tunnels created via API key authentication to be automatically associated with users
// DEPRECATED: Use SetAPIKeyValidatorWithLimits to also get rate limits
func (ts *TunnelServer) SetAPIKeyValidator(validator func(ctx context.Context, apiKey string) (userID string, err error)) {
	ts.apiKeyValidator = validator
}

// SetAPIKeyValidatorWithLimits sets an API key validator that returns user ID and rate limits
// This allows tunnels to use the API key's configured rate limits
func (ts *TunnelServer) SetAPIKeyValidatorWithLimits(validator func(ctx context.Context, apiKey string) (userID string, rateLimitPerMinute, rateLimitPerDay int, err error)) {
	ts.apiKeyValidatorWithLimits = validator
}

// Start starts the tunnel server
func (ts *TunnelServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/tunnel", ts.handleTunnelConnection)
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
		Addr:         fmt.Sprintf(":%d", ts.port),
		Handler:      mux,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
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

	// Always require authentication for tunnel operations
	if initMsg.Token == "" {
		ts.logger.Warn().
			Str("protocol", initMsg.Protocol).
			Str("local_url", initMsg.LocalURL).
			Msg("Tunnel connection rejected - authentication required")
		ws.WriteJSON(map[string]string{
			"error":   "authentication required",
			"message": "Please run 'uniroute auth login' to authenticate before creating tunnels",
		})
		ws.Close()
		return
	}

	// Validate token - check if it's an API key or JWT token
	var authenticatedUserID string
	var authErr error
	var apiKeyRateLimitPerMinute, apiKeyRateLimitPerDay int
	var isAPIKey bool // Track if this is an API key for rate limit application

	if strings.HasPrefix(initMsg.Token, "ur_") {
		isAPIKey = true
		if ts.apiKeyValidatorWithLimits != nil {
			authenticatedUserID, apiKeyRateLimitPerMinute, apiKeyRateLimitPerDay, authErr = ts.apiKeyValidatorWithLimits(r.Context(), initMsg.Token)
			if authErr == nil && authenticatedUserID != "" {
				ts.logger.Info().
					Str("user_id", authenticatedUserID).
					Int("rate_limit_per_minute", apiKeyRateLimitPerMinute).
					Int("rate_limit_per_day", apiKeyRateLimitPerDay).
					Msg("API key validated with rate limits")
			} else {
				ts.logger.Warn().
					Err(authErr).
					Str("user_id", authenticatedUserID).
					Msg("API key validation failed")
			}
		} else if ts.apiKeyValidator != nil {
			authenticatedUserID, authErr = ts.apiKeyValidator(r.Context(), initMsg.Token)
		} else {
			ts.logger.Warn().Msg("API key validator not configured")
			ws.WriteJSON(map[string]string{
				"error":   "authentication not configured",
				"message": "Tunnel server requires API key validation but API key validator is not configured",
			})
			ws.Close()
			return
		}
	} else {
		if ts.jwtValidator == nil {
			ts.logger.Warn().Msg("JWT validator not configured")
			ws.WriteJSON(map[string]string{
				"error":   "authentication not configured",
				"message": "Tunnel server requires JWT validation but JWT validator is not configured",
			})
			ws.Close()
			return
		}
		authenticatedUserID, authErr = ts.jwtValidator(initMsg.Token)
	}

	if authErr != nil || authenticatedUserID == "" {
		ts.logger.Warn().
			Err(authErr).
			Msg("Tunnel connection rejected - invalid or expired token")
		ws.WriteJSON(map[string]string{
			"error":   "invalid or expired token",
			"message": "Please run 'uniroute auth login' to authenticate again",
		})
		ws.Close()
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
	if protocol != ProtocolHTTP && protocol != ProtocolTCP && protocol != ProtocolTLS && protocol != ProtocolUDP {
		ts.logger.Error().Str("protocol", protocol).Msg("Invalid protocol")
		ws.WriteJSON(map[string]string{"error": "invalid protocol, must be http, tcp, tls, or udp"})
		ws.Close()
		return
	}

	// Check if client wants to resume existing tunnel
	var subdomain string
	var tunnelID string
	var isResume bool
	var autoFoundTunnel bool // Track if tunnel was auto-found (to preserve isResume flag)
	// authenticatedUserID is already extracted above and will be used in resume logic

	if initMsg.ForceNew {
		isResume = false
		subdomain = ""
		tunnelID = ""
	} else if initMsg.Subdomain == "" && initMsg.TunnelID == "" && initMsg.Host == "" && authenticatedUserID != "" && ts.repository != nil {

		userUUID, parseErr := uuid.Parse(authenticatedUserID)
		if parseErr == nil {
			userTunnels, err := ts.repository.ListTunnelsByUser(r.Context(), userUUID)
			if err == nil && len(userTunnels) > 0 {
				// Query already returns tunnels ordered by: active first, then most recent (created_at DESC)
				// So the first tunnel in the list is already the best one to resume
				// This works for all protocols (HTTP, TCP, TLS, UDP) - protocol is not filtered
				// The database query uses: ORDER BY CASE WHEN status = 'active' THEN 0 ELSE 1 END, created_at DESC
				bestTunnel := userTunnels[0] // First tunnel is already the best (active first, then most recent)

				if bestTunnel != nil && bestTunnel.UserID == authenticatedUserID {
					subdomain = bestTunnel.Subdomain
					tunnelID = bestTunnel.ID
					isResume = true
					autoFoundTunnel = true
				} else if bestTunnel != nil {
					ts.logger.Warn().
						Str("tunnel_id", bestTunnel.ID).
						Str("tunnel_user_id", bestTunnel.UserID).
						Str("authenticated_user_id", authenticatedUserID).
						Msg("Found tunnel but user_id mismatch - will create new tunnel")
				}
			} else if err != nil {
				ts.logger.Debug().
					Err(err).
					Str("user_id", authenticatedUserID).
					Msg("Failed to list tunnels for user")
			} else {
				ts.logger.Debug().
					Str("user_id", authenticatedUserID).
					Msg("No existing tunnels found for user - will create new tunnel")
			}
		} else {
			ts.logger.Debug().
				Err(parseErr).
				Str("user_id", authenticatedUserID).
				Msg("Failed to parse user_id as UUID - cannot look up user tunnels")
		}
	}

	if autoFoundTunnel && isResume && subdomain != "" && tunnelID != "" {
	} else if initMsg.Subdomain != "" || initMsg.TunnelID != "" {
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
						Str("user_id", dbTunnel.UserID).
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
							Str("user_id", dbTunnel.UserID).
							Msg("Found tunnel in database by ID")
					}
				} else {
					// Tunnel ID is not a valid UUID (might be hex-formatted old ID)
					// Try to extract subdomain from the hex ID and look up by subdomain
					// Hex IDs are typically 12 characters, and subdomains are derived from them
					ts.logger.Debug().
						Err(parseErr).
						Str("tunnel_id", initMsg.TunnelID).
						Msg("Failed to parse tunnel ID as UUID - trying to look up by subdomain")

					// If the tunnel ID looks like a hex ID (12 chars), try to use it as subdomain
					if len(initMsg.TunnelID) == 12 {
						ts.logger.Debug().
							Str("tunnel_id", initMsg.TunnelID).
							Str("subdomain", initMsg.TunnelID).
							Msg("Treating hex tunnel ID as subdomain for lookup")
						dbTunnel, err = ts.repository.GetTunnelBySubdomain(context.Background(), initMsg.TunnelID)
						if err != nil {
							ts.logger.Debug().
								Err(err).
								Str("tunnel_id", initMsg.TunnelID).
								Msg("Tunnel not found by subdomain (from hex ID)")
						} else if dbTunnel != nil {
							ts.logger.Debug().
								Str("tunnel_id", dbTunnel.ID).
								Str("subdomain", dbTunnel.Subdomain).
								Str("status", dbTunnel.Status).
								Str("user_id", dbTunnel.UserID).
								Msg("Found tunnel in database by subdomain (from hex ID)")
						}
					} else {
						err = parseErr
						ts.logger.Debug().
							Err(parseErr).
							Str("tunnel_id", initMsg.TunnelID).
							Msg("Tunnel ID is not a valid UUID and not a 12-char hex ID - cannot look up")
					}
				}
			}

			if err == nil && dbTunnel != nil {
				// authenticatedUserID was already validated and extracted above (line 289)
				// Skip tunnels with null user_id or belonging to other users
				// authenticatedUserID is in scope from the auth check above
				if authenticatedUserID != "" {
					if dbTunnel.UserID == "" || dbTunnel.UserID == "null" {
						ts.logger.Warn().
							Str("tunnel_id", dbTunnel.ID).
							Str("subdomain", dbTunnel.Subdomain).
							Str("authenticated_user_id", authenticatedUserID).
							Str("db_user_id", dbTunnel.UserID).
							Msg("Tunnel has null/empty user_id - REJECTING resume (will create new tunnel)")
						// Don't set resume variables - will create new tunnel below
						dbTunnel = nil
						isResume = false
					} else if dbTunnel.UserID != authenticatedUserID {
						ts.logger.Info().
							Str("tunnel_id", dbTunnel.ID).
							Str("subdomain", dbTunnel.Subdomain).
							Str("tunnel_user_id", dbTunnel.UserID).
							Str("authenticated_user_id", authenticatedUserID).
							Msg("Tunnel belongs to different user - skipping resume (will create new tunnel)")
						// Don't set resume variables - will create new tunnel below
						dbTunnel = nil
						isResume = false
					} else {
						// Tunnel belongs to authenticated user - proceed with resume
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
							Str("user_id", dbTunnel.UserID).
							Str("requested_subdomain", initMsg.Subdomain).
							Str("requested_tunnel_id", initMsg.TunnelID).
							Str("database_subdomain", dbTunnel.Subdomain).
							Msg("Resuming existing tunnel (from database) - tunnel belongs to authenticated user")
					}
				} else {
					// This should not happen since auth is required, but handle it gracefully
					ts.logger.Warn().
						Str("tunnel_id", dbTunnel.ID).
						Str("subdomain", dbTunnel.Subdomain).
						Msg("Tunnel found but user is not authenticated - skipping resume (will create new tunnel)")
					// Don't set resume variables - will create new tunnel below
					dbTunnel = nil
				}
			}

			// If dbTunnel was set to nil (skipped), ensure we don't resume
			// UNLESS it was auto-found and validated (in which case we trust the auto-find)
			if dbTunnel == nil {
				// If tunnel was auto-found and validated, preserve isResume = true
				// The auto-find already validated the tunnel belongs to the user
				if autoFoundTunnel && isResume && subdomain != "" && tunnelID != "" {
					// Tunnel was auto-found and validated - keep isResume = true
					// The lookup might have failed due to timing or other issues, but we trust the auto-find
					ts.logger.Info().
						Str("subdomain", subdomain).
						Str("tunnel_id", tunnelID).
						Str("authenticated_user_id", authenticatedUserID).
						Msg("Auto-found tunnel validated - preserving resume flag even though lookup didn't find it (will resume auto-found tunnel)")
					// Keep isResume = true, subdomain, and tunnelID as set by auto-find
				} else {
					// Tunnel was not auto-found, or was rejected - create new tunnel
					isResume = false
					// Only clear subdomain/tunnelID if they weren't set by auto-find
					if !autoFoundTunnel {
						subdomain = ""
						tunnelID = ""
					}
					ts.logger.Info().
						Str("requested_subdomain", initMsg.Subdomain).
						Str("requested_tunnel_id", initMsg.TunnelID).
						Str("authenticated_user_id", authenticatedUserID).
						Bool("auto_found", autoFoundTunnel).
						Msg("No tunnel found with matching user_id - creating new tunnel")
				}
			} else {
				// Tunnel not found in memory or database, create new one
				if err != nil {
					ts.logger.Info().
						Err(err).
						Str("requested_subdomain", initMsg.Subdomain).
						Str("requested_tunnel_id", initMsg.TunnelID).
						Str("authenticated_user_id", authenticatedUserID).
						Msg("Tunnel not found in database - creating new tunnel")
				} else {
					ts.logger.Info().
						Str("requested_subdomain", initMsg.Subdomain).
						Str("requested_tunnel_id", initMsg.TunnelID).
						Str("authenticated_user_id", authenticatedUserID).
						Msg("No tunnel found in memory or database - creating new tunnel")
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

	// Log the current state before deciding to create new tunnel or resume
	if autoFoundTunnel {
		ts.logger.Info().
			Bool("is_resume", isResume).
			Str("subdomain", subdomain).
			Str("tunnel_id", tunnelID).
			Str("authenticated_user_id", authenticatedUserID).
			Msg("Auto-found tunnel state - proceeding to resume logic")
	}

	// Create new tunnel if not resuming
	if !isResume {
		// Check if client requested a specific subdomain via Host field
		var err error
		if initMsg.Host != "" {
			// Validate requested subdomain
			if err := validateSubdomain(initMsg.Host); err != nil {
				ts.logger.Warn().
					Err(err).
					Str("requested_host", initMsg.Host).
					Msg("Invalid subdomain requested")
				ws.WriteJSON(map[string]string{"error": "invalid subdomain: " + err.Error()})
				ws.Close()
				return
			}

			// If ForceNew is set, we'll create a new tunnel with this subdomain
			// But we still need to check if it's available (unless it's the user's own tunnel)
			if !initMsg.ForceNew {
				// Check if subdomain is available
				available := true
				if ts.domainManager != nil && ts.repository != nil {
					available, err = ts.domainManager.CheckSubdomainAvailability(context.Background(), ts.repository, initMsg.Host)
					if err != nil {
						ts.logger.Error().Err(err).Str("requested_host", initMsg.Host).Msg("Failed to check subdomain availability")
						ws.WriteJSON(map[string]string{"error": "failed to check subdomain availability"})
						ws.Close()
						return
					}
				}

				// Also check if subdomain is already in use by an active tunnel
				ts.tunnelsMu.RLock()
				_, existsInMemory := ts.tunnels[initMsg.Host]
				ts.tunnelsMu.RUnlock()

				if !available || existsInMemory {
					ts.logger.Warn().
						Str("requested_host", initMsg.Host).
						Bool("available_in_db", available).
						Bool("exists_in_memory", existsInMemory).
						Msg("Requested subdomain is not available")
					ws.WriteJSON(map[string]string{"error": "subdomain '" + initMsg.Host + "' is not available"})
					ws.Close()
					return
				}
			} else {
				// ForceNew is set - we'll create a new tunnel, but if the subdomain exists and belongs to this user,
				// we might want to reuse it. For now, we'll create a new tunnel with a different subdomain if needed.
				// But actually, with ForceNew, the user wants a NEW tunnel, so we should allow reusing their own subdomain
				ts.logger.Info().
					Str("requested_host", initMsg.Host).
					Bool("force_new", initMsg.ForceNew).
					Msg("ForceNew is set - will create new tunnel with requested subdomain (may replace existing)")
			}

			// Use requested subdomain
			subdomain = initMsg.Host
			ts.logger.Info().
				Str("requested_subdomain", subdomain).
				Bool("force_new", initMsg.ForceNew).
				Msg("Using requested subdomain")
		} else {
			// Allocate random subdomain using domain manager
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
		}
		// Always generate a new UUID for new tunnels
		tunnelID = generateID()
	} else {
		// For resume, validate that tunnelID is a valid UUID
		// If client sent an old hex-format ID, we need to use the one from database
		if tunnelID != "" {
			if _, err := uuid.Parse(tunnelID); err != nil {
				// Only try to fix invalid IDs if tunnel was NOT auto-found
				// Auto-found tunnels should already have valid UUIDs from database
				if !autoFoundTunnel {
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
				} else {
					// Auto-found tunnel with invalid ID - this shouldn't happen, but log it
					ts.logger.Error().
						Err(err).
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Msg("CRITICAL: Auto-found tunnel has invalid UUID - this should not happen")
					// Don't clear isResume for auto-found tunnels - trust the auto-find
				}
			}
		} else if autoFoundTunnel {
			// Auto-found tunnel but tunnelID is empty - this shouldn't happen
			ts.logger.Error().
				Str("subdomain", subdomain).
				Msg("CRITICAL: Auto-found tunnel but tunnelID is empty - this should not happen")
			// Don't clear isResume - trust the auto-find, it will be handled in resume logic
		}
	}

	// Create or update tunnel connection
	var tunnel *TunnelConnection
	if subdomain == "" {
		ts.logger.Error().Msg("CRITICAL: subdomain is empty - cannot create tunnel")
		ws.WriteJSON(map[string]string{"error": "subdomain is required"})
		return
	}

	if isResume {
		// SIMPLIFIED: Always create a fresh tunnel connection for resume, just like new tunnel
		// This ensures resume works exactly the same way as new tunnel creation
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

		// Configure rate limits from API key BEFORE registering tunnel
		// This ensures rate limits are set before any requests can come in
		// ALWAYS apply rate limits if API key was used (even if values are 0, we should still set them)
		if isAPIKey {
			if apiKeyRateLimitPerMinute > 0 || apiKeyRateLimitPerDay > 0 {
				// Calculate hourly limit from daily (more accurate than dividing by 24)
				// If daily is very high, set a reasonable hourly limit
				hourlyLimit := apiKeyRateLimitPerDay / 24
				if hourlyLimit < apiKeyRateLimitPerMinute*60 {
					// Hourly should be at least 60x the per-minute limit
					hourlyLimit = apiKeyRateLimitPerMinute * 60
				}

				rateLimitConfig := &RateLimitConfig{
					RequestsPerMinute: apiKeyRateLimitPerMinute,
					RequestsPerHour:   hourlyLimit,
					RequestsPerDay:    apiKeyRateLimitPerDay,
					BurstSize:         50, // Higher burst for high-limit API keys
				}
				ts.rateLimiter.SetRateLimit(tunnelID, rateLimitConfig)
			}
		}

		ts.tunnelsMu.Lock()
		existingTunnel := ts.tunnels[subdomain]
		ts.tunnels[subdomain] = tunnel
		registeredCount := len(ts.tunnels)
		ts.tunnelsMu.Unlock()

		if existingTunnel != nil && existingTunnel != tunnel {
			existingTunnel.mu.RLock()
			oldWSConn := existingTunnel.WSConn
			existingTunnel.mu.RUnlock()

			existingTunnel.mu.Lock()
			existingTunnel.WSConn = nil
			existingTunnel.mu.Unlock()
			if oldWSConn != nil && oldWSConn != ws {
				go func() {
					// Give old handler a moment to detect the replacement
					time.Sleep(50 * time.Millisecond)
					oldWSConn.Close()
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

		// Allocate UDP port for UDP tunnels if needed
		if protocol == ProtocolUDP {
			udpPort := ts.allocateUDPPort(tunnel)
			if udpPort > 0 {
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("protocol", protocol).
					Int("udp_port", udpPort).
					Msg("Allocated UDP port for resumed tunnel")
			}
		}

		// Update database: mark as active, update LocalURL, and update last_active_at when resuming
		if ts.repository != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// Get existing tunnel from database to check user_id
				// Try to parse as UUID first, if that fails, try to find by subdomain
				var existingTunnel *Tunnel
				tunnelUUID, parseErr := uuid.Parse(tunnelID)
				if parseErr == nil {
					existingTunnel, _ = ts.repository.GetTunnelByID(ctx, tunnelUUID)
				} else {
					// Tunnel ID might be in hex format, try to find by subdomain instead
					ts.logger.Debug().
						Err(parseErr).
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Msg("Tunnel ID is not a valid UUID, trying to find tunnel by subdomain")
					if subdomain != "" {
						existingTunnel, _ = ts.repository.GetTunnelBySubdomain(ctx, subdomain)
					}
				}

				// Update LocalURL first (in case it changed, e.g., different port)
				if err := ts.repository.UpdateTunnelLocalURL(ctx, tunnelID, initMsg.LocalURL); err != nil {
					ts.logger.Error().
						Err(err).
						Str("tunnel_id", tunnelID).
						Str("local_url", initMsg.LocalURL).
						Msg("Failed to update tunnel LocalURL on resume")
				}

				// If tunnel doesn't have user_id and token is provided, try to associate it
				if existingTunnel != nil && existingTunnel.UserID == "" && initMsg.Token != "" {
					if ts.jwtValidator != nil {
						if extractedUserID, err := ts.jwtValidator(initMsg.Token); err == nil && extractedUserID != "" {
							if userUUID, parseErr := uuid.Parse(extractedUserID); parseErr == nil {
								// Use the tunnel ID from the database (which is a valid UUID)
								associateTunnelID := existingTunnel.ID
								if err := ts.repository.AssociateTunnelWithUser(ctx, uuid.MustParse(associateTunnelID), userUUID); err != nil {
									ts.logger.Warn().
										Err(err).
										Str("tunnel_id", tunnelID).
										Str("db_tunnel_id", associateTunnelID).
										Str("user_id", extractedUserID).
										Msg("Failed to associate tunnel with user on resume")
								} else {
									ts.logger.Info().
										Str("tunnel_id", tunnelID).
										Str("db_tunnel_id", associateTunnelID).
										Str("subdomain", subdomain).
										Str("user_id", extractedUserID).
										Msg("Associated tunnel with user on resume (tunnel was created without user_id)")
								}
							}
						} else {
							ts.logger.Warn().
								Err(err).
								Str("tunnel_id", tunnelID).
								Msg("Failed to extract user ID from token when trying to associate tunnel on resume")
						}
					} else {
						ts.logger.Debug().
							Str("tunnel_id", tunnelID).
							Msg("Token provided but JWT validator not configured - cannot associate tunnel with user on resume")
					}
				} else if existingTunnel != nil && existingTunnel.UserID != "" {
					ts.logger.Debug().
						Str("tunnel_id", tunnelID).
						Str("user_id", existingTunnel.UserID).
						Msg("Tunnel already has user_id - preserving it on resume")
				} else if existingTunnel == nil {
					ts.logger.Warn().
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Msg("Could not find tunnel in database to check user_id - association skipped")
				}

				// Update status to active (this preserves existing user_id)
				if err := ts.repository.UpdateTunnelStatus(ctx, tunnelID, "active"); err != nil {
					ts.logger.Error().
						Err(err).
						Str("tunnel_id", tunnelID).
						Msg("Failed to update tunnel status to active on resume from database")
				} else {
					// Verify user_id and custom domain are preserved after status update
					if existingTunnel != nil {
						logEntry := ts.logger.Info().
							Str("tunnel_id", tunnelID).
							Str("subdomain", subdomain).
							Str("user_id", existingTunnel.UserID)

						if existingTunnel.CustomDomain != "" {
							logEntry = logEntry.Str("custom_domain", existingTunnel.CustomDomain)
						}

						logEntry.Msg("Updated tunnel status to active - user_id and custom domain preserved")
					} else {
						ts.logger.Info().
							Str("tunnel_id", tunnelID).
							Str("subdomain", subdomain).
							Msg("Updated tunnel status to active")
					}
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

		// Configure rate limits from API key BEFORE registering tunnel
		// This ensures rate limits are set before any requests can come in
		// ALWAYS apply rate limits if API key was used (even if values are 0, we should still set them)
		if isAPIKey {
			if apiKeyRateLimitPerMinute > 0 || apiKeyRateLimitPerDay > 0 {
				// Calculate hourly limit from daily (more accurate than dividing by 24)
				// If daily is very high, set a reasonable hourly limit
				hourlyLimit := apiKeyRateLimitPerDay / 24
				if hourlyLimit < apiKeyRateLimitPerMinute*60 {
					// Hourly should be at least 60x the per-minute limit
					hourlyLimit = apiKeyRateLimitPerMinute * 60
				}

				rateLimitConfig := &RateLimitConfig{
					RequestsPerMinute: apiKeyRateLimitPerMinute,
					RequestsPerHour:   hourlyLimit,
					RequestsPerDay:    apiKeyRateLimitPerDay,
					BurstSize:         50, // Higher burst for high-limit API keys
				}
				ts.rateLimiter.SetRateLimit(tunnelID, rateLimitConfig)
			}
		}
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

		if isAPIKey {
			if apiKeyRateLimitPerMinute > 0 || apiKeyRateLimitPerDay > 0 {
				hourlyLimit := apiKeyRateLimitPerDay / 24
				if hourlyLimit < apiKeyRateLimitPerMinute*60 {
					hourlyLimit = apiKeyRateLimitPerMinute * 60
				}

				rateLimitConfig := &RateLimitConfig{
					RequestsPerMinute: apiKeyRateLimitPerMinute,
					RequestsPerHour:   hourlyLimit,
					RequestsPerDay:    apiKeyRateLimitPerDay,
					BurstSize:         50,
				}
				ts.rateLimiter.SetRateLimit(tunnelID, rateLimitConfig)
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
	} else if tunnelProtocol == ProtocolUDP {
		// Get allocated UDP port
		ts.portMapMu.RLock()
		var udpPort int
		for port, t := range ts.portMap {
			if t.ID == tunnelID {
				udpPort = port
				break
			}
		}
		ts.portMapMu.RUnlock()

		if udpPort > 0 {
			// For UDP, public URL is host:port format
			if ts.domainManager != nil && ts.domainManager.baseDomain != "" {
				publicURL = fmt.Sprintf("%s:%d", ts.domainManager.GetPublicURL(subdomain, ts.port, false), udpPort)
			} else {
				publicURL = fmt.Sprintf("%s.localhost:%d", subdomain, udpPort)
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
		// Use authenticatedUserID that was already validated above (works for both JWT and API keys)
		var userID string
		if authenticatedUserID != "" {
			// authenticatedUserID was already extracted and validated during authentication
			// This works for both JWT tokens and API keys
			userID = authenticatedUserID
			ts.logger.Info().
				Str("tunnel_id", tunnelID).
				Str("user_id", userID).
				Str("subdomain", subdomain).
				Bool("is_api_key", strings.HasPrefix(initMsg.Token, "ur_")).
				Msg("Using authenticated user ID - tunnel will be associated with user")
		} else {
			// authenticatedUserID should always be set if authentication succeeded
			// This is a fallback in case something went wrong
			ts.logger.Warn().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Str("has_token", fmt.Sprintf("%v", initMsg.Token != "")).
				Msg("authenticatedUserID is empty but creating tunnel - attempting to extract from token")

			if initMsg.Token != "" {
				// Fallback: try to extract from token if authenticatedUserID wasn't set (shouldn't happen)
				if ts.jwtValidator != nil {
					if extractedUserID, err := ts.jwtValidator(initMsg.Token); err == nil && extractedUserID != "" {
						userID = extractedUserID
						ts.logger.Info().
							Str("tunnel_id", tunnelID).
							Str("user_id", userID).
							Msg("Extracted user ID from JWT token - tunnel will be associated with user")
					} else {
						ts.logger.Warn().
							Err(err).
							Str("tunnel_id", tunnelID).
							Msg("Failed to extract user ID from JWT token (tunnel will be unassociated)")
					}
				} else if ts.apiKeyValidator != nil && strings.HasPrefix(initMsg.Token, "ur_") {
					// Try API key validator as fallback
					if extractedUserID, err := ts.apiKeyValidator(r.Context(), initMsg.Token); err == nil && extractedUserID != "" {
						userID = extractedUserID
						ts.logger.Info().
							Str("tunnel_id", tunnelID).
							Str("user_id", userID).
							Msg("Extracted user ID from API key - tunnel will be associated with user")
					} else {
						ts.logger.Warn().
							Err(err).
							Str("tunnel_id", tunnelID).
							Msg("Failed to extract user ID from API key (tunnel will be unassociated)")
					}
				} else {
					ts.logger.Warn().
						Str("tunnel_id", tunnelID).
						Msg("Auth token provided but validators not configured - tunnel will be unassociated. Set JWT_SECRET or API_KEY_SECRET in tunnel server config.")
				}
			} else {
				ts.logger.Warn().
					Str("tunnel_id", tunnelID).
					Str("subdomain", subdomain).
					Msg("No auth token provided and authenticatedUserID is empty - tunnel will be unassociated")
			}
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

	// The handler starts its message loop immediately, so it's ready as soon as goroutine starts
	go ts.handleTunnelMessages(tunnel)

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
					tunnel.mu.RUnlock()

					// Update last_active_at in database (don't change request count on ping)
					if err := ts.repository.UpdateTunnelActivity(ctx, tunnelID, 0); err != nil {
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
			tunnel.mu.RLock()
			wsConn := tunnel.WSConn
			tunnel.mu.RUnlock()
			if wsConn != nil {
				pongMsg := TunnelMessage{Type: MsgTypePong}
				tunnel.writeMu.Lock()
				tunnel.mu.RLock()
				currentWSConn := tunnel.WSConn
				tunnel.mu.RUnlock()
				if currentWSConn == wsConn && currentWSConn != nil {
					if err := currentWSConn.WriteJSON(pongMsg); err != nil {
						ts.logger.Warn().Err(err).Msg("Failed to send pong response")
					}
				}
				tunnel.writeMu.Unlock()
			}
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
		case MsgTypeUDPData:
			// Forward UDP data to remote address
			if msg.RequestID != "" {
				ts.forwardUDPData(tunnel, msg.RequestID, msg.Data)
			}
		case MsgTypeUDPError:
			// Handle UDP errors
			if msg.Error != nil {
				ts.handleUDPError(tunnel, msg.RequestID, msg.Error)
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
	if r.URL.Path == "/tunnel" {
		return
	}

	path := r.URL.Path
	if path == "/favicon.ico" || path == "/favicon.png" {
		http.NotFound(w, r)
		return
	}

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

	// Extract subdomain or custom domain from Host header
	host := r.Host
	// Remove port if present
	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		hostname = host
	}

	subdomain := extractSubdomain(host)
	var tunnel *TunnelConnection
	var exists bool
	var lookupSubdomain string

	ts.logger.Debug().
		Str("host", host).
		Str("hostname", hostname).
		Str("extracted_subdomain", subdomain).
		Msg("Extracted subdomain from host header")

	// First, try to find tunnel by subdomain (existing behavior)
	if subdomain != "" {
		ts.tunnelsMu.RLock()
		tunnel, exists = ts.tunnels[subdomain]
		ts.tunnelsMu.RUnlock()
		lookupSubdomain = subdomain
	}

	// If not found by subdomain, check if it's a custom domain
	if !exists && ts.repository != nil {
		dbTunnel, err := ts.repository.GetTunnelByCustomDomain(context.Background(), hostname)
		if err == nil && dbTunnel != nil {
			// Found tunnel by custom domain - use its subdomain to find active connection
			lookupSubdomain = dbTunnel.Subdomain
			ts.tunnelsMu.RLock()
			tunnel, exists = ts.tunnels[lookupSubdomain]
			ts.tunnelsMu.RUnlock()

			ts.logger.Info().
				Str("custom_domain", hostname).
				Str("tunnel_subdomain", lookupSubdomain).
				Bool("tunnel_active", exists).
				Msg("Found tunnel by custom domain")
		}
	}

	if subdomain == "" && !exists {
		// No subdomain and no custom domain match, return info page or 404
		ts.logger.Debug().Str("host", host).Msg("No subdomain or custom domain found, showing root page")
		ts.handleRootRequest(w, r)
		return
	}
	// Log all available tunnels for debugging
	ts.tunnelsMu.RLock()
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
		Str("subdomain", lookupSubdomain).
		Str("host", host).
		Bool("exists_in_memory", exists).
		Strs("available_tunnels", availableSubdomains).
		Interface("tunnel_details", tunnelDetails).
		Str("request_path", r.URL.Path).
		Str("request_method", r.Method).
		Str("request_host", r.Host).
		Bool("has_repository", ts.repository != nil).
		Msg("HTTP request received - checking tunnel in memory")

	if exists && tunnel != nil {
		tunnel.mu.RLock()
		tunnelID := tunnel.ID
		tunnelProtocol := tunnel.Protocol
		tunnel.mu.RUnlock()

		// Only HTTP tunnels are handled through HTTP requests
		// TCP/TLS/UDP tunnels are handled through their respective listeners
		if tunnelProtocol != ProtocolHTTP {
			ts.logger.Warn().
				Str("subdomain", lookupSubdomain).
				Str("tunnel_id", tunnelID).
				Str("protocol", tunnelProtocol).
				Msg("HTTP request received for non-HTTP tunnel - protocol mismatch")
			ts.writeErrorPage(w, r, tunnel, http.StatusBadRequest, "Protocol Mismatch", "This tunnel is not an HTTP tunnel.", fmt.Sprintf("The tunnel '%s' is configured for %s protocol, not HTTP.", html.EscapeString(lookupSubdomain), tunnelProtocol))
			return
		}

		// Check if tunnel is actually connected (WebSocket must be active)
		// Re-check connection state to avoid race conditions
		tunnel.mu.RLock()
		wsConn := tunnel.WSConn
		tunnel.mu.RUnlock()

		if wsConn == nil {
			ts.logger.Warn().
				Str("subdomain", lookupSubdomain).
				Str("tunnel_id", tunnelID).
				Msg("Tunnel exists but WebSocket connection is nil - tunnel is disconnected")
			// Check database to see if tunnel exists but is disconnected
			if ts.repository != nil {
				dbTunnel, err := ts.repository.GetTunnelBySubdomain(context.Background(), lookupSubdomain)
				if err == nil && dbTunnel != nil {
					ts.writeErrorPage(w, r, nil, http.StatusServiceUnavailable, "The endpoint is offline", "The tunnel exists but is not currently connected.", fmt.Sprintf("The endpoint '%s' is associated with a tunnel, but the tunnel client is not connected. Please start the tunnel client to resume this tunnel.", html.EscapeString(lookupSubdomain)))
					return
				}
			}
			// If no database or not found, show 404
			ts.writeErrorPage(w, r, nil, http.StatusNotFound, "Tunnel Not Found", "The requested tunnel does not exist.", fmt.Sprintf("The subdomain '%s' is not associated with any tunnel.", html.EscapeString(lookupSubdomain)))
			return
		} else {
			// Check if handler is ready (for resumed tunnels, handler might not be ready yet)
			tunnel.mu.RLock()
			handlerReady := tunnel.handlerReady
			tunnel.mu.RUnlock()

			if !handlerReady {
				ts.logger.Debug().
					Str("tunnel_id", tunnel.ID).
					Str("subdomain", lookupSubdomain).
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
						Str("subdomain", lookupSubdomain).
						Msg("Tunnel handler still not ready after wait - proceeding anyway (handler should be ready soon)")
					// Continue anyway - handler should be ready very soon
					// The request will be queued in the request tracker and handled when response arrives
				} else {
					ts.logger.Debug().
						Str("tunnel_id", tunnel.ID).
						Str("subdomain", lookupSubdomain).
						Msg("Tunnel handler is ready - proceeding with request")
				}
			}
		}
	}

	if !exists {
		// Tunnel not in memory - check database to see if it exists but is disconnected
		if ts.repository != nil {
			// First, wait longer in case resume is in progress (max 1 second)
			// This is important because resume can take a moment to register the tunnel
			for i := 0; i < 100; i++ {
				time.Sleep(10 * time.Millisecond)
				ts.tunnelsMu.RLock()
				tunnel, exists = ts.tunnels[lookupSubdomain]
				ts.tunnelsMu.RUnlock()
				if exists && tunnel != nil {
					ts.logger.Info().
						Str("subdomain", lookupSubdomain).
						Int("wait_iterations", i+1).
						Msg("Tunnel appeared in memory during wait - resume likely in progress")

					tunnel.mu.RLock()
					handlerReady := tunnel.handlerReady
					wsConn := tunnel.WSConn
					tunnel.mu.RUnlock()

					if wsConn != nil && !handlerReady {
						ts.logger.Debug().
							Str("subdomain", lookupSubdomain).
							Msg("Tunnel found but handler not ready - waiting for handler")
						// Wait up to 500ms more for handler to be ready
						for j := 0; j < 50; j++ {
							time.Sleep(10 * time.Millisecond)
							tunnel.mu.RLock()
							handlerReady = tunnel.handlerReady
							tunnel.mu.RUnlock()
							if handlerReady {
								ts.logger.Info().
									Str("subdomain", lookupSubdomain).
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
				dbTunnel, err := ts.repository.GetTunnelBySubdomain(context.Background(), lookupSubdomain)
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
					ts.writeErrorPage(w, r, nil, http.StatusServiceUnavailable, "The endpoint is offline", "The tunnel exists but is not currently connected.", fmt.Sprintf("The endpoint '%s' is associated with a tunnel, but the tunnel client is not connected. Please start the tunnel client to resume this tunnel.", html.EscapeString(subdomain)))
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
		ts.logger.Warn().
			Str("subdomain", subdomain).
			Str("tunnel_id", tunnel.ID).
			Msg("WebSocket connection is nil - tunnel may be reconnecting, waiting briefly")

		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			tunnel.mu.RLock()
			wsConn = tunnel.WSConn
			tunnel.mu.RUnlock()
			if wsConn != nil {
				ts.logger.Info().
					Str("subdomain", subdomain).
					Msg("Tunnel reconnected, proceeding with request")
				break
			}
		}

		if wsConn == nil {
			ts.logger.Error().
				Str("subdomain", subdomain).
				Str("tunnel_id", tunnel.ID).
				Msg("WebSocket connection still nil after wait - returning 503")
			ts.writeErrorPage(w, r, tunnel, http.StatusServiceUnavailable, "Tunnel Connection Lost", "The tunnel connection was lost.", "The WebSocket connection is not available. Please reconnect the tunnel client.")
			return
		}
	}

	if r.URL.Path == "/tunnel" {
		return
	}

	connectionHeader := strings.ToLower(r.Header.Get("Connection"))
	upgradeHeader := strings.ToLower(r.Header.Get("Upgrade"))
	isWebSocketUpgrade := strings.Contains(connectionHeader, "upgrade") &&
		upgradeHeader == "websocket"

	if isWebSocketUpgrade {
		ts.proxyWebSocket(tunnel, w, r)
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
	// Pass tunnel subdomain to writeResponse for correct redirect rewriting
	ts.forwardHTTPRequest(tunnel, w, r)
}

func (ts *TunnelServer) proxyWebSocket(tunnel *TunnelConnection, w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path
	defer func() {
		if p := recover(); p != nil {
			ts.logger.Error().Interface("panic", p).Str("path", requestPath).Msg("WebSocket proxy panic recovered")
		}
	}()
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ts.logger.Error().
			Err(err).
			Str("path", r.URL.Path).
			Str("query", r.URL.RawQuery).
			Msg("Failed to upgrade client WebSocket connection")
		return
	}
	defer clientConn.Close()

	ts.logger.Info().
		Str("path", r.URL.Path).
		Str("query", r.URL.RawQuery).
		Str("local_url", tunnel.LocalURL).
		Msg("WebSocket connection upgraded, connecting to local server")

	localURL, err := url.Parse(tunnel.LocalURL)
	if err != nil {
		ts.logger.Error().Err(err).Str("local_url", tunnel.LocalURL).Msg("Failed to parse local URL")
		clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Invalid tunnel configuration"))
		return
	}

	wsScheme := "ws"
	if localURL.Scheme == "https" {
		wsScheme = "wss"
	}

	wsPath := r.URL.Path
	if wsPath == "" || wsPath == "/" {
		wsPath = "/"
	}
	wsURL := fmt.Sprintf("%s://%s%s", wsScheme, localURL.Host, wsPath)
	if r.URL.RawQuery != "" {
		wsURL += "?" + r.URL.RawQuery
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	headers := make(http.Header)
	for k, v := range r.Header {
		lowerKey := strings.ToLower(k)
		if lowerKey == "connection" ||
			lowerKey == "upgrade" ||
			lowerKey == "sec-websocket-key" ||
			lowerKey == "sec-websocket-version" ||
			lowerKey == "sec-websocket-extensions" ||
			lowerKey == "sec-websocket-protocol" {
			continue
		}
		headers[k] = v
	}

	headers.Set("Host", localURL.Host)

	if origin := r.Header.Get("Origin"); origin != "" {
		headers.Set("Origin", origin)
	}

	remoteConn, resp, err := dialer.Dial(wsURL, headers)
	if err != nil {
		ts.logger.Warn().
			Err(err).
			Str("ws_url", wsURL).
			Str("local_url", tunnel.LocalURL).
			Str("origin", r.Header.Get("Origin")).
			Int("status_code", func() int {
				if resp != nil {
					return resp.StatusCode
				}
				return 0
			}()).
			Msg("Failed to connect to local server WebSocket - local server may not support WebSocket or may be down")

		closeMsg := websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, "Failed to connect to local server")
		clientConn.WriteMessage(websocket.CloseMessage, closeMsg)
		clientConn.Close()
		return
	}

	ts.logger.Info().
		Str("ws_url", wsURL).
		Str("path", r.URL.Path).
		Msg("Successfully connected to local server WebSocket, starting proxy")
	defer remoteConn.Close()

	done := make(chan struct{}, 2)

	go func() {
		defer func() { done <- struct{}{} }()
		for {
			messageType, data, err := clientConn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					ts.logger.Debug().Err(err).Msg("Client WebSocket read error")
				}
				return
			}
			if err := remoteConn.WriteMessage(messageType, data); err != nil {
				ts.logger.Debug().Err(err).Msg("Remote WebSocket write error")
				return
			}
		}
	}()

	go func() {
		defer func() { done <- struct{}{} }()
		for {
			messageType, data, err := remoteConn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					ts.logger.Debug().Err(err).Msg("Remote WebSocket read error")
				}
				return
			}
			if err := clientConn.WriteMessage(messageType, data); err != nil {
				ts.logger.Debug().Err(err).Msg("Client WebSocket write error")
				return
			}
		}
	}()

	<-done
	ts.logger.Debug().Str("path", requestPath).Msg("WebSocket proxy connection closed")
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
	// Note: Rate limits are per-tunnel, not per-user or per-API-key
	// This ensures each tunnel has its own rate limit bucket
	allowed, err := ts.rateLimiter.CheckRateLimit(r.Context(), tunnel.ID)
	if err != nil {
		ts.logger.Error().Err(err).Str("tunnel_id", tunnel.ID).Msg("Rate limit check failed")
		// Allow request if rate limit check fails (fail open)
	} else if !allowed {
		// Get current config for logging
		currentConfig := ts.rateLimiter.GetRateLimit(tunnel.ID)
		ts.logger.Warn().
			Str("tunnel_id", tunnel.ID).
			Str("subdomain", tunnel.Subdomain).
			Str("local_url", tunnel.LocalURL).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("current_limit_per_minute", currentConfig.RequestsPerMinute).
			Int("current_limit_per_hour", currentConfig.RequestsPerHour).
			Int("current_limit_per_day", currentConfig.RequestsPerDay).
			Msg("Rate limit exceeded for tunnel - request blocked (if using API key, ensure tunnel was created after API key login)")
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
	// Re-check connection state right before writing to avoid race conditions
	tunnel.mu.RLock()
	wsConn := tunnel.WSConn
	handlerReady := tunnel.handlerReady
	tunnel.mu.RUnlock()

	if wsConn == nil {
		ts.requestTracker.FailRequest(requestID, fmt.Errorf("tunnel connection lost"))
		ts.writeErrorPage(w, r, tunnel, http.StatusBadGateway, "Tunnel Connection Lost", "The tunnel connection was lost while processing your request.", "The tunnel client disconnected. Please reconnect the tunnel client.")
		return
	}

	// Check if handler is ready (for resumed tunnels)
	if !handlerReady {
		ts.logger.Debug().
			Str("request_id", requestID).
			Str("tunnel_id", tunnel.ID).
			Msg("Tunnel handler not ready yet - waiting briefly")
		// Wait up to 200ms for handler to be ready
		for i := 0; i < 20; i++ {
			time.Sleep(10 * time.Millisecond)
			tunnel.mu.RLock()
			handlerReady = tunnel.handlerReady
			wsConn = tunnel.WSConn // Re-check connection
			tunnel.mu.RUnlock()
			if handlerReady && wsConn != nil {
				break
			}
			if wsConn == nil {
				// Connection was closed during wait
				ts.requestTracker.FailRequest(requestID, fmt.Errorf("tunnel connection lost during handler wait"))
				ts.writeErrorPage(w, r, tunnel, http.StatusBadGateway, "Tunnel Connection Lost", "The tunnel connection was lost while processing your request.", "The tunnel client disconnected. Please reconnect the tunnel client.")
				return
			}
		}
		if !handlerReady {
			ts.logger.Warn().
				Str("request_id", requestID).
				Str("tunnel_id", tunnel.ID).
				Msg("Tunnel handler still not ready after wait - proceeding anyway")
		}
	}

	tunnel.writeMu.Lock()

	tunnel.mu.RLock()
	finalWSConn := tunnel.WSConn
	tunnel.mu.RUnlock()

	if finalWSConn == nil {
		tunnel.writeMu.Unlock()
		ts.requestTracker.FailRequest(requestID, fmt.Errorf("tunnel connection lost"))
		ts.writeErrorPage(w, r, tunnel, http.StatusServiceUnavailable, "Tunnel Connection Lost", "The tunnel connection was lost while processing your request.", "The tunnel client disconnected. Please reconnect the tunnel client.")
		return
	}

	finalWSConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	err = finalWSConn.WriteJSON(msg)
	finalWSConn.SetWriteDeadline(time.Time{})
	tunnel.writeMu.Unlock()

	if err != nil {
		// Check if this is a "write closed" error - connection was closed during write
		errStr := err.Error()
		isWriteClosed := strings.Contains(errStr, "write closed") ||
			strings.Contains(errStr, "use of closed network connection") ||
			websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)

		ts.requestTracker.FailRequest(requestID, err)

		if isWriteClosed {
			ts.logger.Warn().
				Err(err).
				Str("request_id", requestID).
				Str("tunnel_id", tunnel.ID).
				Msg("WebSocket connection closed during write - tunnel may be reconnecting")
			// Return 503 (Service Unavailable) instead of 502 for connection closed errors
			// This indicates the tunnel exists but is temporarily unavailable
			ts.writeErrorPage(w, r, tunnel, http.StatusServiceUnavailable, "Tunnel Temporarily Unavailable", "The tunnel connection was closed while processing your request.", "The tunnel client may be reconnecting. Please try again in a moment.")
		} else {
			ts.logger.Error().
				Err(err).
				Str("request_id", requestID).
				Str("tunnel_id", tunnel.ID).
				Msg("Failed to send request to tunnel client - connection may be broken")
			ts.writeErrorPage(w, r, tunnel, http.StatusBadGateway, "Tunnel Connection Error", "Failed to forward request to tunnel client.", html.EscapeString(err.Error()))
		}
		return
	}
	// Clear write deadline after successful write
	finalWSConn.SetWriteDeadline(time.Time{})

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
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

	ts.writeResponse(w, r, response)

	latency := time.Since(start)
	tunnel.mu.Lock()
	tunnel.RequestCount++
	tunnel.mu.Unlock()

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

	isError := response.Status >= 400
	ts.statsCollector.RecordRequest(tunnel.ID, int(latency.Milliseconds()), len(reqData.Body), len(response.Body), isError)

	if ts.repository != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := ts.repository.UpdateTunnelActivity(ctx, tunnel.ID, 1); err != nil {
				ts.logger.Error().Err(err).Str("tunnel_id", tunnel.ID).Msg("Failed to update tunnel activity")
			}
		}()
	}
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

	// Normalize path - ensure root path is "/" not empty
	path := r.URL.Path
	if path == "" {
		path = "/"
	}

	return &HTTPRequest{
		RequestID: requestID,
		Method:    r.Method,
		Path:      path,
		Query:     r.URL.RawQuery,
		Headers:   headers,
		Body:      body,
	}, nil
}

// writeResponse writes HTTP response from tunnel
func (ts *TunnelServer) writeResponse(w http.ResponseWriter, r *http.Request, resp *HTTPResponse) {
	// Once WriteHeader is called, headers cannot be modified

	// Get tunnel URL for rewriting redirects
	// This prevents redirects from accidentally creating URLs with different subdomains
	tunnelHost := r.Host
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	tunnelURL := fmt.Sprintf("%s://%s", scheme, tunnelHost)

	// Validate that we can extract a subdomain from the Host header
	// This ensures redirects will work correctly
	requestSubdomain := extractSubdomain(tunnelHost)
	if requestSubdomain == "" && !strings.Contains(tunnelHost, "localhost") {
		ts.logger.Warn().
			Str("host", tunnelHost).
			Str("request_path", r.URL.Path).
			Msg("Cannot extract subdomain from Host header - redirect rewriting may be incorrect")
	}

	// Rewrite Location header for redirects (3xx status codes)
	// If local server redirects to localhost:PORT, rewrite to tunnel URL
	// External redirects (like naijacrawl.com) pass through unchanged
	if resp.Status >= 300 && resp.Status < 400 {
		// Check for Location header (case-insensitive)
		location := ""
		for k, v := range resp.Headers {
			if strings.EqualFold(k, "Location") && v != "" {
				location = v
				break
			}
		}

		if location != "" {
			ts.logger.Debug().
				Str("redirect_location", location).
				Int("status_code", resp.Status).
				Msg("Processing redirect Location header")

			// Parse the location URL
			locationURL, err := url.Parse(location)
			if err == nil {
				// Check if it's a localhost or 127.0.0.1 URL
				host := locationURL.Hostname()
				if host == "localhost" || host == "127.0.0.1" {
					// Rewrite to use tunnel URL
					tunnelURLParsed, err := url.Parse(tunnelURL)
					if err == nil {
						// Preserve the path and query from the original location
						tunnelURLParsed.Path = locationURL.Path
						tunnelURLParsed.RawQuery = locationURL.RawQuery
						tunnelURLParsed.Fragment = locationURL.Fragment
						// Update the Location header with the rewritten URL
						// Use canonical header name
						rewrittenURL := tunnelURLParsed.String()
						resp.Headers["Location"] = rewrittenURL

						// Validate that the rewritten URL uses the same subdomain as the request
						rewrittenSubdomain := extractSubdomain(tunnelURLParsed.Host)
						requestSubdomain := extractSubdomain(tunnelHost)
						if rewrittenSubdomain != requestSubdomain && rewrittenSubdomain != "" && requestSubdomain != "" {
							ts.logger.Warn().
								Str("original_location", location).
								Str("rewritten_location", rewrittenURL).
								Str("request_subdomain", requestSubdomain).
								Str("rewritten_subdomain", rewrittenSubdomain).
								Str("request_host", tunnelHost).
								Msg("Redirect rewrite changed subdomain - this may cause issues")
						}

						ts.logger.Debug().
							Str("original_location", location).
							Str("rewritten_location", rewrittenURL).
							Str("subdomain", requestSubdomain).
							Msg("Rewrote redirect Location header to use tunnel URL")
					}
				} else if host != "" {
					// External redirect (like naijacrawl.com) - pass through unchanged
					ts.logger.Debug().
						Str("external_redirect", location).
						Msg("Passing through external redirect unchanged")
				} else {
					// Relative redirect (like /path) - pass through unchanged
					ts.logger.Debug().
						Str("relative_redirect", location).
						Msg("Passing through relative redirect unchanged")
				}
			} else {
				ts.logger.Warn().
					Err(err).
					Str("location", location).
					Msg("Failed to parse Location URL, passing through unchanged")
			}
		} else {
			ts.logger.Debug().
				Int("status_code", resp.Status).
				Msg("Redirect response but no Location header found")
		}
	}

	for k, v := range resp.Headers {
		if strings.EqualFold(k, "Content-Length") {
			continue
		}

		if strings.EqualFold(k, "Content-Security-Policy") {
			continue
		}

		if strings.EqualFold(k, "Content-Type") && strings.Contains(strings.ToLower(v), "text/html") {
			bodyStr := string(resp.Body)
			if len(bodyStr) > 0 && (strings.Contains(bodyStr, "import ") ||
				strings.Contains(bodyStr, "export ") ||
				strings.Contains(bodyStr, "class ") ||
				strings.Contains(bodyStr, "function ") ||
				strings.HasPrefix(bodyStr, "import") ||
				strings.HasPrefix(bodyStr, "export")) {
				w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
				continue
			}
		}

		w.Header().Set(k, v)
	}

	var finalBody []byte
	if len(resp.Body) > 0 {
		body := resp.Body
		contentType := resp.Headers["Content-Type"]

		if strings.Contains(contentType, "text/html") {
			w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; img-src 'self' data: https:; font-src 'self' data: https:; connect-src 'self' *;")
		}

		if strings.Contains(contentType, "text/html") ||
			strings.Contains(contentType, "text/css") ||
			strings.Contains(contentType, "application/javascript") ||
			strings.Contains(contentType, "text/javascript") {
			tunnelHost := r.Host
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			tunnelURL := fmt.Sprintf("%s://%s", scheme, tunnelHost)

			bodyStr := string(body)

			re := regexp.MustCompile(`(https?://)localhost:\d+(/[^"'\s>]*)?`)
			bodyStr = re.ReplaceAllStringFunc(bodyStr, func(match string) string {
				pathMatch := regexp.MustCompile(`localhost:\d+(.*)`)
				if pathMatch.MatchString(match) {
					path := pathMatch.FindStringSubmatch(match)[1]
					return tunnelURL + path
				}
				return tunnelURL
			})

			re2 := regexp.MustCompile(`(https?://)127\.0\.0\.1:\d+(/[^"'\s>]*)?`)
			bodyStr = re2.ReplaceAllStringFunc(bodyStr, func(match string) string {
				pathMatch := regexp.MustCompile(`127\.0\.0\.1:\d+(.*)`)
				if pathMatch.MatchString(match) {
					path := pathMatch.FindStringSubmatch(match)[1]
					return tunnelURL + path
				}
				return tunnelURL
			})

			if strings.Contains(contentType, "text/html") {
				tunnelMonitorScript := `<script>
(function() {
	var checkInterval = 5000; // Check every 5 seconds (less aggressive)
	var failures = 0;
	var reloading = false;
	var lastCheck = Date.now();
	var consecutiveFailures = 0; // Track consecutive failures
	var consecutiveSuccesses = 0; // Track consecutive successes (for recovery detection)
	var maxFailures = 3; // Require 3 consecutive failures before reloading
	var maxSuccesses = 2; // Require 2 consecutive successes before reloading on recovery
	var wasOffline = false; // Track if we were showing error page
	
	// Check if we're currently on an error page (503, 502, etc.)
	function isOnErrorPage() {
		// Check if page title or content indicates error page
		var title = document.title.toLowerCase();
		var bodyText = document.body ? document.body.innerText.toLowerCase() : '';
		return title.includes('503') || title.includes('502') || title.includes('bad gateway') || 
		       title.includes('service unavailable') || bodyText.includes('tunnel connection') ||
		       bodyText.includes('connection lost') || bodyText.includes('offline');
	}
	
	function checkConnection() {
		if (reloading) return;
		
		var url = window.location.href.split('?')[0] + '?_check=' + Date.now();
		var xhr = new XMLHttpRequest();
		var timedOut = false;
		var timeoutId = setTimeout(function() {
			timedOut = true;
			xhr.abort();
			consecutiveFailures++;
			consecutiveSuccesses = 0;
			wasOffline = true;
			if (consecutiveFailures >= maxFailures) {
				reloading = true;
				window.location.reload();
			}
		}, 3000); // Longer timeout (3 seconds)
		
		xhr.open('GET', url, true);
		xhr.setRequestHeader('Cache-Control', 'no-cache, no-store, must-revalidate');
		xhr.setRequestHeader('Pragma', 'no-cache');
		xhr.setRequestHeader('Expires', '0');
		
		xhr.onload = function() {
			clearTimeout(timeoutId);
			lastCheck = Date.now();
			if (xhr.status >= 200 && xhr.status < 400) {
				// Success - reset failure counter
				consecutiveFailures = 0;
				failures = 0;
				consecutiveSuccesses++;
				
				// If we were offline (showing error page) and now we have successful responses,
				// reload to show the restored page
				if (wasOffline && consecutiveSuccesses >= maxSuccesses) {
					// Check if we're still on an error page
					if (isOnErrorPage()) {
						reloading = true;
						wasOffline = false; // Reset flag before reload
						consecutiveSuccesses = 0; // Reset counter
						window.location.reload();
					} else {
						// We're already on the correct page, just reset flags
						wasOffline = false;
						consecutiveSuccesses = 0;
					}
				}
			} else if (xhr.status >= 500 || xhr.status === 0 || xhr.status === 503 || xhr.status === 502) {
				// Server errors - increment failure counter
				consecutiveFailures++;
				consecutiveSuccesses = 0;
				wasOffline = true;
				if (consecutiveFailures >= maxFailures) {
					reloading = true;
					window.location.reload();
				}
			} else {
				// Other status codes (like 404) - don't count as failures, might be normal
				consecutiveFailures = 0;
				consecutiveSuccesses = 0;
			}
		};
		
		xhr.onerror = function() {
			clearTimeout(timeoutId);
			consecutiveFailures++;
			consecutiveSuccesses = 0;
			wasOffline = true;
			// Only reload if we have multiple consecutive failures AND haven't checked in a while
			if (consecutiveFailures >= maxFailures && (Date.now() - lastCheck > 5000)) {
				reloading = true;
				window.location.reload();
			}
		};
		
		xhr.ontimeout = function() {
			clearTimeout(timeoutId);
			consecutiveFailures++;
			consecutiveSuccesses = 0;
			wasOffline = true;
			if (consecutiveFailures >= maxFailures) {
				reloading = true;
				window.location.reload();
			}
		};
		
		xhr.timeout = 3000; // 3 second timeout
		xhr.send();
	}
	
	function startMonitoring() {
		// Don't start checking immediately - wait a bit after page load
		setTimeout(function() {
			if (document.readyState === 'complete' || document.readyState === 'interactive') {
				setTimeout(checkConnection, 2000); // Wait 2 seconds after page load
			} else {
				window.addEventListener('load', function() {
					setTimeout(checkConnection, 2000); // Wait 2 seconds after page load
				});
			}
		}, 1000);
		
		// Check periodically, but less frequently
		setInterval(checkConnection, checkInterval);
		
		// Only check when page becomes visible (user switched back to tab)
		document.addEventListener('visibilitychange', function() {
			if (!document.hidden && Date.now() - lastCheck > 10000) {
				// Only check if it's been more than 10 seconds since last check
				setTimeout(checkConnection, 500);
			}
		});
	}
	
	// Only reload on actual network offline events, and only after multiple failures
	window.addEventListener('offline', function() {
		consecutiveFailures++;
		if (consecutiveFailures >= maxFailures && !reloading) {
			reloading = true;
			setTimeout(function() {
				window.location.reload();
			}, 2000); // Wait 2 seconds before reloading
		}
	});
	
	// Don't reload on resource errors - these are often normal (missing images, etc.)
	// Only reload if we get many consecutive resource errors AND connection failures
	var resourceErrors = 0;
	window.addEventListener('error', function(e) {
		// Ignore resource errors - they're often normal and shouldn't trigger reloads
		// Only track them for debugging, but don't reload based on them
		if (e.target && (e.target.tagName === 'SCRIPT' || e.target.tagName === 'LINK' || e.target.tagName === 'IMG' || e.target.tagName === 'IFRAME')) {
			resourceErrors++;
			// Don't reload on resource errors - they're often normal
		}
	}, true);
	
	startMonitoring();
	
	// Only reload if offline AND we've had multiple failures
	if (!navigator.onLine) {
		consecutiveFailures++;
		if (consecutiveFailures >= maxFailures && !reloading) {
			reloading = true;
			setTimeout(function() {
				window.location.reload();
			}, 2000);
		}
	}
})();
</script>`
				if strings.Contains(bodyStr, "</body>") {
					bodyStr = strings.Replace(bodyStr, "</body>", tunnelMonitorScript+"</body>", 1)
				} else if strings.Contains(bodyStr, "</html>") {
					bodyStr = strings.Replace(bodyStr, "</html>", tunnelMonitorScript+"</html>", 1)
				} else {
					bodyStr = bodyStr + tunnelMonitorScript
				}
				ts.logger.Info().Str("path", r.URL.Path).Msg("Injected tunnel monitoring script into HTML response")
			}

			finalBody = []byte(bodyStr)
		} else {
			finalBody = body
		}

		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(finalBody)))
	} else {
		w.Header().Set("Content-Length", "0")
	}

	w.WriteHeader(resp.Status)

	if len(finalBody) > 0 {
		if n, err := w.Write(finalBody); err != nil {
			ts.logger.Error().
				Err(err).
				Int("bytes_written", n).
				Int("body_size", len(finalBody)).
				Str("path", r.URL.Path).
				Msg("Failed to write response body")
		} else if n != len(finalBody) {
			ts.logger.Warn().
				Int("bytes_written", n).
				Int("body_size", len(finalBody)).
				Str("path", r.URL.Path).
				Msg("Partial write of response body")
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
	path := strings.ToLower(r.URL.Path)
	if r.URL.RawQuery != "" {
		path = path + "?" + strings.ToLower(r.URL.RawQuery)
	}

	isAsset := strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".jpeg") || strings.HasSuffix(path, ".gif") ||
		strings.HasSuffix(path, ".svg") || strings.HasSuffix(path, ".woff") ||
		strings.HasSuffix(path, ".woff2") || strings.HasSuffix(path, ".ttf") ||
		strings.HasSuffix(path, ".eot") || strings.HasSuffix(path, ".ico") ||
		strings.HasSuffix(path, ".webmanifest") || strings.Contains(path, "/browser-sync/") ||
		strings.Contains(path, "/node_modules/") || strings.Contains(path, "/assets/")

	if isAsset {
		var contentType string
		if strings.HasSuffix(path, ".js") || strings.Contains(path, "/browser-sync/") {
			contentType = "application/javascript; charset=utf-8"
		} else if strings.HasSuffix(path, ".css") {
			contentType = "text/css; charset=utf-8"
		} else if strings.HasSuffix(path, ".svg") {
			contentType = "image/svg+xml"
		} else if strings.HasSuffix(path, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(path, ".gif") {
			contentType = "image/gif"
		} else if strings.HasSuffix(path, ".woff") {
			contentType = "font/woff"
		} else if strings.HasSuffix(path, ".woff2") {
			contentType = "font/woff2"
		} else if strings.HasSuffix(path, ".ttf") {
			contentType = "font/ttf"
		} else if strings.HasSuffix(path, ".ico") {
			contentType = "image/x-icon"
		} else if strings.HasSuffix(path, ".webmanifest") {
			contentType = "application/manifest+json"
		} else {
			contentType = "application/javascript; charset=utf-8"
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(statusCode)
		return
	}

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

	var iconSVG, iconColor, errorCode string
	switch statusCode {
	case http.StatusNotFound:
		iconSVG = `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>`
		iconColor = "#f59e0b"
		errorCode = `<div class="error-code">ERR_UNIROUTE_404</div>`
	case http.StatusBadGateway:
		iconSVG = `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>`
		iconColor = "#ef4444"
		errorCode = `<div class="error-code">ERR_UNIROUTE_502</div>`
	case http.StatusServiceUnavailable:
		iconSVG = `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z"></path>`
		iconColor = "#f59e0b"
		errorCode = `<div class="error-code">ERR_UNIROUTE_503</div>`
	default:
		iconSVG = `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>`
		iconColor = "#ef4444"
		errorCode = `<div class="error-code">ERR_UNIROUTE_500</div>`
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self';")

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
		.error-code {
			font-size: 14px;
			font-weight: 600;
			color: #94a3b8;
			text-align: center;
			margin-bottom: 8px;
			letter-spacing: 0.5px;
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
		.status-indicator {
			text-align: center;
			margin-top: 24px;
			color: #94a3b8;
			font-size: 13px;
		}
		.status-indicator.checking {
			color: #60a5fa;
		}
		.status-indicator.online {
			color: #22c55e;
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
		
		%s
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
		
		<div class="status-indicator" id="statusIndicator">Checking connection...</div>
		
		<div class="footer">
			Powered by <a href="https://uniroute.co" target="_blank">UniRoute</a>
		</div>
	</div>
	<script>
		(function() {
			var checkInterval = 2000;
			var statusEl = document.getElementById('statusIndicator');
			var isChecking = false;
			var isReloading = false;
			
			function checkConnection() {
				if (isChecking || isReloading) return;
				isChecking = true;
				
				statusEl.textContent = 'Checking connection...';
				statusEl.className = 'status-indicator checking';
				
				var url = window.location.href.split('?')[0] + '?t=' + Date.now();
				fetch(url, {
					method: 'HEAD',
					cache: 'no-cache',
					headers: {
						'Cache-Control': 'no-cache',
						'Pragma': 'no-cache'
					}
				}).then(function(response) {
					isChecking = false;
					if (response.status === 200 || response.status === 304) {
						statusEl.textContent = 'Connection restored! Refreshing...';
						statusEl.className = 'status-indicator online';
						isReloading = true;
						setTimeout(function() {
							window.location.reload();
						}, 500);
					} else {
						statusEl.textContent = 'Still offline. Checking again in ' + (checkInterval / 1000) + ' seconds...';
						statusEl.className = 'status-indicator';
					}
				}).catch(function(error) {
					isChecking = false;
					statusEl.textContent = 'Still offline. Checking again in ' + (checkInterval / 1000) + ' seconds...';
					statusEl.className = 'status-indicator';
				});
			}
			
			checkConnection();
			setInterval(checkConnection, checkInterval);
		})();
	</script>
</body>
</html>`, title, iconColor, iconSVG, errorCode, title, subtitle, publicURL, localURL, statusCode, details)

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

		if protocol == ProtocolUDP {
			ts.portMapMu.Lock()
			for port, t := range ts.portMap {
				if t.ID == tunnelID {
					delete(ts.portMap, port)
					// Close UDP listener
					ts.udpListenersMu.Lock()
					if listener, exists := ts.udpListeners[port]; exists {
						listener.Close()
						delete(ts.udpListeners, port)
					}
					ts.udpListenersMu.Unlock()
					ts.logger.Info().
						Int("port", port).
						Str("tunnel_id", tunnelID).
						Msg("Released UDP port")
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

// forwardUDPData forwards UDP data to the appropriate remote address
func (ts *TunnelServer) forwardUDPData(tunnel *TunnelConnection, connectionID string, data []byte) {
	ts.udpConnMu.RLock()
	udpConn, exists := ts.udpConnections[connectionID]
	ts.udpConnMu.RUnlock()

	if !exists {
		ts.logger.Warn().
			Str("connection_id", connectionID).
			Str("tunnel_id", tunnel.ID).
			Msg("UDP connection not found")
		return
	}

	// Get the UDP listener for this tunnel's port
	tunnel.mu.RLock()
	tunnelID := tunnel.ID
	tunnel.mu.RUnlock()

	ts.portMapMu.RLock()
	var port int
	for p, t := range ts.portMap {
		if t.ID == tunnelID {
			port = p
			break
		}
	}
	ts.portMapMu.RUnlock()

	if port == 0 {
		ts.logger.Error().
			Str("connection_id", connectionID).
			Str("tunnel_id", tunnelID).
			Msg("UDP port not found for tunnel")
		return
	}

	// Get the UDP listener
	ts.udpListenersMu.RLock()
	listener, exists := ts.udpListeners[port]
	ts.udpListenersMu.RUnlock()

	if !exists {
		ts.logger.Error().
			Int("port", port).
			Str("connection_id", connectionID).
			Msg("UDP listener not found")
		return
	}

	// Write data to the remote address
	udpConn.mu.RLock()
	addr := udpConn.Addr
	udpConn.mu.RUnlock()

	_, err := listener.WriteTo(data, addr)
	if err != nil {
		ts.logger.Error().
			Err(err).
			Str("connection_id", connectionID).
			Str("remote_addr", addr.String()).
			Msg("Failed to write UDP data")
		ts.closeUDPConnection(connectionID)
	}
}

// handleUDPError handles UDP errors from tunnel client
func (ts *TunnelServer) handleUDPError(tunnel *TunnelConnection, connectionID string, err *HTTPError) {
	ts.logger.Error().
		Str("connection_id", connectionID).
		Str("error", err.Error).
		Str("message", err.Message).
		Msg("UDP connection error from tunnel")

	// Close the UDP connection
	ts.closeUDPConnection(connectionID)
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
	tunnel.mu.RLock()
	tunnelID := tunnel.ID
	tunnel.mu.RUnlock()

	// Check rate limit
	if ts.rateLimiter != nil {
		allowed, err := ts.rateLimiter.CheckRateLimit(context.Background(), tunnelID)
		if err != nil {
			ts.logger.Error().
				Err(err).
				Str("tunnel_id", tunnelID).
				Msg("Rate limit check failed for TCP connection")
		}
		if !allowed {
			ts.logger.Warn().
				Str("tunnel_id", tunnelID).
				Msg("TCP connection rejected - rate limit exceeded")
			conn.Close()
			return
		}
		ts.rateLimiter.RecordRequest(context.Background(), tunnelID)
	}

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

	tunnel.writeMu.Lock()
	err := wsConn.WriteJSON(msg)
	tunnel.writeMu.Unlock()

	if err != nil {
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
				tunnel.writeMu.Lock()
				wsConn.WriteJSON(closeMsg)
				tunnel.writeMu.Unlock()
				return
			}

			if n > 0 {
				// Forward data to tunnel
				dataMsg := TunnelMessage{
					Type:      msgType,
					RequestID: connectionID,
					Data:      buffer[:n],
				}
				tunnel.writeMu.Lock()
				err := wsConn.WriteJSON(dataMsg)
				tunnel.writeMu.Unlock()

				if err != nil {
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

// allocateUDPPort allocates a UDP port for a tunnel and starts listening on it
func (ts *TunnelServer) allocateUDPPort(tunnel *TunnelConnection) int {
	ts.portMapMu.Lock()
	defer ts.portMapMu.Unlock()

	// Find next available port
	maxPort := ts.tcpPortBase + 10000 // Allow up to 10000 UDP tunnels
	startPort := ts.nextTCPPort

	for i := 0; i < 10000; i++ {
		port := (startPort + i) % maxPort
		if port < ts.tcpPortBase {
			port = ts.tcpPortBase + (port % 10000)
		}

		// Check if port is already allocated
		if _, exists := ts.portMap[port]; !exists {
			// Try to listen on this port to ensure it's available
			testConn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
			if err == nil {
				testConn.Close()

				// Allocate port
				ts.portMap[port] = tunnel
				ts.nextTCPPort = port + 1

				// Start listener for this port in background
				go ts.startUDPListener(port, tunnel)

				return port
			}
		}
	}

	ts.logger.Error().Msg("Failed to allocate UDP port - no available ports")
	return 0
}

// startUDPListener starts a UDP listener on a specific port for a tunnel
func (ts *TunnelServer) startUDPListener(port int, tunnel *TunnelConnection) {
	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		ts.logger.Error().
			Err(err).
			Int("port", port).
			Str("tunnel_id", tunnel.ID).
			Msg("Failed to start UDP listener for port")
		return
	}

	// Store listener
	ts.udpListenersMu.Lock()
	ts.udpListeners[port] = conn
	ts.udpListenersMu.Unlock()

	ts.logger.Info().
		Int("port", port).
		Str("tunnel_id", tunnel.ID).
		Str("protocol", tunnel.Protocol).
		Msg("UDP listener started for tunnel port")

	// Handle UDP packets
	buffer := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			// Check if tunnel still exists
			ts.portMapMu.RLock()
			_, exists := ts.portMap[port]
			ts.portMapMu.RUnlock()

			if !exists {
				ts.logger.Debug().
					Int("port", port).
					Msg("UDP listener closed - tunnel removed")
				conn.Close()
				ts.udpListenersMu.Lock()
				delete(ts.udpListeners, port)
				ts.udpListenersMu.Unlock()
				return
			}

			ts.logger.Error().
				Err(err).
				Int("port", port).
				Msg("Failed to read UDP packet")
			continue
		}

		// Handle UDP packet
		go ts.handleUDPPacket(tunnel, port, addr, buffer[:n])
	}
}

// handleUDPPacket handles an incoming UDP packet
func (ts *TunnelServer) handleUDPPacket(tunnel *TunnelConnection, port int, addr net.Addr, data []byte) {
	tunnel.mu.RLock()
	tunnelID := tunnel.ID
	tunnel.mu.RUnlock()

	// Check rate limit
	if ts.rateLimiter != nil {
		allowed, err := ts.rateLimiter.CheckRateLimit(context.Background(), tunnelID)
		if err != nil {
			ts.logger.Error().
				Err(err).
				Str("tunnel_id", tunnelID).
				Msg("Rate limit check failed for UDP packet")
		}
		if !allowed {
			ts.logger.Warn().
				Str("tunnel_id", tunnelID).
				Msg("UDP packet rejected - rate limit exceeded")
			return
		}
		ts.rateLimiter.RecordRequest(context.Background(), tunnelID)
	}

	// Generate connection ID based on remote address (UDP is connectionless)
	connectionID := fmt.Sprintf("%s-%d", addr.String(), time.Now().UnixNano())

	// Store connection info (UDP is stateless, but we track for metrics)
	ts.udpConnMu.Lock()
	ts.udpConnections[connectionID] = &UDPConnection{
		ID:        connectionID,
		TunnelID:  tunnel.ID,
		Addr:      addr,
		CreatedAt: time.Now(),
	}
	ts.udpConnMu.Unlock()

	// Get WebSocket connection
	tunnel.mu.RLock()
	wsConn := tunnel.WSConn
	tunnel.mu.RUnlock()

	if wsConn == nil {
		ts.logger.Error().
			Str("tunnel_id", tunnel.ID).
			Msg("Tunnel WebSocket connection not available")
		ts.closeUDPConnection(connectionID)
		return
	}

	// Send UDP data to tunnel client
	msg := TunnelMessage{
		Type:      MsgTypeUDPData,
		RequestID: connectionID,
		Data:      data,
	}

	tunnel.writeMu.Lock()
	err := wsConn.WriteJSON(msg)
	tunnel.writeMu.Unlock()

	if err != nil {
		ts.logger.Error().
			Err(err).
			Str("connection_id", connectionID).
			Msg("Failed to send UDP packet to tunnel")
		ts.closeUDPConnection(connectionID)
		return
	}

	ts.logger.Debug().
		Str("connection_id", connectionID).
		Str("tunnel_id", tunnel.ID).
		Int("data_size", len(data)).
		Msg("UDP packet forwarded to tunnel client")
}

// closeUDPConnection closes and removes a UDP connection
func (ts *TunnelServer) closeUDPConnection(connectionID string) {
	ts.udpConnMu.Lock()
	defer ts.udpConnMu.Unlock()

	udpConn, exists := ts.udpConnections[connectionID]
	if exists {
		delete(ts.udpConnections, connectionID)
		ts.logger.Debug().
			Str("connection_id", connectionID).
			Str("tunnel_id", udpConn.TunnelID).
			Msg("UDP connection closed")
	}
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
