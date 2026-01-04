package storage

import (
	"time"

	"github.com/google/uuid"
)

// APIKey represents an API key in the database
type APIKey struct {
	ID                 uuid.UUID  `db:"id"`
	UserID             uuid.UUID  `db:"user_id"`
	LookupHash         string     `db:"lookup_hash"`       // SHA256 hash for fast lookup
	VerificationHash   string     `db:"verification_hash"` // bcrypt hash for verification
	Name               string     `db:"name"`
	RateLimitPerMinute int        `db:"rate_limit_per_minute"`
	RateLimitPerDay    int        `db:"rate_limit_per_day"`
	CreatedAt          time.Time  `db:"created_at"`
	ExpiresAt          *time.Time `db:"expires_at"`
	IsActive           bool       `db:"is_active"`
}

// User represents a user in the database
type User struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// UserProviderKey represents a user's provider API key (BYOK)
type UserProviderKey struct {
	ID              uuid.UUID  `db:"id"`
	UserID          uuid.UUID  `db:"user_id"`
	Provider        string     `db:"provider"` // 'openai', 'anthropic', 'google'
	APIKeyEncrypted string     `db:"api_key_encrypted"` // Encrypted provider API key
	IsActive        bool       `db:"is_active"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}
