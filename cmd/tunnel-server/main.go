package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/config"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/logger"
	"github.com/rs/zerolog"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	defaultPort := 8080
	if p := os.Getenv("PORT"); p != "" {
		var n int
		if _, err := fmt.Sscanf(p, "%d", &n); err == nil && n > 0 {
			defaultPort = n
		}
	}
	var (
		port     = flag.Int("port", defaultPort, "Port to run tunnel server on")
		env      = flag.String("env", getEnv("ENV", "development"), "Environment (development/production)")
		logLevel = flag.String("log-level", getEnv("LOG_LEVEL", "info"), "Log level (debug/info/warn/error)")
	)
	flag.Parse()

	var log zerolog.Logger
	if *env == "development" {
		log = logger.NewDebug()
	} else {
		log = logger.New()
	}

	switch *logLevel {
	case "debug":
		log = log.Level(zerolog.DebugLevel)
	case "warn":
		log = log.Level(zerolog.WarnLevel)
	case "error":
		log = log.Level(zerolog.ErrorLevel)
	default:
		log = log.Level(zerolog.InfoLevel)
	}

	log.Info().
		Int("port", *port).
		Str("environment", *env).
		Msg("Starting UniRoute Tunnel Server")

	cfg := config.Load()
	server := tunnel.NewTunnelServer(*port, log, cfg.TunnelOrigins)

	baseDomain := getEnv("TUNNEL_BASE_DOMAIN", "")
	if baseDomain != "" {
		domainManager := tunnel.NewDomainManager(baseDomain, log)
		server.SetDomainManager(domainManager)
		log.Info().Str("base_domain", baseDomain).Msg("Domain manager configured")
	}

	if cfg.DatabaseURL != "" {
		postgresClient, err := storage.NewPostgresClient(cfg.DatabaseURL, log)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to PostgreSQL, continuing without database")
		} else {
			defer postgresClient.Close()
			repo := tunnel.NewTunnelRepository(postgresClient.Pool(), log)
			server.SetRepository(repo)
			log.Info().Msg("Database connected, request logging enabled")

			if cfg.JWTSecret != "" && cfg.JWTSecret != "change-me-in-production-jwt-secret-min-32-chars" {
				jwtService := security.NewJWTService(cfg.JWTSecret)
				server.SetJWTValidator(func(tokenString string) (string, error) {
					claims, err := jwtService.ValidateToken(tokenString)
					if err != nil {
						return "", err
					}
					return claims.UserID, nil
				})
				log.Info().Msg("JWT validator configured - tunnels will be auto-associated with authenticated users")
			} else {
				// SECURITY: In production, fail if JWT_SECRET is not set
				if cfg.Environment == "production" {
					log.Fatal().Msg("SECURITY ERROR: JWT_SECRET must be set in production! Tunnels require JWT validation for user association. Set JWT_SECRET environment variable.")
				}
				log.Warn().Msg("JWT_SECRET not configured or using default - tunnels will NOT be auto-associated with users. Set JWT_SECRET environment variable to enable.")
			}

			if cfg.APIKeySecret != "" && cfg.APIKeySecret != "change-me-in-production" {
				apiKeyRepo := storage.NewAPIKeyRepository(postgresClient.Pool())
				apiKeyService := security.NewAPIKeyServiceV2(apiKeyRepo, cfg.APIKeySecret)

				server.SetAPIKeyValidatorWithLimits(func(ctx context.Context, apiKey string) (string, int, int, error) {
					keyRecord, err := apiKeyService.ValidateAPIKey(ctx, apiKey)
					if err != nil || keyRecord == nil {
						return "", 0, 0, fmt.Errorf("invalid API key: %w", err)
					}
					return keyRecord.UserID.String(), keyRecord.RateLimitPerMinute, keyRecord.RateLimitPerDay, nil
				})

				server.SetAPIKeyValidator(func(ctx context.Context, apiKey string) (string, error) {
					keyRecord, err := apiKeyService.ValidateAPIKey(ctx, apiKey)
					if err != nil || keyRecord == nil {
						return "", fmt.Errorf("invalid API key: %w", err)
					}
					return keyRecord.UserID.String(), nil
				})
				log.Info().Msg("API key validator configured with rate limits - tunnels will use API key's rate limits")
			} else {
				// SECURITY: In production, fail if API_KEY_SECRET is not set
				if cfg.Environment == "production" {
					log.Fatal().Msg("SECURITY ERROR: API_KEY_SECRET must be set in production! Tunnels require API key validation for user association. Set API_KEY_SECRET environment variable.")
				}
				log.Warn().Msg("API_KEY_SECRET not configured or using default - API key authentication will NOT work. Set API_KEY_SECRET environment variable to enable.")
			}
		}
	}

	if cfg.RedisURL != "" {
		redisClient, err := storage.NewRedisClient(cfg.RedisURL, log)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to Redis, using in-memory rate limiting")
		} else {
			defer redisClient.Close()
			redisRateLimiter := tunnel.NewRedisRateLimiter(redisClient, log)
			server.SetRateLimiter(redisRateLimiter)
			server.SetStatsRedis(redisClient)
			log.Info().Msg("Redis connected, distributed rate limiting and stats enabled")
		}
	}

	if err := server.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start tunnel server")
		os.Exit(1)
	}
}
