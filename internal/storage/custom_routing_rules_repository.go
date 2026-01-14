package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CustomRoutingRule represents a custom routing rule
type CustomRoutingRule struct {
	ID             int                    `db:"id"`
	Name           string                 `db:"name"`
	ConditionType  string                 `db:"condition_type"`
	ConditionValue map[string]interface{} `db:"condition_value"`
	ProviderName   string                 `db:"provider_name"`
	Priority       int                    `db:"priority"`
	Enabled        bool                   `db:"enabled"`
	Description    *string                `db:"description"`
	CreatedAt      time.Time              `db:"created_at"`
	UpdatedAt      time.Time              `db:"updated_at"`
	CreatedBy      *uuid.UUID             `db:"created_by"`
	UpdatedBy      *uuid.UUID             `db:"updated_by"`
}

// CustomRoutingRulesRepository handles custom routing rules database operations
type CustomRoutingRulesRepository struct {
	pool *pgxpool.Pool
}

// NewCustomRoutingRulesRepository creates a new custom routing rules repository
func NewCustomRoutingRulesRepository(pool *pgxpool.Pool) *CustomRoutingRulesRepository {
	return &CustomRoutingRulesRepository{
		pool: pool,
	}
}

// GetActiveRules retrieves all active custom routing rules, sorted by priority
func (r *CustomRoutingRulesRepository) GetActiveRules(ctx context.Context) ([]*CustomRoutingRule, error) {
	query := `
		SELECT id, name, condition_type, condition_value, provider_name, priority, enabled, 
		       description, created_at, updated_at, created_by, updated_by
		FROM custom_routing_rules
		WHERE enabled = true
		ORDER BY priority DESC, created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom routing rules: %w", err)
	}
	defer rows.Close()

	var rules []*CustomRoutingRule
	for rows.Next() {
		var rule CustomRoutingRule
		var conditionValueJSON []byte
		var description *string

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.ConditionType,
			&conditionValueJSON,
			&rule.ProviderName,
			&rule.Priority,
			&rule.Enabled,
			&description,
			&rule.CreatedAt,
			&rule.UpdatedAt,
			&rule.CreatedBy,
			&rule.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan custom routing rule: %w", err)
		}

		// Parse JSON condition value
		if len(conditionValueJSON) > 0 {
			if err := json.Unmarshal(conditionValueJSON, &rule.ConditionValue); err != nil {
				return nil, fmt.Errorf("failed to parse condition_value JSON: %w", err)
			}
		}

		rule.Description = description
		rules = append(rules, &rule)
	}

	return rules, nil
}

// SaveRules saves or updates custom routing rules
// rules is a slice of maps with rule data
func (r *CustomRoutingRulesRepository) SaveRules(ctx context.Context, rules []map[string]interface{}, updatedBy *uuid.UUID) error {
	// For simplicity, we'll delete all existing rules and insert new ones
	// In production, you might want to do upsert logic

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete all existing rules
	_, err = tx.Exec(ctx, "DELETE FROM custom_routing_rules")
	if err != nil {
		return fmt.Errorf("failed to delete existing rules: %w", err)
	}

	// Insert new rules
	for _, ruleMap := range rules {
		name, _ := ruleMap["name"].(string)
		conditionType, _ := ruleMap["condition_type"].(string)
		conditionValue := ruleMap["condition_value"]
		providerName, _ := ruleMap["provider_name"].(string)
		priority, _ := ruleMap["priority"].(float64)
		enabled, _ := ruleMap["enabled"].(bool)
		description, _ := ruleMap["description"].(string)

		// Convert condition value to JSON
		conditionValueJSON, err := json.Marshal(conditionValue)
		if err != nil {
			return fmt.Errorf("failed to marshal condition_value: %w", err)
		}

		query := `
			INSERT INTO custom_routing_rules 
			(name, condition_type, condition_value, provider_name, priority, enabled, description, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		`

		_, err = tx.Exec(ctx, query,
			name,
			conditionType,
			conditionValueJSON,
			providerName,
			int(priority),
			enabled,
			description,
			updatedBy,
		)
		if err != nil {
			return fmt.Errorf("failed to insert custom routing rule: %w", err)
		}
	}

	return tx.Commit(ctx)
}
