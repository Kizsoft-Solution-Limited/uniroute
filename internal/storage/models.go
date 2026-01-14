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
	ID              uuid.UUID  `db:"id"`
	Email           string     `db:"email"`
	Name            string     `db:"name"`
	PasswordHash    string     `db:"password_hash"`
	EmailVerified   bool       `db:"email_verified"`
	Roles           []string   `db:"roles"` // Array of roles: ['user'], ['admin'], or ['user', 'admin']
	RoutingStrategy *string    `db:"routing_strategy"` // User-specific routing strategy (NULL = use default)
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	Token     string     `db:"token"`
	ExpiresAt time.Time  `db:"expires_at"`
	Used      bool       `db:"used"`
	CreatedAt time.Time  `db:"created_at"`
}

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	Token     string     `db:"token"`
	ExpiresAt time.Time  `db:"expires_at"`
	Used      bool       `db:"used"`
	CreatedAt time.Time  `db:"created_at"`
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

// ErrorLog represents a frontend error log
type ErrorLog struct {
	ID         uuid.UUID              `db:"id"`
	UserID     *uuid.UUID             `db:"user_id"` // Nullable
	ErrorType  string                 `db:"error_type"` // 'exception', 'message', 'network', 'server'
	Message    string                 `db:"message"`
	StackTrace *string                `db:"stack_trace"` // Nullable
	URL        *string                `db:"url"` // Nullable
	UserAgent  *string                `db:"user_agent"` // Nullable
	IPAddress  *string                `db:"ip_address"` // Nullable
	Context    map[string]interface{} `db:"context"` // JSONB
	Severity   string                 `db:"severity"` // 'error', 'warning', 'info'
	Resolved   bool                   `db:"resolved"`
	CreatedAt  time.Time              `db:"created_at"`
}
