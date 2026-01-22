package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/api"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/config"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/email"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/gateway"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/oauth"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	var log zerolog.Logger
	if cfg.Environment == "development" {
		log = logger.NewDebug()
	} else {
		log = logger.New()
	}

	log.Info().Msg("Starting UniRoute Gateway...")

	// SECURITY: Warn if using default secrets in production
	if cfg.Environment == "production" {
		if cfg.APIKeySecret == "change-me-in-production" {
			log.Fatal().Msg("SECURITY ERROR: API_KEY_SECRET must be changed from default in production! Set a secure secret via environment variable.")
		}
		if cfg.JWTSecret == "change-me-in-production-jwt-secret-min-32-chars" {
			log.Fatal().Msg("SECURITY ERROR: JWT_SECRET must be changed from default in production! Set a secure secret (min 32 chars) via environment variable.")
		}
	}

	// Initialize in-memory API key service (fallback when database is not available)
	apiKeyService := security.NewAPIKeyService(cfg.APIKeySecret)

	// Initialize database and Redis services (optional, enables advanced features)
	var apiKeyServiceV2 *security.APIKeyServiceV2
	var jwtService *security.JWTService
	var rateLimiter *security.RateLimiter
	var authRateLimiter *security.AuthRateLimiter
	var postgresClient *storage.PostgresClient

	// Initialize database and Redis if configured
	if cfg.DatabaseURL != "" && cfg.RedisURL != "" {
		log.Info().Msg("Initializing database and Redis services...")

		// Initialize Redis
		redisClient, err := storage.NewRedisClient(cfg.RedisURL, log)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to Redis, continuing without rate limiting")
		} else {
			rateLimiter = security.NewRateLimiter(redisClient)
			authRateLimiter = security.NewAuthRateLimiter(redisClient)
			log.Info().Msg("Redis connected - rate limiting enabled")
			// Note: redisClient will be closed when the process exits
		}

		// Initialize PostgreSQL
		postgresClient, err = storage.NewPostgresClient(cfg.DatabaseURL, log)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to PostgreSQL, using in-memory API keys")
		} else {
			// Initialize API key repository
			apiKeyRepo := storage.NewAPIKeyRepository(postgresClient.Pool())
			apiKeyServiceV2 = security.NewAPIKeyServiceV2(apiKeyRepo, cfg.APIKeySecret)
			log.Info().Msg("PostgreSQL connected - database-backed API keys enabled")
			// Note: postgresClient will be closed when the process exits
		}

		// Initialize JWT service
		if cfg.JWTSecret != "" {
			jwtService = security.NewJWTService(cfg.JWTSecret)
			log.Info().Msg("JWT service initialized")
		}
	}

	// Generate a default API key for testing (when database is not available)
	// SECURITY: Only log in development mode, never in production
	if apiKeyServiceV2 == nil {
		defaultKey, err := apiKeyService.GenerateAPIKey()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to generate default API key")
		}
		if cfg.Environment == "development" {
			log.Info().Str("api_key", defaultKey).Msg("Generated default API key (save this!)")
		} else {
			// In production, log to stderr only (not to log files)
			fmt.Fprintf(os.Stderr, "⚠️  Generated default API key: %s (save this securely!)\n", defaultKey)
			log.Info().Msg("Generated default API key (check stderr for the key)")
		}
	}

	// Initialize router
	router := gateway.NewRouter()

	// Load default routing strategy from database if available
	if postgresClient != nil {
		settingsRepo := storage.NewSystemSettingsRepository(postgresClient.Pool())
		ctx := context.Background()
		strategy, err := settingsRepo.GetDefaultRoutingStrategy(ctx)
		if err == nil && strategy != "" {
			strategyType := gateway.StrategyType(strategy)
			router.SetStrategyType(strategyType)
			log.Info().
				Str("strategy", strategy).
				Msg("Loaded default routing strategy from database")

			// If custom strategy, load custom rules
			if strategyType == gateway.StrategyCustom {
				customRulesRepo := storage.NewCustomRoutingRulesRepository(postgresClient.Pool())
				rules, err := customRulesRepo.GetActiveRules(ctx)
				if err == nil && len(rules) > 0 {
					// Convert database rules to gateway routing rules
					costCalculator := router.GetCostCalculator()
					latencyTracker := router.GetLatencyTracker()
					routingRules := make([]gateway.RoutingRule, 0, len(rules))
					for _, rule := range rules {
						// Build condition function
						condition := func(rule *storage.CustomRoutingRule, costCalc *gateway.CostCalculator, latTracker *gateway.LatencyTracker) func(providers.ChatRequest) bool {
							return func(req providers.ChatRequest) bool {
								switch rule.ConditionType {
								case "model":
									if model, ok := rule.ConditionValue["model"].(string); ok {
										return req.Model == model
									}
								case "cost_threshold":
									// Check if estimated cost is below threshold
									if maxCost, ok := rule.ConditionValue["max_cost"].(float64); ok {
										// Estimate cost for the request
										estimatedCost := costCalc.EstimateCost(rule.ProviderName, req.Model, req.Messages)
										return estimatedCost <= maxCost
									}
									return false
								case "latency_threshold":
									// Check if average latency is below threshold
									if maxLatencyMs, ok := rule.ConditionValue["max_latency_ms"].(float64); ok {
										avgLatency := latTracker.GetAverageLatency(rule.ProviderName)
										return avgLatency.Milliseconds() <= int64(maxLatencyMs)
									}
									return false
								}
								return false
							}
						}(rule, costCalculator, latencyTracker)

						routingRule := gateway.RoutingRule{
							Provider:  rule.ProviderName,
							Priority:  rule.Priority,
							Condition: condition,
						}
						routingRules = append(routingRules, routingRule)
					}

					customStrategy := gateway.NewCustomStrategy(routingRules)
					router.SetCustomStrategy(customStrategy)
					log.Info().
						Int("rules_count", len(rules)).
						Msg("Loaded custom routing rules from database")
				} else {
					log.Debug().
						Err(err).
						Msg("No custom routing rules found, using fallback")
				}
			}
		} else {
			log.Debug().
				Err(err).
				Msg("Using default routing strategy (model-based)")
		}
	}

	// Register local LLM provider (always available)
	localProvider := providers.NewLocalProvider(cfg.OllamaBaseURL, log)
	router.RegisterProvider(localProvider)
	log.Info().
		Str("provider", "local").
		Str("base_url", cfg.OllamaBaseURL).
		Msg("Registered local LLM provider")

	// Register cloud providers (if API keys are configured)
	if cfg.OpenAIAPIKey != "" {
		openAIProvider := providers.NewOpenAIProvider(cfg.OpenAIAPIKey, "", log)
		router.RegisterProvider(openAIProvider)
		log.Info().
			Str("provider", "openai").
			Msg("Registered OpenAI provider")
	} else {
		log.Debug().Msg("OpenAI API key not configured, skipping OpenAI provider")
	}

	if cfg.AnthropicAPIKey != "" {
		anthropicProvider := providers.NewAnthropicProvider(cfg.AnthropicAPIKey, "", log)
		router.RegisterProvider(anthropicProvider)
		log.Info().
			Str("provider", "anthropic").
			Msg("Registered Anthropic provider")
	} else {
		log.Debug().Msg("Anthropic API key not configured, skipping Anthropic provider")
	}

	if cfg.GoogleAPIKey != "" {
		googleProvider := providers.NewGoogleProvider(cfg.GoogleAPIKey, "", log)
		router.RegisterProvider(googleProvider)
		log.Info().
			Str("provider", "google").
			Msg("Registered Google provider")
	} else {
		log.Debug().Msg("Google API key not configured, skipping Google provider")
	}

	log.Info().
		Strs("providers", router.ListProviders()).
		Msg("All providers registered")

	// Initialize request repository for analytics and usage tracking
	var requestRepo *storage.RequestRepository
	if postgresClient != nil {
		requestRepo = storage.NewRequestRepository(postgresClient.Pool())
		log.Info().Msg("Request repository initialized - usage tracking enabled")
	}

	// BYOK: Initialize provider key service if database is available
	var providerKeyService *security.ProviderKeyService
	if postgresClient != nil && cfg.ProviderKeyEncryptionKey != "" {
		// Create provider key repository
		providerKeyRepo := storage.NewProviderKeyRepository(postgresClient.Pool())

		// Create provider key service with encryption
		// Use JWT secret as encryption key if provider key encryption key not set
		encryptionKey := cfg.ProviderKeyEncryptionKey
		if encryptionKey == "" {
			encryptionKey = cfg.JWTSecret // Fallback to JWT secret
		}

		service, err := security.NewProviderKeyService(providerKeyRepo, encryptionKey)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize provider key service, BYOK disabled")
		} else {
			providerKeyService = service
			router.SetProviderKeyService(service)
			log.Info().Msg("Provider key service initialized - BYOK enabled")
		}
	}

	// Set server-level provider keys (fallback)
	router.SetServerProviderKeys(gateway.ServerProviderKeys{
		OpenAI:    cfg.OpenAIAPIKey,
		Anthropic: cfg.AnthropicAPIKey,
		Google:    cfg.GoogleAPIKey,
	})

	// Set custom rules service for user-specific custom routing rules
	if postgresClient != nil {
		customRulesRepo := storage.NewCustomRoutingRulesRepository(postgresClient.Pool())
		costCalculator := router.GetCostCalculator()
		latencyTracker := router.GetLatencyTracker()
		customRulesService := gateway.NewCustomRulesServiceAdapter(customRulesRepo, costCalculator, latencyTracker)
		router.SetCustomRulesService(customRulesService)
		log.Info().Msg("Custom rules service initialized - user-specific custom routing rules enabled")
	}

	// Initialize email service (reads config from environment variables)
	emailService := email.NewEmailService(log)

	// Log SMTP configuration status
	smtpConfig := emailService.GetConfig()
	if configured, ok := smtpConfig["configured"].(bool); ok && configured {
		if host, ok := smtpConfig["host"].(string); ok {
			if port, ok := smtpConfig["port"].(int); ok {
				log.Info().
					Str("host", host).
					Int("port", port).
					Msg("Email service initialized")
			}
		}
	} else {
		log.Warn().Msg("SMTP not configured - set SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD environment variables")
	}

	// Initialize OAuth service if configured
	var oauthService *oauth.OAuthService
	if postgresClient != nil && (cfg.GoogleOAuthClientID != "" || cfg.XOAuthClientID != "" || cfg.GithubOAuthClientID != "") {
		userRepo := storage.NewUserRepository(postgresClient, log)
		// Backend URL for OAuth callbacks (OAuth providers redirect here)
		backendURL := cfg.BackendURL
		if backendURL == "" {
			// Auto-detect: use localhost for development, or construct from PORT
			if cfg.Environment == "development" || cfg.Environment == "local" {
				backendURL = fmt.Sprintf("http://localhost:%s", cfg.Port)
			} else {
				// Production: must be set via BACKEND_URL env var
				log.Warn().Msg("BACKEND_URL not set in production - OAuth may not work correctly")
				backendURL = fmt.Sprintf("http://localhost:%s", cfg.Port) // Fallback
			}
		}
		oauthService = oauth.NewOAuthService(
			cfg.GoogleOAuthClientID,
			cfg.GoogleOAuthClientSecret,
			cfg.XOAuthClientID,
			cfg.XOAuthClientSecret,
			cfg.GithubOAuthClientID,
			cfg.GithubOAuthClientSecret,
			backendURL,      // Backend URL for OAuth callbacks
			cfg.FrontendURL, // Frontend URL for final redirect
			userRepo,
		)
		if oauthService.IsGoogleConfigured() {
			log.Info().Str("backend_url", backendURL).Msg("Google OAuth initialized")
		}
		if oauthService.IsXConfigured() {
			log.Info().Str("backend_url", backendURL).Msg("X OAuth initialized")
		}
		if oauthService.IsGithubConfigured() {
			log.Info().Str("backend_url", backendURL).Msg("GitHub OAuth initialized")
		}
	}

	// Setup API routes with all available services
	httpRouter := api.SetupRouter(
		router,
		apiKeyService,      // In-memory API key service (fallback)
		apiKeyServiceV2,    // Database-backed API key service
		jwtService,         // JWT authentication service
		rateLimiter,        // Rate limiting service
		authRateLimiter,    // Progressive rate limiting for auth endpoints
		cfg.IPWhitelist,    // IP whitelist configuration
		requestRepo,        // Request repository for analytics
		providerKeyService, // BYOK: Provider key service
		postgresClient,     // Database client for user repository
		emailService,       // Email service
		cfg.FrontendURL,    // Frontend URL
		oauthService,       // OAuth service
		cfg.CORSOrigins,    // CORS origins from environment
	)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Info().
		Str("address", addr).
		Str("environment", cfg.Environment).
		Msg("Server starting")

	if err := http.ListenAndServe(addr, httpRouter); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
