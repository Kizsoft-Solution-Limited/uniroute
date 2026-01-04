package api

import (
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/api/handlers"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/api/middleware"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/gateway"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// SetupRouter sets up the API routes
func SetupRouter(
	router *gateway.Router,
	apiKeyService *security.APIKeyService, // Phase 1 (in-memory)
	apiKeyServiceV2 *security.APIKeyServiceV2, // Phase 2 (database)
	jwtService *security.JWTService,
	rateLimiter *security.RateLimiter,
	ipWhitelist []string,
	requestRepo *storage.RequestRepository, // Phase 5 (analytics)
	providerKeyService *security.ProviderKeyService, // BYOK: Provider key service
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Apply security headers globally
	r.Use(middleware.SecurityHeadersMiddleware())

	// Apply IP whitelist if configured
	if len(ipWhitelist) > 0 {
		r.Use(middleware.IPWhitelistMiddleware(ipWhitelist))
	}

	// Health check (no auth required)
	healthHandler := handlers.NewHealthHandler()
	r.GET("/health", healthHandler.HandleHealth)

	// Phase 5: Prometheus metrics endpoint (no auth required)
	r.GET("/metrics", handlers.HandleMetrics)

	// API routes (require API key authentication)
	api := r.Group("/v1")

	// Use Phase 2 auth if available, otherwise fall back to Phase 1
	if apiKeyServiceV2 != nil {
		api.Use(middleware.AuthMiddlewareV2(apiKeyServiceV2))

		// Add rate limiting
		api.Use(middleware.RateLimitMiddleware(rateLimiter, func(identifier string) (int, int) {
			// Get limits from API key record in context
			// Default limits if not found
			return 60, 10000
		}))
	} else {
		// Phase 1: Simple API key auth
		api.Use(middleware.AuthMiddleware(apiKeyService))
	}

	chatHandler := handlers.NewChatHandler(router, requestRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
	api.POST("/chat", chatHandler.HandleChat)

	// Phase 5: Analytics endpoints
	if requestRepo != nil {
		analyticsHandler := handlers.NewAnalyticsHandler(requestRepo, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
		api.GET("/analytics/usage", analyticsHandler.GetUsageStats)
		api.GET("/analytics/requests", analyticsHandler.GetRequests)
	}

	// Provider endpoints
	providerHandler := handlers.NewProviderHandler(router, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
	api.GET("/providers", providerHandler.ListProviders)
	api.GET("/providers/:name/health", providerHandler.GetProviderHealth)

	// Phase 4: Routing endpoints
	routingHandler := handlers.NewRoutingHandler(router, zerolog.New(gin.DefaultWriter).With().Timestamp().Logger())
	api.POST("/routing/estimate-cost", routingHandler.GetCostEstimate)
	api.GET("/routing/latency", routingHandler.GetLatencyStats)

	// Admin routes (require JWT authentication)
	if jwtService != nil {
		admin := r.Group("/admin")
		admin.Use(middleware.JWTAuthMiddleware(jwtService))

		apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyServiceV2)
		admin.POST("/api-keys", apiKeyHandler.CreateAPIKey)
		admin.GET("/api-keys", apiKeyHandler.ListAPIKeys)
		admin.DELETE("/api-keys/:id", apiKeyHandler.RevokeAPIKey)

		// Phase 4: Admin routing endpoints
		admin.POST("/routing/strategy", routingHandler.SetRoutingStrategy)
		admin.GET("/routing/strategy", routingHandler.GetRoutingStrategy)

		// BYOK: Provider key management endpoints
		if providerKeyService != nil {
			providerKeyHandler := handlers.NewProviderKeyHandler(providerKeyService)
			admin.POST("/provider-keys", providerKeyHandler.AddProviderKey)
			admin.GET("/provider-keys", providerKeyHandler.ListProviderKeys)
			admin.PUT("/provider-keys/:provider", providerKeyHandler.UpdateProviderKey)
			admin.DELETE("/provider-keys/:provider", providerKeyHandler.DeleteProviderKey)
			admin.POST("/provider-keys/:provider/test", providerKeyHandler.TestProviderKey)
		}
	}

	return r
}
