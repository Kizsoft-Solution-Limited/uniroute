package api

import (
	"net/http"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/api/handlers"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/api/middleware"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/email"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/gateway"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/oauth"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// SetupRouter sets up the API routes and middleware
func SetupRouter(
	router *gateway.Router,
	apiKeyService *security.APIKeyService, // In-memory API key service (fallback)
	apiKeyServiceV2 *security.APIKeyServiceV2, // Database-backed API key service
	jwtService *security.JWTService,
	rateLimiter *security.RateLimiter,
	authRateLimiter *security.AuthRateLimiter, // Progressive rate limiter for auth endpoints
	ipWhitelist []string,
	requestRepo *storage.RequestRepository, // Request repository for analytics
	providerKeyService *security.ProviderKeyService, // BYOK: Provider key service
	postgresClient *storage.PostgresClient, // Database client for user repository
	emailService interface{}, // Email service (can be nil)
	frontendURL string, // Frontend URL for email links
	oauthService *oauth.OAuthService, // OAuth service (can be nil)
	corsOrigins []string, // Custom CORS origins (optional, uses defaults if empty)
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Apply CORS middleware first (before other middleware)
	r.Use(middleware.CORSMiddleware(corsOrigins))

	// Apply security headers globally
	r.Use(middleware.SecurityHeadersMiddleware())

	// Apply IP whitelist if configured
	if len(ipWhitelist) > 0 {
		r.Use(middleware.IPWhitelistMiddleware(ipWhitelist))
	}

	// Apply error logging middleware (after other middleware, before routes)
	if postgresClient != nil {
		errorLogRepo := storage.NewErrorLogRepository(postgresClient.Pool())
		r.Use(middleware.ErrorLoggingMiddleware(errorLogRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger()))
	}

	// Health check (no auth required)
	healthHandler := handlers.NewHealthHandler()
	r.GET("/health", healthHandler.HandleHealth)

	// Swagger UI documentation (no auth required, but supports access_token query param)
	swaggerHandler := handlers.NewSwaggerHandler(jwtService)
	r.GET("/swagger", swaggerHandler.HandleSwaggerUI)
	r.GET("/swagger.json", swaggerHandler.HandleSwaggerJSON)

	// Prometheus metrics endpoint (no auth required)
	r.GET("/metrics", handlers.HandleMetrics)

	// Error logging endpoint (no auth required, but rate limited)
	if postgresClient != nil {
		errorLogRepo := storage.NewErrorLogRepository(postgresClient.Pool())
		errorLogHandler := handlers.NewErrorLogHandler(errorLogRepo)
		r.POST("/api/errors/log", errorLogHandler.HandleLogError)
	}

	// Auth routes (no auth required for register/login)
	if postgresClient != nil && jwtService != nil {
		userRepo := storage.NewUserRepository(postgresClient, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())

		// Convert emailService to *email.EmailService if not nil
		var emailSvc *email.EmailService
		if emailService != nil {
			if svc, ok := emailService.(*email.EmailService); ok {
				emailSvc = svc
			}
		}

		authHandler := handlers.NewAuthHandler(userRepo, jwtService, emailSvc, authRateLimiter, frontendURL, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())

		auth := r.Group("/auth")
		
		// OAuth routes
		if oauthService != nil {
			oauthHandler := handlers.NewOAuthHandler(oauthService, jwtService, frontendURL)
			auth.GET("/google", oauthHandler.HandleGoogleAuth)
			auth.GET("/google/callback", oauthHandler.HandleGoogleCallback)
			auth.GET("/x", oauthHandler.HandleXAuth)
			auth.GET("/x/callback", oauthHandler.HandleXCallback)
		}
		
		auth.POST("/register", authHandler.HandleRegister)

		// Login with progressive rate limiting (max 5 attempts before 15min block)
		loginGroup := auth.Group("")
		if authRateLimiter != nil {
			loginGroup.Use(middleware.AuthRateLimitMiddleware(authRateLimiter, 5))
		}
		loginGroup.POST("/login", authHandler.HandleLogin)

		auth.POST("/logout", authHandler.HandleLogout)

		// Password reset with progressive rate limiting (max 5 attempts before 15min block)
		passwordResetGroup := auth.Group("")
		if authRateLimiter != nil {
			passwordResetGroup.Use(middleware.AuthRateLimitMiddleware(authRateLimiter, 5))
		}
		passwordResetGroup.POST("/password-reset", authHandler.HandlePasswordResetRequest)
		passwordResetGroup.POST("/password-reset/confirm", authHandler.HandlePasswordResetConfirm)

		// Email verification with progressive rate limiting (max 5 attempts before 15min block)
		verifyGroup := auth.Group("")
		if authRateLimiter != nil {
			verifyGroup.Use(middleware.AuthRateLimitMiddleware(authRateLimiter, 5))
		}
		verifyGroup.POST("/verify-email", authHandler.HandleVerifyEmail)
		verifyGroup.POST("/resend-verification", authHandler.HandleResendVerification)

		// Protected auth routes (require JWT)
		authProtected := auth.Group("")
		authProtected.Use(middleware.JWTAuthMiddleware(jwtService))
		userHandler := handlers.NewUserHandler(userRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		authProtected.GET("/profile", authHandler.HandleProfile)
		authProtected.PUT("/profile", userHandler.HandleUpdateProfile)
		authProtected.PUT("/profile/password", userHandler.HandleChangePassword)
		authProtected.POST("/refresh", authHandler.HandleRefresh)

		// API key management (user routes - users manage their own keys)
		if apiKeyServiceV2 != nil {
			apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyServiceV2)
			authProtected.POST("/api-keys", apiKeyHandler.CreateAPIKey)
			authProtected.GET("/api-keys", apiKeyHandler.ListAPIKeys)
			authProtected.DELETE("/api-keys/:id", apiKeyHandler.RevokeAPIKey)
		}

		// Provider key management (user routes - users manage their own provider keys BYOK)
		if providerKeyService != nil {
			providerKeyHandler := handlers.NewProviderKeyHandler(providerKeyService)
			authProtected.POST("/provider-keys", providerKeyHandler.AddProviderKey)
			authProtected.GET("/provider-keys", providerKeyHandler.ListProviderKeys)
			authProtected.PUT("/provider-keys/:provider", providerKeyHandler.UpdateProviderKey)
			authProtected.DELETE("/provider-keys/:provider", providerKeyHandler.DeleteProviderKey)
			authProtected.POST("/provider-keys/:provider/test", providerKeyHandler.TestProviderKey)
		}

		// Conversation management endpoints
		var convRepo *storage.ConversationRepository
		if postgresClient != nil {
			convRepo = storage.NewConversationRepository(postgresClient.Pool())
			convHandler := handlers.NewConversationHandler(convRepo)
			authProtected.GET("/conversations", convHandler.ListConversations)
			authProtected.POST("/conversations", convHandler.CreateConversation)
			authProtected.GET("/conversations/:id", convHandler.GetConversation)
			authProtected.PUT("/conversations/:id", convHandler.UpdateConversation)
			authProtected.DELETE("/conversations/:id", convHandler.DeleteConversation)
		}

		// Frontend chat endpoint (JWT authentication for direct UI usage)
		// This allows users to chat directly without creating an API key first
		chatHandler := handlers.NewChatHandler(router, requestRepo, convRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		authProtected.POST("/chat", chatHandler.HandleChat)

		// Frontend analytics endpoints (JWT authentication for direct UI usage)
		if requestRepo != nil {
			analyticsHandler := handlers.NewAnalyticsHandler(requestRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
			authProtected.GET("/analytics/usage", analyticsHandler.GetUsageStats)
			authProtected.GET("/analytics/requests", analyticsHandler.GetRequests)
		}

		// Frontend tunnel endpoints (JWT authentication for direct UI usage)
		// Note: postgresClient is already checked in outer if condition
		tunnelRepo := tunnel.NewTunnelRepository(postgresClient.Pool(), zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		tunnelHandler := handlers.NewTunnelHandler(tunnelRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		authProtected.GET("/tunnels", tunnelHandler.ListTunnels)
		authProtected.GET("/tunnels/stats", tunnelHandler.GetTunnelStats)
		authProtected.GET("/tunnels/:id", tunnelHandler.GetTunnel)
		authProtected.POST("/tunnels/:id/disconnect", tunnelHandler.DisconnectTunnel)
		authProtected.POST("/tunnels/:id/associate", tunnelHandler.AssociateTunnel)
	}

	// Initialize routing handler (needed for both user and admin routes)
	var settingsRepo *storage.SystemSettingsRepository
	var userRepo *storage.UserRepository
	var routingHandler *handlers.RoutingHandler
	if postgresClient != nil {
		settingsRepo = storage.NewSystemSettingsRepository(postgresClient.Pool())
		userRepo = storage.NewUserRepository(postgresClient, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		
		// Set routing strategy services in router
		router.SetRoutingStrategyService(settingsRepo)
		router.SetUserRoutingStrategyService(userRepo)
		
		routingHandler = handlers.NewRoutingHandler(router, settingsRepo, userRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		
		// User routing strategy endpoints (user-facing)
		if jwtService != nil {
			authProtected := r.Group("/auth")
			authProtected.Use(middleware.JWTAuthMiddleware(jwtService))
			authProtected.GET("/routing/strategy", routingHandler.GetUserRoutingStrategy)
			authProtected.PUT("/routing/strategy", routingHandler.SetUserRoutingStrategy)
			authProtected.DELETE("/routing/strategy", routingHandler.ClearUserRoutingStrategy)
			
			// User custom routing rules endpoints
			if postgresClient != nil {
				customRulesRepo := storage.NewCustomRoutingRulesRepository(postgresClient.Pool())
				userCustomRulesHandler := handlers.NewCustomRulesHandler(router, customRulesRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
				authProtected.GET("/routing/custom-rules", userCustomRulesHandler.GetCustomRules)
				authProtected.POST("/routing/custom-rules", userCustomRulesHandler.SetCustomRules)
			}
		}
	}

	// API routes (require API key authentication)
	api := r.Group("/v1")

	// Use database-backed auth (required)
	if apiKeyServiceV2 != nil {
		api.Use(middleware.AuthMiddleware(apiKeyServiceV2))

		// Add rate limiting
		api.Use(middleware.RateLimitMiddleware(rateLimiter, func(identifier string) (int, int) {
			// Get limits from API key record in context
			// Default limits if not found
			return 60, 10000
		}))
	} else {
		// Database-backed API key service is required for API routes
		// If not available, API routes will return 401
		api.Use(func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "API key service not available - database connection required",
			})
			c.Abort()
		})
	}

	chatHandler := handlers.NewChatHandler(router, requestRepo, nil, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
	api.POST("/chat", chatHandler.HandleChat)

	// Analytics endpoints for usage tracking
	if requestRepo != nil {
		analyticsHandler := handlers.NewAnalyticsHandler(requestRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		api.GET("/analytics/usage", analyticsHandler.GetUsageStats)
		api.GET("/analytics/requests", analyticsHandler.GetRequests)
	}

	// Tunnel endpoints (require API key authentication)
	if postgresClient != nil {
		tunnelRepo := tunnel.NewTunnelRepository(postgresClient.Pool(), zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		tunnelHandler := handlers.NewTunnelHandler(tunnelRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		api.GET("/tunnels", tunnelHandler.ListTunnels)
		// Note: /tunnels/stats moved to admin routes - admin only
		api.GET("/tunnels/:id", tunnelHandler.GetTunnel)
		api.POST("/tunnels/:id/disconnect", tunnelHandler.DisconnectTunnel)
	}

	// Provider endpoints
	providerHandler := handlers.NewProviderHandler(router, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
	api.GET("/providers", providerHandler.ListProviders)
	api.GET("/providers/:name/health", providerHandler.GetProviderHealth)

	// Routing endpoints for cost estimation and latency stats
	if routingHandler != nil {
		api.POST("/routing/estimate-cost", routingHandler.GetCostEstimate)
		api.GET("/routing/latency", routingHandler.GetLatencyStats)
	}

	// Admin routes (require JWT authentication and admin role)
	if jwtService != nil {
		admin := r.Group("/admin")
		admin.Use(middleware.JWTAuthMiddleware(jwtService))
		admin.Use(middleware.AdminMiddleware()) // Require admin role

		// Admin routing endpoints for strategy management
		if settingsRepo != nil {
			admin.POST("/routing/strategy", routingHandler.SetRoutingStrategy)
			admin.GET("/routing/strategy", routingHandler.GetRoutingStrategy)
			admin.POST("/routing/strategy/lock", routingHandler.SetRoutingStrategyLock)

			// Custom routing rules (admin only)
			if postgresClient != nil {
				customRulesRepo := storage.NewCustomRoutingRulesRepository(postgresClient.Pool())
				customRulesHandler := handlers.NewCustomRulesHandler(router, customRulesRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
				admin.GET("/routing/custom-rules", customRulesHandler.GetCustomRules)
				admin.POST("/routing/custom-rules", customRulesHandler.SetCustomRules)
			}
		}

		// Tunnel statistics (admin only - shows all tunnels)
		if postgresClient != nil {
			tunnelRepo := tunnel.NewTunnelRepository(postgresClient.Pool(), zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
			tunnelHandler := handlers.NewTunnelHandler(tunnelRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
			admin.GET("/tunnels/stats", tunnelHandler.GetTunnelStats)
		}

		// Error logs management (admin only)
		if postgresClient != nil {
			errorLogRepo := storage.NewErrorLogRepository(postgresClient.Pool())
			errorLogHandler := handlers.NewErrorLogHandler(errorLogRepo)
			admin.GET("/errors", errorLogHandler.HandleGetErrorLogs)
			admin.PATCH("/errors/:id/resolve", errorLogHandler.HandleMarkResolved)
		}

		// Email testing (admin only)
		if emailService != nil {
			if svc, ok := emailService.(*email.EmailService); ok {
				emailTestHandler := handlers.NewEmailTestHandler(svc, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
				admin.GET("/email/config", emailTestHandler.HandleGetEmailConfig)
				admin.POST("/email/test", emailTestHandler.HandleTestEmail)
			}
		}

		// User management (admin only)
		if postgresClient != nil {
			adminUserRepo := storage.NewUserRepository(postgresClient, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
			userHandler := handlers.NewUserHandler(adminUserRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
			admin.GET("/users", userHandler.HandleListUsers)
			admin.PUT("/users/:id/roles", userHandler.HandleUpdateUserRoles)
		}
	}

	return r
}
