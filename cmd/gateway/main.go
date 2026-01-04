package main

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/api"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/config"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/gateway"
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

	// Phase 1: Initialize in-memory API key service (fallback)
	apiKeyService := security.NewAPIKeyService(cfg.APIKeySecret)

	// Phase 2: Initialize database and Redis (optional)
	var apiKeyServiceV2 *security.APIKeyServiceV2
	var jwtService *security.JWTService
	var rateLimiter *security.RateLimiter
	var postgresClient *storage.PostgresClient

	// Try to initialize Phase 2 services if configured
	if cfg.DatabaseURL != "" && cfg.RedisURL != "" {
		log.Info().Msg("Initializing Phase 2 services (database & Redis)...")

		// Initialize Redis
		redisClient, err := storage.NewRedisClient(cfg.RedisURL, log)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to Redis, continuing without rate limiting")
		} else {
			rateLimiter = security.NewRateLimiter(redisClient)
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

	// Generate a default API key for testing (Phase 1 fallback)
	if apiKeyServiceV2 == nil {
		defaultKey, err := apiKeyService.GenerateAPIKey()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to generate default API key")
		}
		log.Info().Str("api_key", defaultKey).Msg("Generated default API key (save this!)")
	}

	// Initialize router
	router := gateway.NewRouter()

	// Register local LLM provider (always available)
	localProvider := providers.NewLocalProvider(cfg.OllamaBaseURL, log)
	router.RegisterProvider(localProvider)
	log.Info().
		Str("provider", "local").
		Str("base_url", cfg.OllamaBaseURL).
		Msg("Registered local LLM provider")

	// Phase 3: Register cloud providers (if API keys are configured)
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

	// Phase 5: Initialize request repository if database is available
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

	// Setup API routes (with Phase 2 & 5 services if available)
	httpRouter := api.SetupRouter(
		router,
		apiKeyService,   // Phase 1 fallback
		apiKeyServiceV2, // Phase 2 (database)
		jwtService,      // Phase 2 (JWT)
		rateLimiter,     // Phase 2 (rate limiting)
		cfg.IPWhitelist, // Phase 2 (IP whitelist)
		requestRepo,     // Phase 5 (analytics)
		providerKeyService, // BYOK: Provider key service
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
