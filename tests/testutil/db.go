package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	config := GetTestConfig(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, config.PostgresURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return pool
}

// CleanupTestDB cleans up test database
func CleanupTestDB(t *testing.T, pool *pgxpool.Pool) {
	if pool == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Truncate all tables (in reverse order of dependencies)
	tables := []string{
		"error_logs",
		"email_verification_tokens",
		"password_reset_tokens",
		"users",
		"api_keys",
		"provider_keys",
		"requests",
		"tunnel_connections",
		"tunnel_domains",
	}

	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			// Table might not exist, that's okay
			t.Logf("Warning: Could not truncate table %s: %v", table, err)
		}
	}

	pool.Close()
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, pool *pgxpool.Pool, email, password, name string) uuid.UUID {
	ctx := context.Background()

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	var userID uuid.UUID
	err = pool.QueryRow(ctx,
		`INSERT INTO users (id, email, password_hash, name, email_verified, roles, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW())
		 RETURNING id`,
		email, string(hashedPassword), name, true, []string{"user"},
	).Scan(&userID)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}

// CreateTestAPIKey creates a test API key
func CreateTestAPIKey(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID, name string) uuid.UUID {
	ctx := context.Background()

	// Generate a test API key
	testKey := "ur_test_" + uuid.New().String()

	// Hash the key using bcrypt (as done in the actual service)
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(testKey), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash API key: %v", err)
	}

	var apiKeyID uuid.UUID
	err = pool.QueryRow(ctx,
		`INSERT INTO api_keys (id, user_id, name, lookup_hash, verification_hash, rate_limit_per_minute, rate_limit_per_day, created_at, is_active)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW(), true)
		 RETURNING id`,
		userID, name, testKey[:32], string(hashedKey), 60, 10000,
	).Scan(&apiKeyID)

	if err != nil {
		t.Fatalf("Failed to create test API key: %v", err)
	}

	return apiKeyID
}

// Note: This file uses pgxpool directly. If you need database/sql compatibility,
// you can add a wrapper function here.
