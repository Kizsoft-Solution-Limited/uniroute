package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SystemSetting struct {
	ID          int        `db:"id"`
	Key         string     `db:"key"`
	Value       string     `db:"value"`
	Description *string    `db:"description"`
	UpdatedAt   time.Time  `db:"updated_at"`
	UpdatedBy   *uuid.UUID `db:"updated_by"`
	CreatedAt   time.Time  `db:"created_at"`
}

type SystemSettingsRepository struct {
	pool *pgxpool.Pool
}

func NewSystemSettingsRepository(pool *pgxpool.Pool) *SystemSettingsRepository {
	return &SystemSettingsRepository{
		pool: pool,
	}
}

func (r *SystemSettingsRepository) GetSetting(ctx context.Context, key string) (*SystemSetting, error) {
	query := `
		SELECT id, key, value, description, updated_at, updated_by, created_at
		FROM system_settings
		WHERE key = $1
	`

	var setting SystemSetting
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&setting.ID,
		&setting.Key,
		&setting.Value,
		&setting.Description,
		&setting.UpdatedAt,
		&setting.UpdatedBy,
		&setting.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	return &setting, nil
}

func (r *SystemSettingsRepository) SetSetting(ctx context.Context, key, value string, updatedBy *uuid.UUID) error {
	query := `
		INSERT INTO system_settings (key, value, updated_by, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (key) 
		DO UPDATE SET 
			value = EXCLUDED.value,
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, query, key, value, updatedBy)
	if err != nil {
		return fmt.Errorf("failed to set setting: %w", err)
	}

	return nil
}

func (r *SystemSettingsRepository) GetDefaultRoutingStrategy(ctx context.Context) (string, error) {
	setting, err := r.GetSetting(ctx, "default_routing_strategy")
	if err != nil {
		return "", err
	}
	if setting == nil {
		// Default to "model" if not set
		return "model", nil
	}
	return setting.Value, nil
}

func (r *SystemSettingsRepository) SetDefaultRoutingStrategy(ctx context.Context, strategy string, updatedBy *uuid.UUID) error {
	return r.SetSetting(ctx, "default_routing_strategy", strategy, updatedBy)
}

func (r *SystemSettingsRepository) IsRoutingStrategyLocked(ctx context.Context) (bool, error) {
	setting, err := r.GetSetting(ctx, "routing_strategy_locked")
	if err != nil {
		return false, err
	}
	if setting == nil {
		return false, nil // Default to unlocked
	}
	// Trim whitespace and compare case-insensitively
	value := strings.TrimSpace(strings.ToLower(setting.Value))
	return value == "true", nil
}

func (r *SystemSettingsRepository) SetRoutingStrategyLock(ctx context.Context, locked bool, updatedBy *uuid.UUID) error {
	value := "false"
	if locked {
		value = "true"
	}
	return r.SetSetting(ctx, "routing_strategy_locked", value, updatedBy)
}

// Deprecated: Use GetDefaultRoutingStrategy instead
func (r *SystemSettingsRepository) GetRoutingStrategy(ctx context.Context) (string, error) {
	return r.GetDefaultRoutingStrategy(ctx)
}

// Deprecated: Use SetDefaultRoutingStrategy instead
func (r *SystemSettingsRepository) SetRoutingStrategy(ctx context.Context, strategy string, updatedBy *uuid.UUID) error {
	return r.SetDefaultRoutingStrategy(ctx, strategy, updatedBy)
}

