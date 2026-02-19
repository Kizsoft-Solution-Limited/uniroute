package storage

import (
	"time"

	"github.com/google/uuid"
)

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

type PasswordResetToken struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	Token     string     `db:"token"`
	ExpiresAt time.Time  `db:"expires_at"`
	Used      bool       `db:"used"`
	CreatedAt time.Time  `db:"created_at"`
}

type EmailVerificationToken struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	Token     string     `db:"token"`
	ExpiresAt time.Time  `db:"expires_at"`
	Used      bool       `db:"used"`
	CreatedAt time.Time  `db:"created_at"`
}

type UserProviderKey struct {
	ID              uuid.UUID  `db:"id"`
	UserID          uuid.UUID  `db:"user_id"`
	Provider        string     `db:"provider"` // 'openai', 'anthropic', 'google'
	APIKeyEncrypted string     `db:"api_key_encrypted"` // Encrypted provider API key
	IsActive        bool       `db:"is_active"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

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

type Conversation struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Title     *string   `json:"title" db:"title"`
	Model     *string   `json:"model" db:"model"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Message struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	ConversationID uuid.UUID              `json:"conversation_id" db:"conversation_id"`
	Role           string                 `json:"role" db:"role"`
	Content        interface{}            `json:"content" db:"content"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
}
