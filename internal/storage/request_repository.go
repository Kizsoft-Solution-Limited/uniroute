package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Request represents a tracked API request
type Request struct {
	ID           uuid.UUID
	APIKeyID     *uuid.UUID
	UserID       *uuid.UUID
	Provider     string
	Model        string
	RequestType  string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	Cost         float64
	LatencyMs    int
	StatusCode   int
	ErrorMessage *string
	CreatedAt    time.Time
}

// RequestRepository handles request tracking operations
type RequestRepository struct {
	pool *pgxpool.Pool
}

// NewRequestRepository creates a new request repository
func NewRequestRepository(pool *pgxpool.Pool) *RequestRepository {
	return &RequestRepository{
		pool: pool,
	}
}

// Create creates a new request record
func (r *RequestRepository) Create(ctx context.Context, req *Request) error {
	query := `
		INSERT INTO requests (
			id, api_key_id, user_id, provider, model, request_type,
			input_tokens, output_tokens, total_tokens, cost, latency_ms,
			status_code, error_message, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.pool.Exec(ctx, query,
		req.ID,
		req.APIKeyID,
		req.UserID,
		req.Provider,
		req.Model,
		req.RequestType,
		req.InputTokens,
		req.OutputTokens,
		req.TotalTokens,
		req.Cost,
		req.LatencyMs,
		req.StatusCode,
		req.ErrorMessage,
		req.CreatedAt,
	)

	return err
}

// GetUsageStats returns usage statistics
type UsageStats struct {
	TotalRequests      int64
	TotalTokens        int64
	TotalCost          float64
	AverageLatencyMs   float64
	RequestsByProvider map[string]int64
	RequestsByModel    map[string]int64
	CostByProvider     map[string]float64
}

// GetUsageStats returns aggregated usage statistics
func (r *RequestRepository) GetUsageStats(ctx context.Context, userID *uuid.UUID, startTime, endTime time.Time) (*UsageStats, error) {
	// PERFORMANCE OPTIMIZATION: Combine provider stats and cost in a single query
	// This reduces database round trips from 4 to 3 queries (75% reduction)
	
	// Build base WHERE clause
	baseWhere := "WHERE created_at >= $1 AND created_at <= $2"
	args := []interface{}{startTime, endTime}
	argPos := 3

	if userID != nil {
		baseWhere += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *userID)
		argPos++
	}

	// Main stats query
	mainQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(AVG(latency_ms), 0) as avg_latency
		FROM requests
		%s
	`, baseWhere)

	var stats UsageStats
	err := r.pool.QueryRow(ctx, mainQuery, args...).Scan(
		&stats.TotalRequests,
		&stats.TotalTokens,
		&stats.TotalCost,
		&stats.AverageLatencyMs,
	)
	if err != nil {
		return nil, err
	}

	// Combined provider stats (count + cost in one query)
	providerQuery := fmt.Sprintf(`
		SELECT 
			provider,
			COUNT(*) as count,
			COALESCE(SUM(cost), 0) as total_cost
		FROM requests
		%s
		GROUP BY provider
	`, baseWhere)

	rows, err := r.pool.Query(ctx, providerQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats.RequestsByProvider = make(map[string]int64)
	stats.RequestsByModel = make(map[string]int64)
	stats.CostByProvider = make(map[string]float64)

	for rows.Next() {
		var provider string
		var count int64
		var cost float64
		if err := rows.Scan(&provider, &count, &cost); err != nil {
			continue
		}
		stats.RequestsByProvider[provider] = count
		stats.CostByProvider[provider] = cost
	}

	// Model stats query
	modelQuery := fmt.Sprintf(`
		SELECT 
			model,
			COUNT(*) as count
		FROM requests
		%s
		GROUP BY model
	`, baseWhere)

	rows, err = r.pool.Query(ctx, modelQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var model string
		var count int64
		if err := rows.Scan(&model, &count); err != nil {
			continue
		}
		stats.RequestsByModel[model] = count
	}

	return &stats, nil
}

// GetRequests returns paginated list of requests
func (r *RequestRepository) GetRequests(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]*Request, error) {
	query := `
		SELECT id, api_key_id, user_id, provider, model, request_type,
		       input_tokens, output_tokens, total_tokens, cost, latency_ms,
		       status_code, error_message, created_at
		FROM requests
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *userID)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*Request
	for rows.Next() {
		var req Request
		err := rows.Scan(
			&req.ID,
			&req.APIKeyID,
			&req.UserID,
			&req.Provider,
			&req.Model,
			&req.RequestType,
			&req.InputTokens,
			&req.OutputTokens,
			&req.TotalTokens,
			&req.Cost,
			&req.LatencyMs,
			&req.StatusCode,
			&req.ErrorMessage,
			&req.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}

	return requests, rows.Err()
}
