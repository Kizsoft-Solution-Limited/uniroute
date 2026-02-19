package api

import (
	"net/http"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/api/handlers"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/api/middleware"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/email"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/gateway"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/oauth"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func SetupRouter(
	router *gateway.Router,
	apiKeyService *security.APIKeyService,
	apiKeyServiceV2 *security.APIKeyServiceV2,
	jwtService *security.JWTService,
	rateLimiter *security.RateLimiter,
	authRateLimiter *security.AuthRateLimiter,
	ipWhitelist []string,
	requestRepo *storage.RequestRepository,
	providerKeyService *security.ProviderKeyService,
	postgresClient *storage.PostgresClient,
	emailService interface{},
	frontendURL string,
	oauthService *oauth.OAuthService,
	corsOrigins []string,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(middleware.CORSMiddleware(corsOrigins))
	r.Use(middleware.SecurityHeadersMiddleware())
	if len(ipWhitelist) > 0 {
		r.Use(middleware.IPWhitelistMiddleware(ipWhitelist))
	}
	if postgresClient != nil {
		errorLogRepo := storage.NewErrorLogRepository(postgresClient.Pool())
		r.Use(middleware.ErrorLoggingMiddleware(errorLogRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger()))
	}

	healthHandler := handlers.NewHealthHandler()
	r.GET("/", func(c *gin.Context) {
		if frontendURL != "" {
			c.Redirect(http.StatusFound, frontendURL)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "UniRoute API", "docs": "/swagger", "health": "/health"})
	})
	r.GET("/health", healthHandler.HandleHealth)
	swaggerHandler := handlers.NewSwaggerHandler(jwtService)
	r.GET("/swagger", swaggerHandler.HandleSwaggerUI)
	r.GET("/swagger.json", swaggerHandler.HandleSwaggerJSON)
	r.GET("/metrics", handlers.HandleMetrics)
	if postgresClient != nil {
		errorLogRepo := storage.NewErrorLogRepository(postgresClient.Pool())
		errorLogHandler := handlers.NewErrorLogHandler(errorLogRepo)
		r.POST("/api/errors/log", errorLogHandler.HandleLogError)
	}

	if postgresClient != nil && jwtService != nil {
		userRepo := storage.NewUserRepository(postgresClient, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		var emailSvc *email.EmailService
		if emailService != nil {
			if svc, ok := emailService.(*email.EmailService); ok {
				emailSvc = svc
			}
		}

		authHandler := handlers.NewAuthHandler(userRepo, jwtService, emailSvc, authRateLimiter, frontendURL, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())

		auth := r.Group("/auth")
		if oauthService != nil {
			oauthHandler := handlers.NewOAuthHandler(oauthService, jwtService, frontendURL)
			auth.GET("/google", oauthHandler.HandleGoogleAuth)
			auth.GET("/google/callback", oauthHandler.HandleGoogleCallback)
			auth.GET("/x", oauthHandler.HandleXAuth)
			auth.GET("/x/callback", oauthHandler.HandleXCallback)
			auth.GET("/github", oauthHandler.HandleGithubAuth)
			auth.GET("/github/callback", oauthHandler.HandleGithubCallback)
		}

		auth.POST("/register", authHandler.HandleRegister)
		loginGroup := auth.Group("")
		if authRateLimiter != nil {
			loginGroup.Use(middleware.AuthRateLimitMiddleware(authRateLimiter, 5))
		}
		loginGroup.POST("/login", authHandler.HandleLogin)
		auth.POST("/logout", authHandler.HandleLogout)
		passwordResetGroup := auth.Group("")
		if authRateLimiter != nil {
			passwordResetGroup.Use(middleware.AuthRateLimitMiddleware(authRateLimiter, 5))
		}
		passwordResetGroup.POST("/password-reset", authHandler.HandlePasswordResetRequest)
		passwordResetGroup.POST("/password-reset/confirm", authHandler.HandlePasswordResetConfirm)
		verifyGroup := auth.Group("")
		if authRateLimiter != nil {
			verifyGroup.Use(middleware.AuthRateLimitMiddleware(authRateLimiter, 5))
		}
		verifyGroup.POST("/verify-email", authHandler.HandleVerifyEmail)
		verifyGroup.POST("/resend-verification", authHandler.HandleResendVerification)
		authProtected := auth.Group("")
		authProtected.Use(middleware.JWTAuthMiddleware(jwtService))
		userHandler := handlers.NewUserHandler(userRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		authProtected.GET("/profile", authHandler.HandleProfile)
		authProtected.PUT("/profile", userHandler.HandleUpdateProfile)
		authProtected.PUT("/profile/password", userHandler.HandleChangePassword)
		authProtected.POST("/refresh", authHandler.HandleRefresh)

		if apiKeyServiceV2 != nil {
			apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyServiceV2)
			authProtected.POST("/api-keys", apiKeyHandler.CreateAPIKey)
			authProtected.GET("/api-keys", apiKeyHandler.ListAPIKeys)
			authProtected.DELETE("/api-keys/:id", apiKeyHandler.RevokeAPIKey)
		}

		if providerKeyService != nil {
			keyValidator := providers.NewKeyValidator(zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
			providerKeyHandler := handlers.NewProviderKeyHandler(providerKeyService, keyValidator)
			authProtected.POST("/provider-keys", providerKeyHandler.AddProviderKey)
			authProtected.GET("/provider-keys", providerKeyHandler.ListProviderKeys)
			authProtected.PUT("/provider-keys/:provider", providerKeyHandler.UpdateProviderKey)
			authProtected.DELETE("/provider-keys/:provider", providerKeyHandler.DeleteProviderKey)
			authProtected.POST("/provider-keys/:provider/test", providerKeyHandler.TestProviderKey)
		}

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

		chatHandler := handlers.NewChatHandler(router, requestRepo, convRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		chatHandler.SetJWTService(jwtService)
		authProtected.POST("/chat", chatHandler.HandleChat)
		authProtected.POST("/chat/stream", chatHandler.HandleChatStream) // SSE streaming endpoint
		authProtected.GET("/chat/ws", chatHandler.HandleChatWebSocket)

		authProviderHandler := handlers.NewProviderHandler(router, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		authProtected.GET("/providers", authProviderHandler.ListProviders)

		if requestRepo != nil {
			analyticsHandler := handlers.NewAnalyticsHandler(requestRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
			authProtected.GET("/analytics/usage", analyticsHandler.GetUsageStats)
			authProtected.GET("/analytics/requests", analyticsHandler.GetRequests)
		}

		tunnelRepo := tunnel.NewTunnelRepository(postgresClient.Pool(), zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		tunnelHandler := handlers.NewTunnelHandler(tunnelRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		authProtected.GET("/tunnels", tunnelHandler.ListTunnels)
		authProtected.GET("/tunnels/stats", tunnelHandler.GetTunnelStats)
		authProtected.GET("/tunnels/:id", tunnelHandler.GetTunnel)
		authProtected.POST("/tunnels/:id/disconnect", tunnelHandler.DisconnectTunnel)
		authProtected.POST("/tunnels/:id/associate", tunnelHandler.AssociateTunnel)
		authProtected.POST("/tunnels/:id/domain", tunnelHandler.SetCustomDomain)
		authProtected.PUT("/tunnels/:id/domain", tunnelHandler.SetCustomDomain)

		domainRepo := storage.NewCustomDomainRepository(postgresClient.Pool())
		domainHandler := handlers.NewDomainHandler(domainRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		domainManager := tunnel.NewDomainManager("", zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		domainHandler.SetDomainManager(domainManager)
		authProtected.GET("/domains", domainHandler.ListDomains)
		authProtected.POST("/domains", domainHandler.CreateDomain)
		authProtected.DELETE("/domains/:id", domainHandler.DeleteDomain)
		authProtected.POST("/domains/:id/verify", domainHandler.VerifyDomain)
	}

	var settingsRepo *storage.SystemSettingsRepository
	var userRepo *storage.UserRepository
	var routingHandler *handlers.RoutingHandler
	if postgresClient != nil {
		settingsRepo = storage.NewSystemSettingsRepository(postgresClient.Pool())
		userRepo = storage.NewUserRepository(postgresClient, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())

		router.SetRoutingStrategyService(settingsRepo)
		router.SetUserRoutingStrategyService(userRepo)

		routingHandler = handlers.NewRoutingHandler(router, settingsRepo, userRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())

		if jwtService != nil {
			authProtected := r.Group("/auth")
			authProtected.Use(middleware.JWTAuthMiddleware(jwtService))
			authProtected.GET("/routing/strategy", routingHandler.GetUserRoutingStrategy)
			authProtected.PUT("/routing/strategy", routingHandler.SetUserRoutingStrategy)
			authProtected.DELETE("/routing/strategy", routingHandler.ClearUserRoutingStrategy)

			if postgresClient != nil {
				customRulesRepo := storage.NewCustomRoutingRulesRepository(postgresClient.Pool())
				userCustomRulesHandler := handlers.NewCustomRulesHandler(router, customRulesRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
				authProtected.GET("/routing/custom-rules", userCustomRulesHandler.GetCustomRules)
				authProtected.POST("/routing/custom-rules", userCustomRulesHandler.SetCustomRules)
			}
		}
	}

	api := r.Group("/v1")
	if apiKeyServiceV2 != nil {
		var apiUserRepo *storage.UserRepository
		if postgresClient != nil {
			apiUserRepo = storage.NewUserRepository(postgresClient, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		}
		api.Use(middleware.AuthMiddleware(apiKeyServiceV2, apiUserRepo))

		api.Use(middleware.RateLimitMiddleware(rateLimiter, func(identifier string) (int, int) {
			return 60, 10000
		}))
	} else {
		api.Use(func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "API key service not available - database connection required",
			})
			c.Abort()
		})
	}

	chatHandler := handlers.NewChatHandler(router, requestRepo, nil, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
	api.POST("/chat", chatHandler.HandleChat)
	api.POST("/chat/stream", chatHandler.HandleChatStream)
	api.GET("/chat/ws", chatHandler.HandleChatWebSocket)

	if requestRepo != nil {
		analyticsHandler := handlers.NewAnalyticsHandler(requestRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		api.GET("/analytics/usage", analyticsHandler.GetUsageStats)
		api.GET("/analytics/requests", analyticsHandler.GetRequests)
	}

	if postgresClient != nil && userRepo != nil {
		userHandler := handlers.NewUserHandler(userRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		api.GET("/user", userHandler.HandleGetUser)
	}

	if postgresClient != nil {
		tunnelRepo := tunnel.NewTunnelRepository(postgresClient.Pool(), zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		tunnelHandler := handlers.NewTunnelHandler(tunnelRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())

		domainManager := tunnel.NewDomainManager("", zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		tunnelHandler.SetDomainManager(domainManager)

		api.GET("/tunnels", tunnelHandler.ListTunnels)
		api.GET("/tunnels/:id", tunnelHandler.GetTunnel)
		api.POST("/tunnels/:id/disconnect", tunnelHandler.DisconnectTunnel)
		api.POST("/tunnels/:id/domain", tunnelHandler.SetCustomDomain)
		api.PUT("/tunnels/:id/domain", tunnelHandler.SetCustomDomain)
	}

	providerHandler := handlers.NewProviderHandler(router, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
	api.GET("/providers", providerHandler.ListProviders)
	api.GET("/providers/:name/health", providerHandler.GetProviderHealth)

	if routingHandler != nil {
		api.POST("/routing/estimate-cost", routingHandler.GetCostEstimate)
		api.GET("/routing/latency", routingHandler.GetLatencyStats)
	}

	if jwtService != nil {
		admin := r.Group("/admin")
		admin.Use(middleware.JWTAuthMiddleware(jwtService))
		admin.Use(middleware.AdminMiddleware())

		if settingsRepo != nil {
			admin.POST("/routing/strategy", routingHandler.SetRoutingStrategy)
			admin.GET("/routing/strategy", routingHandler.GetRoutingStrategy)
			admin.POST("/routing/strategy/lock", routingHandler.SetRoutingStrategyLock)

			if postgresClient != nil {
				customRulesRepo := storage.NewCustomRoutingRulesRepository(postgresClient.Pool())
				customRulesHandler := handlers.NewCustomRulesHandler(router, customRulesRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
				admin.GET("/routing/custom-rules", customRulesHandler.GetCustomRules)
				admin.POST("/routing/custom-rules", customRulesHandler.SetCustomRules)
			}
		}

		if postgresClient != nil {
			tunnelRepo := tunnel.NewTunnelRepository(postgresClient.Pool(), zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
			tunnelHandler := handlers.NewTunnelHandler(tunnelRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
			admin.GET("/tunnels/stats", tunnelHandler.GetTunnelStats)
			admin.GET("/tunnels", tunnelHandler.HandleAdminListTunnels)
			admin.DELETE("/tunnels/:id", tunnelHandler.HandleAdminDeleteTunnel)
			admin.POST("/tunnels/delete", tunnelHandler.HandleAdminDeleteTunnels)
		}

		if postgresClient != nil {
			errorLogRepo := storage.NewErrorLogRepository(postgresClient.Pool())
			errorLogHandler := handlers.NewErrorLogHandler(errorLogRepo)
			admin.GET("/errors", errorLogHandler.HandleGetErrorLogs)
			admin.PATCH("/errors/:id/resolve", errorLogHandler.HandleMarkResolved)
		}

		if emailService != nil {
			if svc, ok := emailService.(*email.EmailService); ok {
				emailTestHandler := handlers.NewEmailTestHandler(svc, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
				admin.GET("/email/config", emailTestHandler.HandleGetEmailConfig)
				admin.POST("/email/test", emailTestHandler.HandleTestEmail)
			}
		}

		if postgresClient != nil {
			adminUserRepo := storage.NewUserRepository(postgresClient, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
			userHandler := handlers.NewUserHandler(adminUserRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
			admin.GET("/users", userHandler.HandleListUsers)
			admin.PUT("/users/:id/roles", userHandler.HandleUpdateUserRoles)
			admin.DELETE("/users/:id", userHandler.HandleDeleteUser)
			admin.POST("/users/delete", userHandler.HandleDeleteUsers)
		}
	}

	return r
}
