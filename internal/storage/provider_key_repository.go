package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProviderKeyRepository struct {
	pool *pgxpool.Pool
}

func NewProviderKeyRepository(pool *pgxpool.Pool) *ProviderKeyRepository {
	return &ProviderKeyRepository{
		pool: pool,
	}
}

func (r *ProviderKeyRepository) Create(ctx context.Context, key *UserProviderKey) error {
	query := `
		INSERT INTO user_provider_keys (id, user_id, provider, api_key_encrypted, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, provider) 
		DO UPDATE SET 
			api_key_encrypted = EXCLUDED.api_key_encrypted,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		key.ID,
		key.UserID,
		key.Provider,
		key.APIKeyEncrypted,
		key.IsActive,
		key.CreatedAt,
		key.UpdatedAt,
	)

	return err
}

func (r *ProviderKeyRepository) FindByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*UserProviderKey, error) {
	query := `
		SELECT id, user_id, provider, api_key_encrypted, is_active, created_at, updated_at
		FROM user_provider_keys
		WHERE user_id = $1 AND provider = $2 AND is_active = true
	`

	var key UserProviderKey
	err := r.pool.QueryRow(ctx, query, userID, provider).Scan(
		&key.ID,
		&key.UserID,
		&key.Provider,
		&key.APIKeyEncrypted,
		&key.IsActive,
		&key.CreatedAt,
		&key.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func (r *ProviderKeyRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*UserProviderKey, error) {
	query := `
		SELECT id, user_id, provider, api_key_encrypted, is_active, created_at, updated_at
		FROM user_provider_keys
		WHERE user_id = $1 AND is_active = true
		ORDER BY provider
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*UserProviderKey
	for rows.Next() {
		var key UserProviderKey
		if err := rows.Scan(
			&key.ID,
			&key.UserID,
			&key.Provider,
			&key.APIKeyEncrypted,
			&key.IsActive,
			&key.CreatedAt,
			&key.UpdatedAt,
		); err != nil {
			return nil, err
		}
		keys = append(keys, &key)
	}

	return keys, rows.Err()
}

func (r *ProviderKeyRepository) Update(ctx context.Context, key *UserProviderKey) error {
	query := `
		UPDATE user_provider_keys
		SET api_key_encrypted = $1, is_active = $2, updated_at = $3
		WHERE id = $4 AND user_id = $5
	`

	_, err := r.pool.Exec(ctx, query,
		key.APIKeyEncrypted,
		key.IsActive,
		key.UpdatedAt,
		key.ID,
		key.UserID,
	)

	return err
}

func (r *ProviderKeyRepository) Delete(ctx context.Context, userID uuid.UUID, provider string) error {
	query := `
		UPDATE user_provider_keys
		SET is_active = false, updated_at = NOW()
		WHERE user_id = $1 AND provider = $2
	`

	_, err := r.pool.Exec(ctx, query, userID, provider)
	return err
}

func (r *ProviderKeyRepository) DeleteByID(ctx context.Context, userID uuid.UUID, keyID uuid.UUID) error {
	query := `
		UPDATE user_provider_keys
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`

	_, err := r.pool.Exec(ctx, query, keyID, userID)
	return err
}

