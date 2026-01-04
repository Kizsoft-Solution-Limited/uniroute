package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure APIKeyRepository implements APIKeyRepositoryInterface
var _ APIKeyRepositoryInterface = (*APIKeyRepository)(nil)

// APIKeyRepository handles API key database operations
type APIKeyRepository struct {
	pool *pgxpool.Pool
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(pool *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{
		pool: pool,
	}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(ctx context.Context, key *APIKey) error {
	query := `
		INSERT INTO api_keys (id, user_id, lookup_hash, verification_hash, name, rate_limit_per_minute, rate_limit_per_day, expires_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		key.ID,
		key.UserID,
		key.LookupHash,
		key.VerificationHash,
		key.Name,
		key.RateLimitPerMinute,
		key.RateLimitPerDay,
		key.ExpiresAt,
		key.IsActive,
	)

	return err
}

// FindByLookupHash finds an API key by its lookup hash (SHA256)
func (r *APIKeyRepository) FindByLookupHash(ctx context.Context, lookupHash string) (*APIKey, error) {
	query := `
		SELECT id, user_id, lookup_hash, verification_hash, name, rate_limit_per_minute, rate_limit_per_day, created_at, expires_at, is_active
		FROM api_keys
		WHERE lookup_hash = $1 AND is_active = true
	`

	var key APIKey
	err := r.pool.QueryRow(ctx, query, lookupHash).Scan(
		&key.ID,
		&key.UserID,
		&key.LookupHash,
		&key.VerificationHash,
		&key.Name,
		&key.RateLimitPerMinute,
		&key.RateLimitPerDay,
		&key.CreatedAt,
		&key.ExpiresAt,
		&key.IsActive,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Check if expired
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, nil
	}

	return &key, nil
}

// ListByUserID lists all API keys for a user
func (r *APIKeyRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*APIKey, error) {
	query := `
		SELECT id, user_id, lookup_hash, verification_hash, name, rate_limit_per_minute, rate_limit_per_day, created_at, expires_at, is_active
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		var key APIKey
		err := rows.Scan(
			&key.ID,
			&key.UserID,
			&key.LookupHash,
			&key.VerificationHash,
			&key.Name,
			&key.RateLimitPerMinute,
			&key.RateLimitPerDay,
			&key.CreatedAt,
			&key.ExpiresAt,
			&key.IsActive,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, &key)
	}

	return keys, rows.Err()
}

// Update updates an API key
func (r *APIKeyRepository) Update(ctx context.Context, key *APIKey) error {
	query := `
		UPDATE api_keys
		SET name = $1, rate_limit_per_minute = $2, rate_limit_per_day = $3, expires_at = $4, is_active = $5
		WHERE id = $6
	`

	_, err := r.pool.Exec(ctx, query,
		key.Name,
		key.RateLimitPerMinute,
		key.RateLimitPerDay,
		key.ExpiresAt,
		key.IsActive,
		key.ID,
	)

	return err
}

// Delete deletes an API key (soft delete by setting is_active = false)
func (r *APIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE api_keys
		SET is_active = false
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}
