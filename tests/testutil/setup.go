package testutil

import (
	"os"
	"testing"
)

// TestConfig holds test configuration
type TestConfig struct {
	PostgresURL string
	RedisURL    string
	JWTSecret   string
	FrontendURL string
}

// GetTestConfig loads test configuration from environment or defaults
func GetTestConfig(t *testing.T) *TestConfig {
	config := &TestConfig{
		PostgresURL: getEnv("TEST_POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/uniroute_test?sslmode=disable"),
		RedisURL:    getEnv("TEST_REDIS_URL", "redis://localhost:6379/15"),
		JWTSecret:   getEnv("TEST_JWT_SECRET", "test-secret-key-change-in-production"),
		FrontendURL: getEnv("TEST_FRONTEND_URL", "http://localhost:3000"),
	}

	// Validate required config
	if config.PostgresURL == "" {
		t.Fatal("TEST_POSTGRES_URL is required for integration tests")
	}

	return config
}

// getEnv gets environment variable or returns default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SkipIfShort skips the test if -short flag is set
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
}

// SkipIfNoPostgres skips the test if PostgreSQL is not available
func SkipIfNoPostgres(t *testing.T) {
	// This would check if PostgreSQL is available
	// For now, we'll rely on environment variables
	if os.Getenv("SKIP_POSTGRES_TESTS") == "true" {
		t.Skip("Skipping PostgreSQL tests")
	}
}

// SkipIfNoRedis skips the test if Redis is not available
func SkipIfNoRedis(t *testing.T) {
	// This would check if Redis is available
	// For now, we'll rely on environment variables
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis tests")
	}
}

