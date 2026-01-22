package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Port          string
	OllamaBaseURL string
	APIKeySecret  string
	JWTSecret     string
	Environment   string
	DatabaseURL   string
	RedisURL      string
	IPWhitelist   []string
	// Cloud provider API keys (server-level, fallback for BYOK)
	OpenAIAPIKey    string
	AnthropicAPIKey string
	GoogleAPIKey    string
	// BYOK: Encryption key for user provider keys
	ProviderKeyEncryptionKey string
	// Email/SMTP configuration
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	FrontendURL  string
	// OAuth Configuration
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string
	XOAuthClientID          string
	XOAuthClientSecret      string
	GithubOAuthClientID     string
	GithubOAuthClientSecret string
	// Backend URL for OAuth callbacks (optional, auto-detected from PORT if not set)
	BackendURL string
	// Allowed CORS origins (comma-separated, optional)
	CORSOrigins []string
	// Allowed tunnel origins (comma-separated, optional)
	TunnelOrigins []string
}

// Load loads configuration from environment variables
// It automatically loads .env file if it exists
func Load() *Config {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Also try .env.local for local overrides
	_ = godotenv.Load(".env.local")

	ipWhitelist := []string{}
	if whitelist := getEnv("IP_WHITELIST", ""); whitelist != "" {
		// Parse comma-separated IPs
		ips := strings.Split(whitelist, ",")
		for _, ip := range ips {
			if trimmed := strings.TrimSpace(ip); trimmed != "" {
				ipWhitelist = append(ipWhitelist, trimmed)
			}
		}
	}

	return &Config{
		Port:                     getEnv("PORT", "8084"),
		OllamaBaseURL:            getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		APIKeySecret:             getEnv("API_KEY_SECRET", "change-me-in-production"),
		JWTSecret:                getEnv("JWT_SECRET", "change-me-in-production-jwt-secret-min-32-chars"),
		Environment:              getEnv("ENV", "development"),
		DatabaseURL:              getEnv("DATABASE_URL", ""),
		RedisURL:                 getEnv("REDIS_URL", ""),
		IPWhitelist:              ipWhitelist,
		OpenAIAPIKey:             getEnv("OPENAI_API_KEY", ""),
		AnthropicAPIKey:          getEnv("ANTHROPIC_API_KEY", ""),
		GoogleAPIKey:             getEnv("GOOGLE_API_KEY", ""),
		ProviderKeyEncryptionKey: getEnv("PROVIDER_KEY_ENCRYPTION_KEY", ""),
		SMTPHost:                 getEnv("SMTP_HOST", ""),
		SMTPPort:                 getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername:             getEnv("SMTP_USERNAME", ""),
		SMTPPassword:             getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:                 getEnv("SMTP_FROM", ""),
		FrontendURL:              getEnv("FRONTEND_URL", "http://localhost:3002"),
		GoogleOAuthClientID:      getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
		GoogleOAuthClientSecret:  getEnv("GOOGLE_OAUTH_CLIENT_SECRET", ""),
		XOAuthClientID:           getEnv("X_OAUTH_CLIENT_ID", ""),
		XOAuthClientSecret:       getEnv("X_OAUTH_CLIENT_SECRET", ""),
		GithubOAuthClientID:      getEnv("GITHUB_OAUTH_CLIENT_ID", ""),
		GithubOAuthClientSecret:  getEnv("GITHUB_OAUTH_CLIENT_SECRET", ""),
		BackendURL:               getEnv("BACKEND_URL", ""), // Empty = auto-detect from PORT
		CORSOrigins:              parseCORSOrigins(getEnv("CORS_ORIGINS", "")),
		TunnelOrigins:            parseCORSOrigins(getEnv("TUNNEL_ORIGINS", "")),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// parseCORSOrigins parses comma-separated CORS origins from environment variable
func parseCORSOrigins(origins string) []string {
	if origins == "" {
		return []string{} // Empty = use defaults in CORS middleware
	}
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, origin := range parts {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
