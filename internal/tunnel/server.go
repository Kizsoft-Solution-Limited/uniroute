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
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

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
	tcpPortBase     int // Base port for TCP/UDP tunnel allocation (default: 20000)
	tcpPortRange    int // Number of ports in the range (default: 10000). Use 101 for 20000-20100 to speed up Docker.
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
	apiKeyValidator func(ctx context.Context, apiKey string) (userID string, err error)
	apiKeyValidatorWithLimits func(ctx context.Context, apiKey string) (userID string, rateLimitPerMinute, rateLimitPerDay int, err error)
	baseDomain      string // Base domain for tunnels (e.g., "uniroute.co")
	websiteURL      string // Website URL for links (e.g., "https://uniroute.co")
}

type TCPConnection struct {
	ID        string
	TunnelID  string
	Conn      net.Conn
	CreatedAt time.Time
	mu        sync.RWMutex
}

type UDPConnection struct {
	ID        string
	TunnelID  string
	Addr      net.Addr // Remote address for UDP
	CreatedAt time.Time
	mu        sync.RWMutex
}

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

func NewTunnelServer(port int, logger zerolog.Logger, allowedOrigins []string) *TunnelServer {
	defaultOrigins := []string{
		"http://localhost",
		"https://localhost",
		"http://127.0.0.1",
		"https://127.0.0.1",
	}

	baseDomain := getEnv("TUNNEL_BASE_DOMAIN", "")
	if baseDomain != "" {
		defaultOrigins = append(defaultOrigins, baseDomain, "."+baseDomain)
	}

	originPatterns := defaultOrigins
	if len(allowedOrigins) > 0 {
		originPatterns = append(originPatterns, allowedOrigins...)
	}

	websiteURL := getEnv("WEBSITE_URL", getEnv("BASE_URL", "https://uniroute.co"))
	if baseDomain == "" {
		baseDomain = getEnv("TUNNEL_BASE_DOMAIN", "uniroute.co")
	}

	return &TunnelServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				for _, pattern := range originPatterns {
					if strings.Contains(origin, pattern) {
						return true
					}
				}
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
		tcpPortBase:     getEnvAsInt("TUNNEL_TCP_PORT_BASE", 20000),
		tcpPortRange:    getEnvAsInt("TUNNEL_TCP_PORT_RANGE", 10000),
		nextTCPPort:     getEnvAsInt("TUNNEL_TCP_PORT_BASE", 20000),
		port:            port,
		logger:          logger,
		subdomainPrefix: "tunnel",
		requestTracker:  NewRequestTracker(logger),
		tokenService:    NewTokenService(logger),
		rateLimiter:     NewTunnelRateLimiter(logger),
		statsCollector:  NewStatsCollector(logger),
		security:        NewSecurityMiddleware(logger),
		requireAuth:     false,
		baseDomain:      baseDomain,
		websiteURL:      websiteURL,
	}
}

