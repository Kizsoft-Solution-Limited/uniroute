package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrorLogRepository handles error log database operations
type ErrorLogRepository struct {
	pool *pgxpool.Pool
}

// NewErrorLogRepository creates a new error log repository
func NewErrorLogRepository(pool *pgxpool.Pool) *ErrorLogRepository {
	return &ErrorLogRepository{pool: pool}
}

// CreateErrorLog saves a new error log to the database
func (r *ErrorLogRepository) CreateErrorLog(ctx context.Context, log *ErrorLog) error {
	// Convert context map to JSONB
	var contextJSON []byte
	var err error
	if log.Context != nil {
		contextJSON, err = json.Marshal(log.Context)
		if err != nil {
			return fmt.Errorf("failed to marshal context: %w", err)
		}
	}

	query := `
		INSERT INTO error_logs (
			user_id, error_type, message, stack_trace, url, user_agent, 
			ip_address, context, severity, resolved
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, created_at
	`

	err = r.pool.QueryRow(
		ctx,
		query,
		log.UserID,
		log.ErrorType,
		log.Message,
		log.StackTrace,
		log.URL,
		log.UserAgent,
		log.IPAddress,
		contextJSON,
		log.Severity,
		log.Resolved,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create error log: %w", err)
	}

	return nil
}

// GetErrorLogs retrieves error logs with optional filters
func (r *ErrorLogRepository) GetErrorLogs(ctx context.Context, filters ErrorLogFilters) ([]ErrorLog, error) {
	query := `
		SELECT id, user_id, error_type, message, stack_trace, url, user_agent,
		       ip_address, context, severity, resolved, created_at
		FROM error_logs
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *filters.UserID)
		argPos++
	}

	if filters.ErrorType != "" {
		query += fmt.Sprintf(" AND error_type = $%d", argPos)
		args = append(args, filters.ErrorType)
		argPos++
	}

	if filters.Severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argPos)
		args = append(args, filters.Severity)
		argPos++
	}

	if filters.Resolved != nil {
		query += fmt.Sprintf(" AND resolved = $%d", argPos)
		args = append(args, *filters.Resolved)
		argPos++
	}

	if filters.Limit > 0 {
		query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argPos)
		args = append(args, filters.Limit)
	} else {
		query += " ORDER BY created_at DESC LIMIT 100"
	}

	var logs []ErrorLog
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query error logs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var log ErrorLog
		var contextJSON []byte
		var userID sql.NullString
		var stackTrace sql.NullString
		var url sql.NullString
		var userAgent sql.NullString
		var ipAddress sql.NullString

		err := rows.Scan(
			&log.ID,
			&userID,
			&log.ErrorType,
			&log.Message,
			&stackTrace,
			&url,
			&userAgent,
			&ipAddress,
			&contextJSON,
			&log.Severity,
			&log.Resolved,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan error log: %w", err)
		}

		// Parse nullable fields
		if userID.Valid {
			uid, err := uuid.Parse(userID.String)
			if err == nil {
				log.UserID = &uid
			}
		}
		if stackTrace.Valid {
			log.StackTrace = &stackTrace.String
		}
		if url.Valid {
			log.URL = &url.String
		}
		if userAgent.Valid {
			log.UserAgent = &userAgent.String
		}
		if ipAddress.Valid {
			log.IPAddress = &ipAddress.String
		}

		// Parse context JSON
		if len(contextJSON) > 0 {
			if err := json.Unmarshal(contextJSON, &log.Context); err != nil {
				log.Context = make(map[string]interface{})
			}
		} else {
			log.Context = make(map[string]interface{})
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// MarkResolved marks an error log as resolved
func (r *ErrorLogRepository) MarkResolved(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE error_logs SET resolved = true WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark error log as resolved: %w", err)
	}
	return nil
}

// ErrorLogFilters represents filters for querying error logs
type ErrorLogFilters struct {
	UserID    *uuid.UUID
	ErrorType string
	Severity  string
	Resolved  *bool
	Limit     int
}

