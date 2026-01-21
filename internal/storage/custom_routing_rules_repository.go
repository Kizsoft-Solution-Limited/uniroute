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
	UserID         *uuid.UUID             `db:"user_id"` // NULL = global/admin rule, non-NULL = user-specific
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

// GetActiveRules retrieves all active custom routing rules (global/admin rules), sorted by priority
func (r *CustomRoutingRulesRepository) GetActiveRules(ctx context.Context) ([]*CustomRoutingRule, error) {
	return r.GetActiveRulesForUser(ctx, nil)
}

// GetActiveRulesForUser retrieves active custom routing rules for a specific user
// If userID is nil, returns global/admin rules only
// If userID is provided, returns user-specific rules (or global rules if user has none)
func (r *CustomRoutingRulesRepository) GetActiveRulesForUser(ctx context.Context, userID *uuid.UUID) ([]*CustomRoutingRule, error) {
	var query string
	var args []interface{}
	
	if userID == nil {
		// Get global/admin rules only (user_id IS NULL)
		query = `
			SELECT id, name, condition_type, condition_value, provider_name, priority, enabled, 
			       description, user_id, created_at, updated_at, created_by, updated_by
			FROM custom_routing_rules
			WHERE enabled = true AND user_id IS NULL
			ORDER BY priority DESC, created_at ASC
		`
	} else {
		// Get user-specific rules, or global rules if user has none
		query = `
			SELECT id, name, condition_type, condition_value, provider_name, priority, enabled, 
			       description, user_id, created_at, updated_at, created_by, updated_by
			FROM custom_routing_rules
			WHERE enabled = true AND (user_id = $1 OR user_id IS NULL)
			ORDER BY user_id NULLS LAST, priority DESC, created_at ASC
		`
		args = []interface{}{userID}
	}

	var rows interface {
		Next() bool
		Scan(dest ...interface{}) error
		Close()
	}
	var err error
	
	if userID == nil {
		rows, err = r.pool.Query(ctx, query)
	} else {
		rows, err = r.pool.Query(ctx, query, args...)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get custom routing rules: %w", err)
	}
	defer rows.Close()

	var rules []*CustomRoutingRule
	for rows.Next() {
		var rule CustomRoutingRule
		var conditionValueJSON []byte
		var description *string
		var userIDPtr *uuid.UUID

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.ConditionType,
			&conditionValueJSON,
			&rule.ProviderName,
			&rule.Priority,
			&rule.Enabled,
			&description,
			&userIDPtr,
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
		rule.UserID = userIDPtr
		rules = append(rules, &rule)
	}

	return rules, nil
}

// SaveRules saves or updates custom routing rules (global/admin rules)
// rules is a slice of maps with rule data
func (r *CustomRoutingRulesRepository) SaveRules(ctx context.Context, rules []map[string]interface{}, updatedBy *uuid.UUID) error {
	return r.SaveRulesForUser(ctx, rules, nil, updatedBy)
}

// SaveRulesForUser saves or updates custom routing rules for a specific user
// If userID is nil, saves as global/admin rules
// If userID is provided, saves as user-specific rules (replaces user's existing rules)
func (r *CustomRoutingRulesRepository) SaveRulesForUser(ctx context.Context, rules []map[string]interface{}, userID *uuid.UUID, updatedBy *uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing rules for this user (or global if userID is nil)
	if userID == nil {
		_, err = tx.Exec(ctx, "DELETE FROM custom_routing_rules WHERE user_id IS NULL")
	} else {
		_, err = tx.Exec(ctx, "DELETE FROM custom_routing_rules WHERE user_id = $1", userID)
	}
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
			(name, condition_type, condition_value, provider_name, priority, enabled, description, user_id, updated_by, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		`

		_, err = tx.Exec(ctx, query,
			name,
			conditionType,
			conditionValueJSON,
			providerName,
			int(priority),
			enabled,
			description,
			userID,
			updatedBy,
		)
		if err != nil {
			return fmt.Errorf("failed to insert custom routing rule: %w", err)
		}
	}

	return tx.Commit(ctx)
}