func getEnvAsInt(key string, defaultValue int) int {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	var intValue int
	fmt.Sscanf(value, "%d", &intValue)
	if intValue == 0 {
		return defaultValue
	}
	return intValue
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (ts *TunnelServer) SetTCPPortBase(basePort int) {
	ts.tcpPortBase = basePort
	ts.nextTCPPort = basePort
}

func (ts *TunnelServer) SetRepository(repo *TunnelRepository) {
	ts.repository = repo
	if repo != nil {
		ts.requestLogger = NewRequestLogger(repo, ts.logger)
	}
}

func (ts *TunnelServer) SetRateLimiter(limiter RateLimiterInterface) {
	ts.rateLimiter = limiter
}

func (ts *TunnelServer) SetStatsRedis(client *storage.RedisClient) {
	ts.statsCollector.SetRedisClient(client)
}

func (ts *TunnelServer) SetDomainManager(manager *DomainManager) {
	ts.domainManager = manager
}

func (ts *TunnelServer) SetRequireAuth(require bool) {
	ts.requireAuth = require
}

func (ts *TunnelServer) SetJWTValidator(validator func(tokenString string) (userID string, err error)) {
	ts.jwtValidator = validator
}

func (ts *TunnelServer) SetAPIKeyValidator(validator func(ctx context.Context, apiKey string) (userID string, err error)) {
	ts.apiKeyValidator = validator
}

func (ts *TunnelServer) SetAPIKeyValidatorWithLimits(validator func(ctx context.Context, apiKey string) (userID string, rateLimitPerMinute, rateLimitPerDay int, err error)) {
	ts.apiKeyValidatorWithLimits = validator
}

func (ts *TunnelServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/tunnel", ts.handleTunnelConnection)
	mux.HandleFunc("/health", ts.handleHealth)
	mux.HandleFunc("/", ts.handleHTTPRequest)
	mux.HandleFunc("/web", ts.handleWebInterface)
	mux.HandleFunc("/api/tunnels", ts.handleListTunnels)
	mux.HandleFunc("/api/tunnels/", func(w http.ResponseWriter, r *http.Request) {
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
	if ts.repository != nil {
		go ts.monitorInactiveTunnels()
	}
	return ts.httpServer.ListenAndServe()
}

func (ts *TunnelServer) handleWebInterface(w http.ResponseWriter, r *http.Request) {
	ts.handleRootRequest(w, r)
}

func (ts *TunnelServer) Stop(ctx context.Context) error {
	return ts.httpServer.Shutdown(ctx)
}

func (ts *TunnelServer) handleTunnelConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := ts.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ts.logger.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}
	var initMsg InitMessage
	if err := ws.ReadJSON(&initMsg); err != nil {
		ts.logger.Error().Err(err).Msg("Failed to read init message")
		return
	}

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

	if initMsg.LocalURL == "" {
		ts.logger.Error().Msg("Local URL is required")
		ws.WriteJSON(map[string]string{"error": "local_url is required"})
		ws.Close()
		return
	}

	if err := validateLocalURL(initMsg.LocalURL); err != nil {
		ts.logger.Error().Err(err).Str("local_url", initMsg.LocalURL).Msg("Invalid local URL")
		ws.WriteJSON(map[string]string{"error": "invalid local_url: " + err.Error()})
		ws.Close()
		return
	}

	if initMsg.Subdomain != "" {
		if err := validateSubdomain(initMsg.Subdomain); err != nil {
			ts.logger.Error().Err(err).Str("subdomain", initMsg.Subdomain).Msg("Invalid subdomain")
			ws.WriteJSON(map[string]string{"error": "invalid subdomain: " + err.Error()})
			ws.Close()
			return
		}
	}

	protocol := initMsg.Protocol
	if protocol == "" {
		protocol = ProtocolHTTP
	}

	if protocol != ProtocolHTTP && protocol != ProtocolTCP && protocol != ProtocolTLS && protocol != ProtocolUDP {
		ts.logger.Error().Str("protocol", protocol).Msg("Invalid protocol")
		ws.WriteJSON(map[string]string{"error": "invalid protocol, must be http, tcp, tls, or udp"})
		ws.Close()
		return
	}

	var subdomain string
	var tunnelID string
	var isResume bool
	var autoFoundTunnel bool

	if initMsg.ForceNew {
		isResume = false
		subdomain = ""
		tunnelID = ""
	} else if initMsg.Subdomain == "" && initMsg.TunnelID == "" && initMsg.Host == "" && authenticatedUserID != "" && ts.repository != nil {

		userUUID, parseErr := uuid.Parse(authenticatedUserID)
		if parseErr == nil {
			userTunnels, err := ts.repository.ListTunnelsByUser(r.Context(), userUUID, initMsg.Protocol)
			if err == nil && len(userTunnels) > 0 {
				ts.tunnelsMu.RLock()
				var matchByLocalURL, firstDisconnected *Tunnel
				for _, dbTunnel := range userTunnels {
					if dbTunnel.UserID != authenticatedUserID {
						continue
					}
					tunnelIDStr := dbTunnel.ID
					existingTunnel, isConnected := ts.tunnels[dbTunnel.Subdomain]
					if isConnected && existingTunnel != nil {
						existingTunnel.mu.RLock()
						existingTunnelID := existingTunnel.ID
						existingWSConn := existingTunnel.WSConn
						existingTunnel.mu.RUnlock()
						if existingTunnelID == tunnelIDStr && existingWSConn != nil {
							ts.logger.Info().
								Str("tunnel_id", tunnelIDStr).
								Str("subdomain", dbTunnel.Subdomain).
								Str("protocol", dbTunnel.Protocol).
								Msg("Tunnel is already connected - skipping and looking for another")
							continue
						}
					}
					if firstDisconnected == nil {
						firstDisconnected = dbTunnel
					}
					if initMsg.LocalURL != "" && dbTunnel.LocalURL == initMsg.LocalURL {
						matchByLocalURL = dbTunnel
						break
					}
				}
				chosen := matchByLocalURL
				if chosen == nil {
					chosen = firstDisconnected
				}
				if chosen != nil {
					subdomain = chosen.Subdomain
					tunnelID = chosen.ID
					isResume = true
					autoFoundTunnel = true
					ts.logger.Info().
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Str("protocol", chosen.Protocol).
						Str("status", chosen.Status).
						Str("local_url", chosen.LocalURL).
						Msg("Found tunnel that is not connected - will resume it")
				}
				ts.tunnelsMu.RUnlock()
				
				if !autoFoundTunnel {
					ts.logger.Info().
						Str("user_id", authenticatedUserID).
						Str("protocol", initMsg.Protocol).
						Int("tunnel_count", len(userTunnels)).
						Msg("All tunnels for user are already connected - will create new tunnel")
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
	} else if initMsg.Subdomain != "" || initMsg.TunnelID != "" || initMsg.Host != "" {
		ts.tunnelsMu.RLock()
		var existingTunnel *TunnelConnection

		if initMsg.Host != "" {
			tunnel := ts.tunnels[initMsg.Host]
			if tunnel != nil {
				tunnel.mu.RLock()
				tunnelProtocol := tunnel.Protocol
				tunnelWSConn := tunnel.WSConn
				tunnel.mu.RUnlock()
				if (initMsg.Protocol == "" || tunnelProtocol == initMsg.Protocol) && tunnelWSConn == nil {
					existingTunnel = tunnel
				} else if tunnelWSConn != nil {
					ts.logger.Info().
						Str("host", initMsg.Host).
						Str("tunnel_id", tunnel.ID).
						Str("protocol", tunnelProtocol).
						Msg("Tunnel found in memory by host but already connected - will check database or create new")
				} else {
					ts.logger.Info().
						Str("host", initMsg.Host).
						Str("tunnel_protocol", tunnelProtocol).
						Str("requested_protocol", initMsg.Protocol).
						Msg("Tunnel found in memory by host but protocol mismatch - will check database or create new")
				}
			}
			
			if existingTunnel == nil && ts.repository != nil {
				dbTunnel, err := ts.repository.GetTunnelByCustomDomain(context.Background(), initMsg.Host)
				if err == nil && dbTunnel != nil {
					if dbTunnel.Protocol != "" && initMsg.Protocol != "" && dbTunnel.Protocol != initMsg.Protocol {
						ts.logger.Info().
							Str("host", initMsg.Host).
							Str("tunnel_protocol", dbTunnel.Protocol).
							Str("requested_protocol", initMsg.Protocol).
							Msg("Tunnel found by custom domain but protocol mismatch - will create new")
					} else {
						ts.tunnelsMu.RLock()
						connectedTunnel, isConnected := ts.tunnels[dbTunnel.Subdomain]
						ts.tunnelsMu.RUnlock()
						
						if isConnected && connectedTunnel != nil {
							connectedTunnel.mu.RLock()
							connectedTunnelID := connectedTunnel.ID
							connectedWSConn := connectedTunnel.WSConn
							connectedTunnel.mu.RUnlock()
							
							if connectedTunnelID == dbTunnel.ID && connectedWSConn != nil {
								ts.logger.Info().
									Str("host", initMsg.Host).
									Str("tunnel_id", dbTunnel.ID).
									Str("subdomain", dbTunnel.Subdomain).
									Str("protocol", dbTunnel.Protocol).
									Msg("Tunnel found by custom domain but already connected - will create new tunnel")
							} else {
								subdomain = dbTunnel.Subdomain
								tunnelID = dbTunnel.ID
								isResume = true
								ts.logger.Info().
									Str("host", initMsg.Host).
									Str("tunnel_id", tunnelID).
									Str("subdomain", subdomain).
									Str("protocol", dbTunnel.Protocol).
									Msg("Found tunnel by custom domain/host - not connected - will resume it")
							}
						} else {
							subdomain = dbTunnel.Subdomain
							tunnelID = dbTunnel.ID
							isResume = true
							ts.logger.Info().
								Str("host", initMsg.Host).
								Str("tunnel_id", tunnelID).
								Str("subdomain", subdomain).
								Str("protocol", dbTunnel.Protocol).
								Msg("Found tunnel by custom domain/host - not connected - will resume it")
						}
					}
				}
			}
		} else if initMsg.Subdomain != "" {
			tunnel := ts.tunnels[initMsg.Subdomain]
			if tunnel != nil {
				tunnel.mu.RLock()
				tunnelProtocol := tunnel.Protocol
				tunnel.mu.RUnlock()
				if initMsg.Protocol == "" || tunnelProtocol == initMsg.Protocol {
					existingTunnel = tunnel
				} else {
					ts.logger.Info().
						Str("subdomain", initMsg.Subdomain).
						Str("tunnel_protocol", tunnelProtocol).
						Str("requested_protocol", initMsg.Protocol).
						Msg("Tunnel found in memory but protocol mismatch - will check database or create new")
				}
			}
		} else if initMsg.TunnelID != "" {
			for _, t := range ts.tunnels {
				if t.ID == initMsg.TunnelID {
					t.mu.RLock()
					tunnelProtocol := t.Protocol
					t.mu.RUnlock()
					if initMsg.Protocol == "" || tunnelProtocol == initMsg.Protocol {
						existingTunnel = t
						break
					} else {
						ts.logger.Info().
							Str("tunnel_id", initMsg.TunnelID).
							Str("tunnel_protocol", tunnelProtocol).
							Str("requested_protocol", initMsg.Protocol).
							Msg("Tunnel found in memory but protocol mismatch - will check database or create new")
					}
				}
			}
		}
		ts.tunnelsMu.RUnlock()

		if existingTunnel != nil {
			existingTunnel.mu.RLock()
			existingProtocol := existingTunnel.Protocol
			existingWSConn := existingTunnel.WSConn
			existingSubdomain := existingTunnel.Subdomain
			existingTunnel.mu.RUnlock()

			if existingWSConn != nil {
				ts.logger.Info().
					Str("tunnel_id", existingTunnel.ID).
					Str("subdomain", existingSubdomain).
					Str("protocol", existingProtocol).
					Msg("Tunnel is already connected - rejecting resume")
				ws.WriteJSON(map[string]interface{}{
					"error":   "tunnel_already_active",
					"message": fmt.Sprintf("Tunnel %s is already connected by another client. Stop the other session or use --new to create a new tunnel.", existingSubdomain),
				})
				return
			} else {
				subdomain = existingTunnel.Subdomain
				tunnelID = existingTunnel.ID
				isResume = true
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("subdomain", subdomain).
					Str("protocol", existingProtocol).
					Msg("Resuming existing tunnel (from memory) - not connected - will create fresh tunnel connection")
			}
		} else if ts.repository != nil {
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
					ts.logger.Debug().
						Err(parseErr).
						Str("tunnel_id", initMsg.TunnelID).
						Msg("Failed to parse tunnel ID as UUID - trying to look up by subdomain")

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
				if authenticatedUserID != "" {
					if dbTunnel.UserID == "" || dbTunnel.UserID == "null" {
						ts.logger.Warn().
							Str("tunnel_id", dbTunnel.ID).
							Str("subdomain", dbTunnel.Subdomain).
							Str("authenticated_user_id", authenticatedUserID).
							Str("db_user_id", dbTunnel.UserID).
							Msg("Tunnel has null/empty user_id - REJECTING resume (will create new tunnel)")
						dbTunnel = nil
						isResume = false
					} else if dbTunnel.UserID != authenticatedUserID {
						ts.logger.Info().
							Str("tunnel_id", dbTunnel.ID).
							Str("subdomain", dbTunnel.Subdomain).
							Str("tunnel_user_id", dbTunnel.UserID).
							Str("authenticated_user_id", authenticatedUserID).
							Msg("Tunnel belongs to different user - skipping resume (will create new tunnel)")
						dbTunnel = nil
						isResume = false
					} else {
						if initMsg.Protocol != "" && dbTunnel.Protocol != "" && dbTunnel.Protocol != initMsg.Protocol {
							ts.logger.Info().
								Str("tunnel_id", dbTunnel.ID).
								Str("subdomain", dbTunnel.Subdomain).
								Str("tunnel_protocol", dbTunnel.Protocol).
								Str("requested_protocol", initMsg.Protocol).
								Msg("Protocol mismatch - tunnel is different type - will create new tunnel")
							dbTunnel = nil
							isResume = false
						} else {
							ts.tunnelsMu.RLock()
							connectedTunnel, isConnected := ts.tunnels[dbTunnel.Subdomain]
							ts.tunnelsMu.RUnlock()
							
							if isConnected && connectedTunnel != nil {
								connectedTunnel.mu.RLock()
								connectedTunnelID := connectedTunnel.ID
								connectedWSConn := connectedTunnel.WSConn
								connectedTunnel.mu.RUnlock()
								
								if connectedTunnelID == dbTunnel.ID && connectedWSConn != nil {
									ts.logger.Info().
										Str("tunnel_id", dbTunnel.ID).
										Str("subdomain", dbTunnel.Subdomain).
										Str("custom_domain", dbTunnel.CustomDomain).
										Str("protocol", dbTunnel.Protocol).
										Msg("Tunnel from database is already connected - rejecting resume")
									ws.WriteJSON(map[string]interface{}{
										"error":   "tunnel_already_active",
										"message": fmt.Sprintf("Tunnel %s is already connected by another client. Stop the other session or use --new to create a new tunnel.", dbTunnel.Subdomain),
									})
									return
								} else {
									subdomain = dbTunnel.Subdomain
									tunnelID = dbTunnel.ID
									isResume = true
									ts.logger.Info().
										Str("tunnel_id", tunnelID).
										Str("subdomain", subdomain).
										Str("custom_domain", dbTunnel.CustomDomain).
										Str("previous_status", dbTunnel.Status).
										Str("protocol", dbTunnel.Protocol).
										Msg("Resuming existing tunnel (from database) - not connected - will become active again")
								}
							} else {
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
									Str("protocol", dbTunnel.Protocol).
									Str("user_id", dbTunnel.UserID).
									Str("requested_subdomain", initMsg.Subdomain).
									Str("requested_tunnel_id", initMsg.TunnelID).
									Str("database_subdomain", dbTunnel.Subdomain).
									Msg("Resuming existing tunnel (from database) - not connected - will become active again")
							}
						}
					}
				} else {
					ts.logger.Warn().
						Str("tunnel_id", dbTunnel.ID).
						Str("subdomain", dbTunnel.Subdomain).
						Msg("Tunnel found but user is not authenticated - skipping resume (will create new tunnel)")
					dbTunnel = nil
				}
			}

			if dbTunnel == nil {
				if autoFoundTunnel && isResume && subdomain != "" && tunnelID != "" {
					ts.logger.Info().
						Str("subdomain", subdomain).
						Str("tunnel_id", tunnelID).
						Str("authenticated_user_id", authenticatedUserID).
						Msg("Auto-found tunnel validated - preserving resume flag even though lookup didn't find it (will resume auto-found tunnel)")
				} else {
					isResume = false
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
			ts.logger.Info().
				Str("requested_subdomain", initMsg.Subdomain).
				Str("requested_tunnel_id", initMsg.TunnelID).
				Msg("Requested tunnel not found in memory (no database), creating new tunnel")
		}
	}

	if autoFoundTunnel {
		ts.logger.Info().
			Bool("is_resume", isResume).
			Str("subdomain", subdomain).
			Str("tunnel_id", tunnelID).
			Str("authenticated_user_id", authenticatedUserID).
			Msg("Auto-found tunnel state - proceeding to resume logic")
	}

	if !isResume {
		var err error
		if initMsg.Host != "" {
			if err := validateSubdomain(initMsg.Host); err != nil {
				ts.logger.Warn().
					Err(err).
					Str("requested_host", initMsg.Host).
					Msg("Invalid subdomain requested")
				ws.WriteJSON(map[string]string{"error": "invalid subdomain: " + err.Error()})
				ws.Close()
				return
			}

			if !initMsg.ForceNew {
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

				ts.tunnelsMu.RLock()
				existingTunnelByHost, existsInMemory := ts.tunnels[initMsg.Host]
				ts.tunnelsMu.RUnlock()

				if existsInMemory && existingTunnelByHost != nil {
					existingTunnelByHost.mu.RLock()
					existingWSConn := existingTunnelByHost.WSConn
					existingTunnelByHost.mu.RUnlock()
					
					if existingWSConn != nil {
						ts.logger.Warn().
							Str("requested_host", initMsg.Host).
							Str("tunnel_id", existingTunnelByHost.ID).
							Msg("Requested subdomain/host is already connected - cannot resume")
						ws.WriteJSON(map[string]string{"error": "subdomain '" + initMsg.Host + "' is already connected by another client"})
						ws.Close()
						return
					}
				}

				if !available {
					ts.logger.Warn().
						Str("requested_host", initMsg.Host).
						Bool("available_in_db", available).
						Msg("Requested subdomain is not available")
					ws.WriteJSON(map[string]string{"error": "subdomain '" + initMsg.Host + "' is not available"})
					ws.Close()
					return
				}
			} else {
				ts.logger.Info().
					Str("requested_host", initMsg.Host).
					Bool("force_new", initMsg.ForceNew).
					Msg("ForceNew is set - will create new tunnel with requested subdomain (may replace existing)")
			}

			subdomain = initMsg.Host
			ts.logger.Info().
				Str("requested_subdomain", subdomain).
				Bool("force_new", initMsg.ForceNew).
				Msg("Using requested subdomain")
		} else {
			if ts.domainManager != nil {
				subdomain, err = ts.domainManager.AllocateSubdomain(context.Background(), ts.repository)
				if err != nil {
					ts.logger.Error().Err(err).Msg("Failed to allocate subdomain")
					ws.WriteJSON(map[string]string{"error": "failed to allocate subdomain"})
					return
				}
			} else {
				subdomain = ts.generateSubdomain()
			}
		}
		tunnelID = generateID()
	} else {
		if tunnelID != "" {
			if _, err := uuid.Parse(tunnelID); err != nil {
				if !autoFoundTunnel {
					ts.logger.Warn().
						Err(err).
						Str("invalid_tunnel_id", tunnelID).
						Msg("Resume tunnel ID is not a valid UUID - using database ID instead")
					if ts.repository != nil {
						var dbTunnel *Tunnel
						if subdomain != "" {
							dbTunnel, err = ts.repository.GetTunnelBySubdomain(context.Background(), subdomain)
						}
						if err == nil && dbTunnel != nil {
							tunnelID = dbTunnel.ID
						} else {
							ts.logger.Warn().Msg("Cannot resume tunnel with invalid ID - creating new tunnel")
							isResume = false
							tunnelID = generateID()
						}
					} else {
						ts.logger.Warn().Msg("Cannot validate tunnel ID without database - creating new tunnel")
						isResume = false
						tunnelID = generateID()
					}
				} else {
					ts.logger.Error().
						Err(err).
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Msg("CRITICAL: Auto-found tunnel has invalid UUID - this should not happen")
				}
			}
		} else if autoFoundTunnel {
			ts.logger.Error().
				Str("subdomain", subdomain).
				Msg("CRITICAL: Auto-found tunnel but tunnelID is empty - this should not happen")
		}
	}

	var tunnel *TunnelConnection
	if subdomain == "" {
		ts.logger.Error().Msg("CRITICAL: subdomain is empty - cannot create tunnel")
		ws.WriteJSON(map[string]string{"error": "subdomain is required"})
		return
	}

	if isResume {
		if subdomain == "" {
			ts.logger.Error().
				Str("tunnel_id", tunnelID).
				Str("init_subdomain", initMsg.Subdomain).
				Msg("CRITICAL: Cannot resume tunnel - subdomain is empty")
			ws.WriteJSON(map[string]string{"error": "subdomain is required for resume"})
			return
		}

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
			handlerReady: false,
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

		ts.tunnelsMu.RLock()
		verifyTunnel, verifyExists := ts.tunnels[subdomain]
		ts.tunnelsMu.RUnlock()

		if !verifyExists || verifyTunnel != tunnel {
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

		if protocol == ProtocolTCP || protocol == ProtocolTLS {
			var existingPort int
			ts.portMapMu.RLock()
			for port, t := range ts.portMap {
				if t.ID == tunnelID {
					existingPort = port
					break
				}
			}
			ts.portMapMu.RUnlock()

			if existingPort > 0 {
				ts.portMapMu.Lock()
				ts.portMap[existingPort] = tunnel
				ts.portMapMu.Unlock()
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("protocol", protocol).
					Int("tcp_port", existingPort).
					Msg("Reusing existing TCP port for resumed tunnel")
			} else {
				tcpPort := ts.allocateTCPPort(tunnel)
				if tcpPort > 0 {
					ts.logger.Info().
						Str("tunnel_id", tunnelID).
						Str("protocol", protocol).
						Int("tcp_port", tcpPort).
						Msg("Allocated new TCP port for resumed tunnel")
				}
			}
		}

		if protocol == ProtocolUDP {
			var existingPort int
			ts.portMapMu.RLock()
			for port, t := range ts.portMap {
				if t.ID == tunnelID {
					existingPort = port
					break
				}
			}
			ts.portMapMu.RUnlock()

			if existingPort > 0 {
				ts.portMapMu.Lock()
				ts.portMap[existingPort] = tunnel
				ts.portMapMu.Unlock()
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("protocol", protocol).
					Int("udp_port", existingPort).
					Msg("Reusing existing UDP port for resumed tunnel")
			} else {
				udpPort := ts.allocateUDPPort(tunnel)
				if udpPort > 0 {
					ts.logger.Info().
						Str("tunnel_id", tunnelID).
						Str("protocol", protocol).
						Int("udp_port", udpPort).
						Msg("Allocated new UDP port for resumed tunnel")
				}
			}
		}

		if ts.repository != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				var existingTunnel *Tunnel
				tunnelUUID, parseErr := uuid.Parse(tunnelID)
				if parseErr == nil {
					existingTunnel, _ = ts.repository.GetTunnelByID(ctx, tunnelUUID)
				} else {
					ts.logger.Debug().
						Err(parseErr).
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Msg("Tunnel ID is not a valid UUID, trying to find tunnel by subdomain")
					if subdomain != "" {
						existingTunnel, _ = ts.repository.GetTunnelBySubdomain(ctx, subdomain)
					}
				}

				if err := ts.repository.UpdateTunnelLocalURL(ctx, tunnelID, initMsg.LocalURL); err != nil {
					ts.logger.Error().
						Err(err).
						Str("tunnel_id", tunnelID).
						Str("local_url", initMsg.LocalURL).
						Msg("Failed to update tunnel LocalURL on resume")
				}

				if existingTunnel != nil && existingTunnel.UserID == "" && initMsg.Token != "" {
					if ts.jwtValidator != nil {
						if extractedUserID, err := ts.jwtValidator(initMsg.Token); err == nil && extractedUserID != "" {
							if userUUID, parseErr := uuid.Parse(extractedUserID); parseErr == nil {
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

				if err := ts.repository.UpdateTunnelStatus(ctx, tunnelID, "active"); err != nil {
					ts.logger.Error().
						Err(err).
						Str("tunnel_id", tunnelID).
						Msg("Failed to update tunnel status to active on resume from database")
				} else {
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
					if err := ts.repository.UpdateTunnelActivity(ctx, tunnelID, 0); err != nil {
						ts.logger.Error().
							Err(err).
							Str("tunnel_id", tunnelID).
							Msg("Failed to update tunnel last_active_at on resume from database")
					}
				}
			}()
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
					BurstSize:         50, // Higher burst for high-limit API keys
				}
				ts.rateLimiter.SetRateLimit(tunnelID, rateLimitConfig)
			}
		}
	} else {
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
			handlerReady: false,
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

		ts.tunnelsMu.Lock()
		existingTunnel := ts.tunnels[subdomain]
		ts.tunnels[subdomain] = tunnel
		ts.tunnelsMu.Unlock()
		if existingTunnel != nil && existingTunnel != tunnel {
			existingTunnel.mu.Lock()
			oldWSConn := existingTunnel.WSConn
			existingTunnel.WSConn = nil
			existingTunnel.mu.Unlock()
			if oldWSConn != nil && oldWSConn != ws {
				go func() {
					time.Sleep(50 * time.Millisecond)
					oldWSConn.Close()
				}()
			}
			ts.logger.Info().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Str("old_tunnel_id", existingTunnel.ID).
				Msg("Replaced existing tunnel with new tunnel for subdomain")
		}

		if protocol == ProtocolTCP || protocol == ProtocolTLS {
			tcpPort := ts.allocateTCPPort(tunnel)
			if tcpPort > 0 {
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("protocol", protocol).
					Int("tcp_port", tcpPort).
					Msg("Allocated new TCP port for tunnel")
			}
		}
		if protocol == ProtocolUDP {
			udpPort := ts.allocateUDPPort(tunnel)
			if udpPort > 0 {
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("protocol", protocol).
					Int("udp_port", udpPort).
					Msg("Allocated new UDP port for tunnel")
			}
		}
	}

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

	var publicURL string
	tunnel.mu.RLock()
	tunnelProtocol := tunnel.Protocol
	tunnel.mu.RUnlock()

	if tunnelProtocol == ProtocolTCP || tunnelProtocol == ProtocolTLS {
		ts.portMapMu.RLock()
		var tcpPort int
		for port, t := range ts.portMap {
			if t == tunnel || t.ID == tunnelID || t.Subdomain == subdomain {
				tcpPort = port
				break
			}
		}
		ts.portMapMu.RUnlock()

		if tcpPort > 0 {
			if ts.domainManager != nil {
				publicURL = fmt.Sprintf("%s:%d", ts.domainManager.GetPublicHost(subdomain), tcpPort)
			} else {
				localhostDomain := getEnv("TUNNEL_LOCALHOST_DOMAIN", "localhost")
				publicURL = fmt.Sprintf("%s.%s:%d", subdomain, localhostDomain, tcpPort)
			}
		} else {
			ts.logger.Warn().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Str("protocol", tunnelProtocol).
				Msg("TCP port not found in portMap - TCP port allocation may have failed")
			localhostDomain := getEnv("TUNNEL_LOCALHOST_DOMAIN", "localhost")
			publicURL = fmt.Sprintf("%s.%s:<tcp_port>", subdomain, localhostDomain)
		}
	} else if tunnelProtocol == ProtocolUDP {
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
			if ts.domainManager != nil {
				publicURL = fmt.Sprintf("%s:%d", ts.domainManager.GetPublicHost(subdomain), udpPort)
			} else {
				localhostDomain := getEnv("TUNNEL_LOCALHOST_DOMAIN", "localhost")
				publicURL = fmt.Sprintf("%s.%s:%d", subdomain, localhostDomain, udpPort)
			}
		} else {
			ts.logger.Warn().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Str("protocol", tunnelProtocol).
				Msg("UDP port not found in portMap - using fallback URL")
			localhostDomain := getEnv("TUNNEL_LOCALHOST_DOMAIN", "localhost")
			publicURL = fmt.Sprintf("%s.%s:%d", subdomain, localhostDomain, ts.port)
		}
	} else {
		localhostDomain := getEnv("TUNNEL_LOCALHOST_DOMAIN", "localhost")
		publicURL = fmt.Sprintf("http://%s.%s:%d", subdomain, localhostDomain, ts.port)
		if ts.domainManager != nil {
			useHTTPS := ts.domainManager.baseDomain != ""
			publicURL = ts.domainManager.GetPublicURL(subdomain, ts.port, useHTTPS)
		}
	}

	if publicURL == "" {
		ts.logger.Warn().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Str("protocol", tunnelProtocol).
			Msg("Public URL is empty - using fallback")
		localhostDomain := getEnv("TUNNEL_LOCALHOST_DOMAIN", "localhost")
		publicURL = fmt.Sprintf("http://%s.%s:%d", subdomain, localhostDomain, ts.port)
	}

	ts.logger.Debug().
		Bool("repository_available", ts.repository != nil).
		Bool("is_resume", isResume).
		Str("tunnel_id", tunnelID).
		Str("subdomain", subdomain).
		Msg("Checking if tunnel should be saved to database")

	if ts.repository != nil && !isResume {
		var userID string
		if authenticatedUserID != "" {
			userID = authenticatedUserID
			ts.logger.Info().
				Str("tunnel_id", tunnelID).
				Str("user_id", userID).
				Str("subdomain", subdomain).
				Bool("is_api_key", strings.HasPrefix(initMsg.Token, "ur_")).
				Msg("Using authenticated user ID - tunnel will be associated with user")
		} else {
			ts.logger.Warn().
				Str("tunnel_id", tunnelID).
				Str("subdomain", subdomain).
				Str("has_token", fmt.Sprintf("%v", initMsg.Token != "")).
				Msg("authenticatedUserID is empty but creating tunnel - attempting to extract from token")

			if initMsg.Token != "" {
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
			UserID:       userID,
			Subdomain:    subdomain,
			LocalURL:     initMsg.LocalURL,
			PublicURL:    publicURL,
			Protocol:     protocol,
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
		} else {
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

	ts.tunnelsMu.RLock()
	registeredTunnel, isRegistered := ts.tunnels[subdomain]
	ts.tunnelsMu.RUnlock()

	if !isRegistered || registeredTunnel != tunnel {
		ts.logger.Error().
			Str("subdomain", subdomain).
			Bool("is_registered", isRegistered).
			Msg("CRITICAL: Tunnel not properly registered before sending response - registering now")
		ts.tunnelsMu.Lock()
		ts.tunnels[subdomain] = tunnel
		ts.tunnelsMu.Unlock()
	}

	if tunnel.WSConn == nil {
		ts.logger.Error().
			Str("subdomain", subdomain).
			Str("tunnel_id", tunnelID).
			Msg("CRITICAL: WebSocket connection is nil - cannot start message handler")
		ws.WriteJSON(map[string]string{"error": "internal error: websocket connection not set"})
		return
	}

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

	region := getEnv("TUNNEL_REGION", "")
	response := InitResponse{
		Type:      MsgTypeTunnelCreated,
		TunnelID:  tunnelID,
		Subdomain: subdomain,
		PublicURL: publicURL,
		Status:    "active",
		Region:    region,
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

func (ts *TunnelServer) handleTunnelMessages(tunnel *TunnelConnection) {
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

	tunnel.mu.Lock()
	tunnel.handlerReady = true
	tunnel.mu.Unlock()

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
		ts.tunnelsMu.Lock()
		ts.tunnels[tunnelSubdomain] = tunnel
		ts.tunnelsMu.Unlock()
		ts.logger.Info().
			Str("tunnel_id", tunnelID).
			Str("subdomain", tunnelSubdomain).
			Msg("Re-registered tunnel after immediate removal")
	}

	defer func() {
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

	heartbeatInterval := 30 * time.Second
	readDeadline := 3 * heartbeatInterval

	for {
		tunnel.mu.RLock()
		currentWSConn := tunnel.WSConn
		tunnel.mu.RUnlock()

		if currentWSConn == nil || currentWSConn != ourWSConn {
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
			tunnel.mu.RLock()
			currentWSConn = tunnel.WSConn
			tunnel.mu.RUnlock()

			if currentWSConn != ourWSConn {
				ts.logger.Debug().
					Str("tunnel_id", tunnelID).
					Str("subdomain", tunnelSubdomain).
					Msg("WebSocket connection replaced during read - old handler exiting (tunnel resumed)")
				return
			}

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

		tunnel.WSConn.SetReadDeadline(time.Time{})

		tunnel.mu.Lock()
		tunnel.LastActive = time.Now()
		tunnel.mu.Unlock()

		if msg.Type == MsgTypePing {
			if ts.repository != nil {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					tunnel.mu.RLock()
					tunnelID := tunnel.ID
					tunnel.mu.RUnlock()

					if err := ts.repository.UpdateTunnelActivity(ctx, tunnelID, 0); err != nil {
						ts.logger.Debug().
							Err(err).
							Str("tunnel_id", tunnelID).
							Msg("Failed to update tunnel last_active_at on ping (non-critical)")
					}
				}()
			}
		}

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
			if msg.Error != nil {
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
			if msg.RequestID != "" {
				ts.forwardTCPData(tunnel, msg.RequestID, msg.Data)
			}
		case MsgTypeTCPError, MsgTypeTLSError:
			if msg.Error != nil {
				ts.handleTCPError(tunnel, msg.RequestID, msg.Error)
			}
		case MsgTypeUDPData:
			if msg.RequestID != "" {
				ts.forwardUDPData(tunnel, msg.RequestID, msg.Data)
			}
		case MsgTypeUDPError:
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

func (ts *TunnelServer) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/tunnel" {
		return
	}

	path := r.URL.Path
	if path == "/favicon.ico" || path == "/favicon.png" {
		http.NotFound(w, r)
		return
	}

	ts.security.AddSecurityHeaders(w, r)
	if r.Method == http.MethodOptions {
		return // Preflight handled
	}

	if err := ts.security.ValidateRequest(r); err != nil {
		ts.logger.Warn().Err(err).Str("method", r.Method).Str("path", r.URL.Path).Msg("Invalid request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	r.URL.Path = ts.security.SanitizePath(r.URL.Path)

	host := r.Host
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

	if subdomain != "" {
		ts.tunnelsMu.RLock()
		tunnel, exists = ts.tunnels[subdomain]
		ts.tunnelsMu.RUnlock()
		lookupSubdomain = subdomain
	}

	if !exists && ts.repository != nil {
		dbTunnel, err := ts.repository.GetTunnelByCustomDomain(context.Background(), hostname)
		if err == nil && dbTunnel != nil {
			if dbTunnel.Protocol != "" && dbTunnel.Protocol != ProtocolHTTP {
				ts.logger.Warn().
					Str("custom_domain", hostname).
					Str("tunnel_protocol", dbTunnel.Protocol).
					Msg("Custom domain found but tunnel is not HTTP - custom domains only work with HTTP tunnels")
			} else {
				lookupSubdomain = dbTunnel.Subdomain
				ts.tunnelsMu.RLock()
				tunnel, exists = ts.tunnels[lookupSubdomain]
				ts.tunnelsMu.RUnlock()

				if exists && tunnel != nil {
					tunnel.mu.RLock()
					tunnelProtocol := tunnel.Protocol
					tunnel.mu.RUnlock()
					if tunnelProtocol != ProtocolHTTP {
						ts.logger.Warn().
							Str("custom_domain", hostname).
							Str("tunnel_subdomain", lookupSubdomain).
							Str("tunnel_protocol", tunnelProtocol).
							Msg("Tunnel found by custom domain but protocol mismatch - custom domains only work with HTTP")
						tunnel = nil
						exists = false
					}
				}

				ts.logger.Info().
					Str("custom_domain", hostname).
					Str("tunnel_subdomain", lookupSubdomain).
					Bool("tunnel_active", exists).
					Msg("Found tunnel by custom domain")
			}
		}
	}

	if subdomain == "" && !exists {
		ts.logger.Debug().Str("host", host).Msg("No subdomain or custom domain found, showing root page")
		ts.handleRootRequest(w, r)
		return
	}
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

		if ts.repository != nil && tunnelID != "" {
			tunnelUUID, parseErr := uuid.Parse(tunnelID)
			if parseErr == nil {
				dbTunnel, err := ts.repository.GetTunnelByID(context.Background(), tunnelUUID)
				if err == nil && dbTunnel != nil && dbTunnel.Status == "inactive" {
					ts.logger.Info().
						Str("tunnel_id", tunnelID).
						Str("subdomain", lookupSubdomain).
						Str("status", dbTunnel.Status).
						Msg("Tunnel is inactive in database (disconnected from dashboard) - closing connection")
					
					tunnel.mu.RLock()
					wsConn := tunnel.WSConn
					tunnel.mu.RUnlock()
					if wsConn != nil {
						wsConn.Close()
					}
					
					ts.removeTunnel(lookupSubdomain)
					
					ts.writeErrorPage(w, r, nil, http.StatusServiceUnavailable, "Tunnel Disconnected", "This tunnel has been disconnected.", "The tunnel was disconnected from the dashboard. Please reconnect the tunnel client.")
					return
				}
			}
		}

		if tunnelProtocol != ProtocolHTTP {
			ts.logger.Warn().
				Str("subdomain", lookupSubdomain).
				Str("tunnel_id", tunnelID).
				Str("protocol", tunnelProtocol).
				Msg("HTTP request received for non-HTTP tunnel - protocol mismatch")
			ts.writeErrorPage(w, r, tunnel, http.StatusBadRequest, "Protocol Mismatch", "This tunnel is not an HTTP tunnel.", fmt.Sprintf("The tunnel '%s' is configured for %s protocol, not HTTP.", html.EscapeString(lookupSubdomain), tunnelProtocol))
			return
		}

		tunnel.mu.RLock()
		wsConn := tunnel.WSConn
		tunnel.mu.RUnlock()

		if wsConn == nil {
			ts.logger.Warn().
				Str("subdomain", lookupSubdomain).
				Str("tunnel_id", tunnelID).
				Msg("Tunnel exists but WebSocket connection is nil - tunnel is disconnected")
			if ts.repository != nil {
				dbTunnel, err := ts.repository.GetTunnelBySubdomain(context.Background(), lookupSubdomain)
				if err == nil && dbTunnel != nil {
					ts.writeErrorPage(w, r, nil, http.StatusServiceUnavailable, "The endpoint is offline", "The tunnel exists but is not currently connected.", fmt.Sprintf("The endpoint '%s' is associated with a tunnel, but the tunnel client is not connected. Please start the tunnel client to resume this tunnel.", html.EscapeString(lookupSubdomain)))
					return
				}
			}
			ts.writeErrorPage(w, r, nil, http.StatusNotFound, "Tunnel Not Found", "The requested tunnel does not exist.", fmt.Sprintf("The subdomain '%s' is not associated with any tunnel.", html.EscapeString(lookupSubdomain)))
			return
		} else {
			tunnel.mu.RLock()
			handlerReady := tunnel.handlerReady
			tunnel.mu.RUnlock()

			if !handlerReady {
				ts.logger.Debug().
					Str("tunnel_id", tunnel.ID).
					Str("subdomain", lookupSubdomain).
					Msg("Tunnel exists but handler not ready yet - waiting briefly")
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
		if ts.repository != nil {
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

			if !exists {
				dbTunnel, err := ts.repository.GetTunnelBySubdomain(context.Background(), lookupSubdomain)
				if err == nil && dbTunnel != nil {
					ts.logger.Info().
						Str("subdomain", lookupSubdomain).
						Str("tunnel_id", dbTunnel.ID).
						Str("status", dbTunnel.Status).
						Str("local_url", dbTunnel.LocalURL).
						Msg("Tunnel exists in database but not in memory - showing disconnected page (503)")
					detailMsg := "The endpoint '%s' is associated with a tunnel, but the tunnel client is not connected. Please start the tunnel client (e.g. run your tunnel command again). If the tunnel server was recently restarted, re-running the tunnel client will reconnect it."
					ts.writeErrorPage(w, r, nil, http.StatusServiceUnavailable, "The endpoint is offline", "The tunnel exists but is not currently connected.", fmt.Sprintf(detailMsg, html.EscapeString(lookupSubdomain)))
					return
				}
			}
		}

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

	if tunnel == nil {
		ts.logger.Error().
			Str("subdomain", subdomain).
			Msg("CRITICAL: Tunnel is nil after all checks - returning 500")
		http.Error(w, "Internal server error: tunnel is nil", http.StatusInternalServerError)
		return
	}

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

func (ts *TunnelServer) handleRootRequest(w http.ResponseWriter, r *http.Request) {
	ts.tunnelsMu.RLock()
	activeTunnelsInMemory := len(ts.tunnels)
	ts.tunnelsMu.RUnlock()

	activeTunnels := activeTunnelsInMemory
	totalTunnels := activeTunnelsInMemory

	if ts.repository != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		tunnels, err := ts.repository.ListAllTunnels(ctx)
		if err != nil {
			ts.logger.Debug().Err(err).Msg("Failed to fetch tunnels from database for root page")
		} else {
			totalTunnels = len(tunnels)
			activeCount := 0
			for _, t := range tunnels {
				if t.Status == "active" {
					activeCount++
				}
			}
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
			Powered by <a href="%s" target="_blank">UniRoute</a> | 
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
</html>`, activeTunnels, totalTunnels, ts.websiteURL)

	w.Write([]byte(html))
}

func (ts *TunnelServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","tunnels":%d}`, len(ts.tunnels))
}

func (ts *TunnelServer) forwardHTTPRequest(tunnel *TunnelConnection, w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := generateID()

	// Note: Rate limits are per-tunnel, not per-user or per-API-key.
	allowed, err := ts.rateLimiter.CheckRateLimit(r.Context(), tunnel.ID)
	if err != nil {
		ts.logger.Error().Err(err).Str("tunnel_id", tunnel.ID).Msg("Rate limit check failed")
	} else if !allowed {
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

	pendingReq, err := ts.requestTracker.RegisterRequest(requestID)
	if err != nil {
		ts.logger.Error().Err(err).Str("request_id", requestID).Msg("Failed to register request")
		http.Error(w, "Failed to register request", http.StatusInternalServerError)
		return
	}

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

	ts.rateLimiter.RecordRequest(r.Context(), tunnel.ID)

	msg := TunnelMessage{
		Type:      MsgTypeHTTPRequest,
		RequestID: requestID,
		Request:   reqData,
	}

	tunnel.mu.RLock()
	wsConn := tunnel.WSConn
	handlerReady := tunnel.handlerReady
	tunnel.mu.RUnlock()

	if wsConn == nil {
		ts.requestTracker.FailRequest(requestID, fmt.Errorf("tunnel connection lost"))
		ts.writeErrorPage(w, r, tunnel, http.StatusBadGateway, "Tunnel Connection Lost", "The tunnel connection was lost while processing your request.", "The tunnel client disconnected. Please reconnect the tunnel client.")
		return
	}

	if !handlerReady {
		ts.logger.Debug().
			Str("request_id", requestID).
			Str("tunnel_id", tunnel.ID).
			Msg("Tunnel handler not ready yet - waiting briefly")
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

		errMsg := err.Error()
		errMsgLower := strings.ToLower(errMsg)

		ts.logger.Info().
			Str("request_id", requestID).
			Str("error", errMsg).
			Str("error_lowercase", errMsgLower).
			Msg("Error received from tunnel client - checking type for custom error page")

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

	ts.logger.Info().
		Str("request_id", requestID).
		Str("tunnel_id", tunnel.ID).
		Str("subdomain", tunnel.Subdomain).
		Int("status_code", response.Status).
		Int("response_size", len(response.Body)).
		Int("header_count", len(response.Headers)).
		Msg("Writing HTTP response to client")

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

func (ts *TunnelServer) serializeRequest(r *http.Request, requestID string) (*HTTPRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	r.Body.Close()

	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

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

func (ts *TunnelServer) writeResponse(w http.ResponseWriter, r *http.Request, resp *HTTPResponse) {
	tunnelHost := r.Host
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	tunnelURL := fmt.Sprintf("%s://%s", scheme, tunnelHost)

	requestSubdomain := extractSubdomain(tunnelHost)
	if requestSubdomain == "" && !strings.Contains(tunnelHost, "localhost") {
		ts.logger.Warn().
			Str("host", tunnelHost).
			Str("request_path", r.URL.Path).
			Msg("Cannot extract subdomain from Host header - redirect rewriting may be incorrect")
	}

	if resp.Status >= 300 && resp.Status < 400 {
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

			locationURL, err := url.Parse(location)
			if err == nil {
				host := locationURL.Hostname()
				if host == "localhost" || host == "127.0.0.1" {
					tunnelURLParsed, err := url.Parse(tunnelURL)
					if err == nil {
						tunnelURLParsed.Path = locationURL.Path
						tunnelURLParsed.RawQuery = locationURL.RawQuery
						tunnelURLParsed.Fragment = locationURL.Fragment
						rewrittenURL := tunnelURLParsed.String()
						resp.Headers["Location"] = rewrittenURL

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
					ts.logger.Debug().
						Str("external_redirect", location).
						Msg("Passing through external redirect unchanged")
				} else {
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


func (ts *TunnelServer) generateSubdomain() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:12] // 12 character subdomain
}

func (ts *TunnelServer) removeTunnel(subdomain string) {
	ts.tunnelsMu.Lock()
	tunnel, exists := ts.tunnels[subdomain]
	if !exists {
		ts.tunnelsMu.Unlock()
		return // Tunnel already removed
	}

	tunnel.mu.RLock()
	tunnelID := tunnel.ID
	currentWSConn := tunnel.WSConn
	tunnel.mu.RUnlock()

	if ts.tunnels[subdomain] != tunnel {
		ts.logger.Debug().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Msg("Tunnel was re-registered - skipping removal (tunnel was resumed)")
		ts.tunnelsMu.Unlock()
		return // Tunnel was re-registered, don't remove it
	}

	tunnel.mu.RLock()
	wsConnStillActive := tunnel.WSConn != nil && tunnel.WSConn == currentWSConn
	tunnel.mu.RUnlock()

	if wsConnStillActive {
		ts.logger.Warn().
			Str("tunnel_id", tunnelID).
			Str("subdomain", subdomain).
			Msg("Removing tunnel with active WebSocket connection - connection may be dead")
	}

	delete(ts.tunnels, subdomain)
	ts.tunnelsMu.Unlock()

	ts.logger.Debug().
		Str("tunnel_id", tunnelID).
		Str("subdomain", subdomain).
		Bool("ws_conn_nil", currentWSConn == nil).
		Msg("Removed tunnel from memory")

	if exists && tunnel != nil {
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

func (ts *TunnelServer) GetTunnel(subdomain string) (*TunnelConnection, bool) {
	ts.tunnelsMu.RLock()
	defer ts.tunnelsMu.RUnlock()
	tunnel, exists := ts.tunnels[subdomain]
	return tunnel, exists
}

func (ts *TunnelServer) monitorInactiveTunnels() {
	ticker := time.NewTicker(2 * time.Second) // Check every 2 seconds for faster response
	defer ticker.Stop()

	for range ticker.C {
		if ts.repository == nil {
			continue
		}

		ts.tunnelsMu.RLock()
		activeTunnels := make([]*TunnelConnection, 0, len(ts.tunnels))
		for _, tunnel := range ts.tunnels {
			activeTunnels = append(activeTunnels, tunnel)
		}
		ts.tunnelsMu.RUnlock()

		for _, tunnel := range activeTunnels {
			tunnel.mu.RLock()
			tunnelID := tunnel.ID
			subdomain := tunnel.Subdomain
			wsConn := tunnel.WSConn
			tunnel.mu.RUnlock()

			if tunnelID == "" {
				continue
			}

			tunnelUUID, parseErr := uuid.Parse(tunnelID)
			if parseErr != nil {
				continue
			}

			dbTunnel, err := ts.repository.GetTunnelByID(context.Background(), tunnelUUID)
			if err != nil {
				continue
			}

			if dbTunnel != nil && dbTunnel.Status == "inactive" {
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("subdomain", subdomain).
					Msg("Tunnel marked inactive in database - closing connection immediately")

				if wsConn != nil {
					closeMsg := websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Tunnel disconnected from dashboard")
					wsConn.SetWriteDeadline(time.Now().Add(1 * time.Second))
					if err := wsConn.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
						ts.logger.Debug().
							Err(err).
							Str("tunnel_id", tunnelID).
							Str("subdomain", subdomain).
							Msg("Failed to send close message, closing connection anyway")
					}
					time.Sleep(50 * time.Millisecond)
					wsConn.Close()
					tunnel.mu.Lock()
					tunnel.WSConn = nil
					tunnel.mu.Unlock()
					ts.logger.Debug().
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Msg("WebSocket connection closed with policy violation code - handler should exit")
					time.Sleep(100 * time.Millisecond)
				} else {
					ts.logger.Debug().
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Msg("WebSocket already nil, removing tunnel from memory")
				}

				ts.tunnelsMu.Lock()
				if ts.tunnels[subdomain] == tunnel {
					delete(ts.tunnels, subdomain)
					ts.logger.Debug().
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Msg("Removed tunnel from memory map")
				} else {
					ts.logger.Debug().
						Str("tunnel_id", tunnelID).
						Str("subdomain", subdomain).
						Msg("Tunnel was re-registered - skipping removal")
				}
				ts.tunnelsMu.Unlock()

				tunnel.mu.RLock()
				protocol := tunnel.Protocol
				tunnel.mu.RUnlock()
				
				if protocol == ProtocolTCP || protocol == ProtocolTLS {
					ts.portMapMu.Lock()
					for port, t := range ts.portMap {
						if t.ID == tunnelID {
							delete(ts.portMap, port)
							ts.logger.Info().
								Int("port", port).
								Str("tunnel_id", tunnelID).
								Msg("Released TCP port after disconnect")
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
							ts.udpListenersMu.Lock()
							if listener, exists := ts.udpListeners[port]; exists {
								listener.Close()
								delete(ts.udpListeners, port)
							}
							ts.udpListenersMu.Unlock()
							ts.logger.Info().
								Int("port", port).
								Str("tunnel_id", tunnelID).
								Msg("Released UDP port after disconnect")
							break
						}
					}
					ts.portMapMu.Unlock()
				}
				
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("subdomain", subdomain).
					Msg("Tunnel fully disconnected and removed after being marked inactive")
			}
		}
	}
}

func (ts *TunnelServer) ListTunnels() []*TunnelConnection {
	ts.tunnelsMu.RLock()
	defer ts.tunnelsMu.RUnlock()

	tunnels := make([]*TunnelConnection, 0, len(ts.tunnels))
	for _, tunnel := range ts.tunnels {
		tunnels = append(tunnels, tunnel)
	}
	return tunnels
}

func generateID() string {
	return uuid.New().String()
}

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

	tcpConn.mu.Lock()
	_, err := tcpConn.Conn.Write(data)
	tcpConn.mu.Unlock()

	if err != nil {
		ts.logger.Error().
			Err(err).
			Str("connection_id", connectionID).
			Msg("Failed to write TCP data")
		ts.closeTCPConnection(connectionID)
	}
}

func (ts *TunnelServer) handleTCPError(tunnel *TunnelConnection, connectionID string, err *HTTPError) {
	ts.logger.Error().
		Str("connection_id", connectionID).
		Str("error", err.Error).
		Str("message", err.Message).
		Msg("TCP connection error from tunnel")

	ts.closeTCPConnection(connectionID)
}

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

func (ts *TunnelServer) handleUDPError(tunnel *TunnelConnection, connectionID string, err *HTTPError) {
	ts.logger.Error().
		Str("connection_id", connectionID).
		Str("error", err.Error).
		Str("message", err.Message).
		Msg("UDP connection error from tunnel")

	ts.closeUDPConnection(connectionID)
}

func (ts *TunnelServer) allocateTCPPort(tunnel *TunnelConnection) int {
	ts.portMapMu.Lock()
	defer ts.portMapMu.Unlock()

	portRange := ts.tcpPortRange
	if portRange < 1 {
		portRange = 10000
	}
	maxPort := ts.tcpPortBase + portRange
	startPort := ts.nextTCPPort

	for i := 0; i < portRange; i++ {
		port := (startPort + i) % maxPort
		if port < ts.tcpPortBase {
			port = ts.tcpPortBase + (port % portRange)
		}

		if _, exists := ts.portMap[port]; !exists {
			testListener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				testListener.Close()
				ts.portMap[port] = tunnel
				ts.nextTCPPort = port + 1
				go ts.startPortListener(port, tunnel)

				return port
			}
		}
	}

	ts.logger.Error().Msg("Failed to allocate TCP port - no available ports")
	return 0
}

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

	for {
		conn, err := listener.Accept()
		if err != nil {
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
		go ts.handleTCPConnection(tunnel, conn)
	}
}

func (ts *TunnelServer) handleTCPConnection(tunnel *TunnelConnection, conn net.Conn) {
	tunnel.mu.RLock()
	tunnelID := tunnel.ID
	subdomain := tunnel.Subdomain
	tunnel.mu.RUnlock()

	if ts.repository != nil && tunnelID != "" {
		tunnelUUID, parseErr := uuid.Parse(tunnelID)
		if parseErr == nil {
			dbTunnel, err := ts.repository.GetTunnelByID(context.Background(), tunnelUUID)
			if err == nil && dbTunnel != nil && dbTunnel.Status == "inactive" {
				// Tunnel was disconnected from dashboard - close connection and remove from memory
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("subdomain", subdomain).
					Str("status", dbTunnel.Status).
					Msg("TCP tunnel is inactive in database (disconnected from dashboard) - rejecting connection")

				tunnel.mu.RLock()
				wsConn := tunnel.WSConn
				tunnel.mu.RUnlock()
				if wsConn != nil {
					wsConn.Close()
				}
				ts.removeTunnel(subdomain)
				conn.Close()
				return
			}
		}
	}

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

	ts.tcpConnMu.Lock()
	ts.tcpConnections[connectionID] = &TCPConnection{
		ID:        connectionID,
		TunnelID:  tunnel.ID,
		Conn:      conn,
		CreatedAt: time.Now(),
	}
	ts.tcpConnMu.Unlock()

	tunnel.mu.RLock()
	protocol := tunnel.Protocol
	wsConn := tunnel.WSConn
	tunnel.mu.RUnlock()

	msgType := MsgTypeTCPData
	if protocol == ProtocolTLS {
		msgType = MsgTypeTLSData
	}

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

func (ts *TunnelServer) allocateUDPPort(tunnel *TunnelConnection) int {
	ts.portMapMu.Lock()
	defer ts.portMapMu.Unlock()

	portRange := ts.tcpPortRange
	if portRange < 1 {
		portRange = 10000
	}
	maxPort := ts.tcpPortBase + portRange
	startPort := ts.nextTCPPort

	for i := 0; i < portRange; i++ {
		port := (startPort + i) % maxPort
		if port < ts.tcpPortBase {
			port = ts.tcpPortBase + (port % portRange)
		}

		if _, exists := ts.portMap[port]; !exists {
			testConn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
			if err == nil {
				testConn.Close()
				ts.portMap[port] = tunnel
				ts.nextTCPPort = port + 1
				go ts.startUDPListener(port, tunnel)
				return port
			}
		}
	}

	ts.logger.Error().Msg("Failed to allocate UDP port - no available ports")
	return 0
}

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

	ts.udpListenersMu.Lock()
	ts.udpListeners[port] = conn
	ts.udpListenersMu.Unlock()

	ts.logger.Info().
		Int("port", port).
		Str("tunnel_id", tunnel.ID).
		Str("protocol", tunnel.Protocol).
		Msg("UDP listener started for tunnel port")

	buffer := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
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
		go ts.handleUDPPacket(tunnel, port, addr, buffer[:n])
	}
}

func (ts *TunnelServer) handleUDPPacket(tunnel *TunnelConnection, port int, addr net.Addr, data []byte) {
	tunnel.mu.RLock()
	tunnelID := tunnel.ID
	tunnel.mu.RUnlock()

	if ts.repository != nil && tunnelID != "" {
		tunnelUUID, parseErr := uuid.Parse(tunnelID)
		if parseErr == nil {
			dbTunnel, err := ts.repository.GetTunnelByID(context.Background(), tunnelUUID)
			if err == nil && dbTunnel != nil && dbTunnel.Status == "inactive" {
				ts.logger.Info().
					Str("tunnel_id", tunnelID).
					Str("status", dbTunnel.Status).
					Msg("UDP tunnel is inactive in database (disconnected from dashboard) - rejecting packet")

				tunnel.mu.RLock()
				wsConn := tunnel.WSConn
				subdomain := tunnel.Subdomain
				tunnel.mu.RUnlock()
				if wsConn != nil {
					wsConn.Close()
				}
				ts.removeTunnel(subdomain)
				return
			}
		}
	}

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

	connectionID := fmt.Sprintf("%s-%d", addr.String(), time.Now().UnixNano())

	ts.udpConnMu.Lock()
	ts.udpConnections[connectionID] = &UDPConnection{
		ID:        connectionID,
		TunnelID:  tunnel.ID,
		Addr:      addr,
		CreatedAt: time.Now(),
	}
	ts.udpConnMu.Unlock()

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
	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		hostname = host
	}

	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return ""
	}

	localhostDomain := getEnv("TUNNEL_LOCALHOST_DOMAIN", "localhost")
	if parts[1] == localhostDomain {
		subdomain := parts[0]
		if len(subdomain) > 63 {
			return ""
		}
		matched, _ := regexp.MatchString("^[a-zA-Z0-9-]+$", subdomain)
		if !matched {
			return ""
		}
		return subdomain
	}

	subdomain := parts[0]
	if len(subdomain) > 63 {
		return ""
	}
	matched, _ := regexp.MatchString("^[a-zA-Z0-9-]+$", subdomain)
	if !matched {
		return ""
	}
	return subdomain
}

func (ts *TunnelServer) getPublicURLForTunnel(tunnel *TunnelConnection) string {
	tunnel.mu.RLock()
	subdomain := tunnel.Subdomain
	protocol := tunnel.Protocol
	tunnel.mu.RUnlock()

	localhostDomain := getEnv("TUNNEL_LOCALHOST_DOMAIN", "localhost")
	if protocol == ProtocolTCP || protocol == ProtocolTLS {
		ts.portMapMu.RLock()
		var tcpPort int
		for port, t := range ts.portMap {
			if t.ID == tunnel.ID {
				tcpPort = port
				break
			}
		}
		ts.portMapMu.RUnlock()
		if tcpPort > 0 {
			if ts.domainManager != nil {
				return fmt.Sprintf("%s:%d", ts.domainManager.GetPublicHost(subdomain), tcpPort)
			}
			return fmt.Sprintf("%s.%s:%d", subdomain, localhostDomain, tcpPort)
		}
	}

	if protocol == ProtocolUDP {
		ts.portMapMu.RLock()
		var udpPort int
		for port, t := range ts.portMap {
			if t.ID == tunnel.ID {
				udpPort = port
				break
			}
		}
		ts.portMapMu.RUnlock()
		if udpPort > 0 {
			if ts.domainManager != nil {
				return fmt.Sprintf("%s:%d", ts.domainManager.GetPublicHost(subdomain), udpPort)
			}
			return fmt.Sprintf("%s.%s:%d", subdomain, localhostDomain, udpPort)
		}
	}

	if ts.domainManager != nil && ts.domainManager.baseDomain != "" {
		useHTTPS := true
		return ts.domainManager.GetPublicURL(subdomain, ts.port, useHTTPS)
	}
	return fmt.Sprintf("http://%s.%s:%d", subdomain, localhostDomain, ts.port)
}

func validateLocalURL(url string) error {
	if len(url) > 2048 {
		return fmt.Errorf("URL too long (max 2048 characters)")
	}

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

	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return nil
	}

	if strings.Contains(url, ":") {
		parts := strings.Split(url, ":")
		if len(parts) == 2 {
			return nil
		}
	}

	return fmt.Errorf("invalid URL format, must be http://..., https://..., or host:port")
}

// reservedSubdomains are system subdomains that cannot be used for user tunnels (app, api, www, dashboard, etc.)
var reservedSubdomains = map[string]bool{
	"www": true,"tunnel": true, "api": true, "app": true, "admin": true, "dashboard": true, "docs": true,
}

func validateSubdomain(subdomain string) error {
	if len(subdomain) == 0 {
		return fmt.Errorf("subdomain cannot be empty")
	}

	if len(subdomain) > 63 {
		return fmt.Errorf("subdomain too long (max 63 characters)")
	}

	if reservedSubdomains[strings.ToLower(subdomain)] {
		return fmt.Errorf("subdomain is reserved for system use")
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$", subdomain)
	if !matched {
		return fmt.Errorf("subdomain must contain only alphanumeric characters and hyphens, and cannot start or end with hyphen")
	}

	return nil
}
