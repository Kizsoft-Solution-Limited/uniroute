package tunnel

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// TunnelRepository handles tunnel database operations
type TunnelRepository struct {
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

// NewTunnelRepository creates a new tunnel repository
func NewTunnelRepository(pool *pgxpool.Pool, logger zerolog.Logger) *TunnelRepository {
	return &TunnelRepository{
		pool:   pool,
		logger: logger,
	}
}

// CreateTunnel creates a new tunnel in the database
func (r *TunnelRepository) CreateTunnel(ctx context.Context, tunnel *Tunnel) error {
	query := `
		INSERT INTO tunnels (
			id, user_id, subdomain, custom_domain, local_url, public_url,
			status, region, created_at, updated_at, last_active_at, request_count, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	var userID *uuid.UUID
	if tunnel.UserID != "" {
		parsed, err := uuid.Parse(tunnel.UserID)
		if err == nil {
			userID = &parsed
		}
	}

	var tunnelID uuid.UUID
	if tunnel.ID != "" {
		parsed, err := uuid.Parse(tunnel.ID)
		if err == nil {
			tunnelID = parsed
		} else {
			tunnelID = uuid.New()
		}
	} else {
		tunnelID = uuid.New()
	}

	now := time.Now()
	_, err := r.pool.Exec(ctx, query,
		tunnelID,
		userID,
		tunnel.Subdomain,
		tunnel.CustomDomain,
		tunnel.LocalURL,
		tunnel.PublicURL,
		tunnel.Status,
		tunnel.Region,
		now,
		now,
		now,
		tunnel.RequestCount,
		nil, // metadata JSONB
	)

	if err != nil {
		return err
	}

	tunnel.ID = tunnelID.String()
	return nil
}

// GetTunnelBySubdomain retrieves a tunnel by subdomain
func (r *TunnelRepository) GetTunnelBySubdomain(ctx context.Context, subdomain string) (*Tunnel, error) {
	query := `
		SELECT id, user_id, subdomain, custom_domain, local_url, public_url,
		       status, region, created_at, updated_at, last_active_at, request_count
		FROM tunnels
		WHERE subdomain = $1 AND status = 'active'
	`

	var tunnel Tunnel
	var userID sql.NullString
	var customDomain sql.NullString
	var lastActive sql.NullTime

	err := r.pool.QueryRow(ctx, query, subdomain).Scan(
		&tunnel.ID,
		&userID,
		&tunnel.Subdomain,
		&customDomain,
		&tunnel.LocalURL,
		&tunnel.PublicURL,
		&tunnel.Status,
		&tunnel.Region,
		&tunnel.CreatedAt,
		&tunnel.UpdatedAt,
		&lastActive,
		&tunnel.RequestCount,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTunnelNotFound
	}
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		tunnel.UserID = userID.String
	}
	if customDomain.Valid {
		tunnel.CustomDomain = customDomain.String
	}
	if lastActive.Valid {
		tunnel.LastActive = lastActive.Time
	}

	return &tunnel, nil
}

// GetTunnelByID retrieves a tunnel by ID
func (r *TunnelRepository) GetTunnelByID(ctx context.Context, tunnelID uuid.UUID) (*Tunnel, error) {
	query := `
		SELECT id, user_id, subdomain, custom_domain, local_url, public_url,
		       status, region, created_at, updated_at, last_active_at, request_count
		FROM tunnels
		WHERE id = $1
	`

	var tunnel Tunnel
	var userID sql.NullString
	var customDomain sql.NullString
	var lastActive sql.NullTime

	err := r.pool.QueryRow(ctx, query, tunnelID).Scan(
		&tunnel.ID,
		&userID,
		&tunnel.Subdomain,
		&customDomain,
		&tunnel.LocalURL,
		&tunnel.PublicURL,
		&tunnel.Status,
		&tunnel.Region,
		&tunnel.CreatedAt,
		&tunnel.UpdatedAt,
		&lastActive,
		&tunnel.RequestCount,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTunnelNotFound
	}
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		tunnel.UserID = userID.String
	}
	if customDomain.Valid {
		tunnel.CustomDomain = customDomain.String
	}
	if lastActive.Valid {
		tunnel.LastActive = lastActive.Time
	}

	return &tunnel, nil
}

// ListTunnelsByUser retrieves all tunnels for a user
func (r *TunnelRepository) ListTunnelsByUser(ctx context.Context, userID uuid.UUID) ([]*Tunnel, error) {
	query := `
		SELECT id, user_id, subdomain, custom_domain, local_url, public_url,
		       status, region, created_at, updated_at, last_active_at, request_count
		FROM tunnels
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tunnels []*Tunnel
	for rows.Next() {
		var tunnel Tunnel
		var userID sql.NullString
		var customDomain sql.NullString
		var lastActive sql.NullTime

		err := rows.Scan(
			&tunnel.ID,
			&userID,
			&tunnel.Subdomain,
			&customDomain,
			&tunnel.LocalURL,
			&tunnel.PublicURL,
			&tunnel.Status,
			&tunnel.Region,
			&tunnel.CreatedAt,
			&tunnel.UpdatedAt,
			&lastActive,
			&tunnel.RequestCount,
		)
		if err != nil {
			return nil, err
		}

		if userID.Valid {
			tunnel.UserID = userID.String
		}
		if customDomain.Valid {
			tunnel.CustomDomain = customDomain.String
		}
		if lastActive.Valid {
			tunnel.LastActive = lastActive.Time
		}

		tunnels = append(tunnels, &tunnel)
	}

	return tunnels, rows.Err()
}

// UpdateTunnelStatus updates tunnel status
func (r *TunnelRepository) UpdateTunnelStatus(ctx context.Context, tunnelID, status string) error {
	query := `
		UPDATE tunnels
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, status, tunnelID)
	return err
}

// UpdateTunnelActivity updates tunnel last active time and request count
func (r *TunnelRepository) UpdateTunnelActivity(ctx context.Context, tunnelID string, requestCount int64) error {
	query := `
		UPDATE tunnels
		SET last_active_at = NOW(), request_count = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, requestCount, tunnelID)
	return err
}

// CreateSession creates a new tunnel session
func (r *TunnelRepository) CreateSession(ctx context.Context, session *TunnelSession) error {
	query := `
		INSERT INTO tunnel_sessions (
			id, tunnel_id, client_id, server_id, connected_at, status
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	sessionID := uuid.New()
	if session.ID != "" {
		parsed, err := uuid.Parse(session.ID)
		if err == nil {
			sessionID = parsed
		}
	}

	tunnelID, err := uuid.Parse(session.TunnelID)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query,
		sessionID,
		tunnelID,
		session.ClientID,
		session.ServerID,
		time.Now(),
		session.Status,
	)

	if err != nil {
		return err
	}

	session.ID = sessionID.String()
	return nil
}

// UpdateSessionStatus updates session status
func (r *TunnelRepository) UpdateSessionStatus(ctx context.Context, sessionID, status string) error {
	query := `
		UPDATE tunnel_sessions
		SET status = $1, disconnected_at = CASE WHEN $1 = 'disconnected' THEN NOW() ELSE disconnected_at END
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, status, sessionID)
	return err
}

// GetTokenInfo retrieves token information
func (r *TunnelRepository) GetTokenInfo(ctx context.Context, tokenHash string) (*TokenInfo, error) {
	query := `
		SELECT token_hash, name, expires_at, created_at, last_used_at, is_active
		FROM tunnel_tokens
		WHERE token_hash = $1 AND is_active = true
	`

	var info TokenInfo
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime

	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&info.TokenHash,
		&info.Name,
		&expiresAt,
		&info.CreatedAt,
		&lastUsedAt,
		&info.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		info.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		info.LastUsedAt = &lastUsedAt.Time
	}

	return &info, nil
}

// UpdateTokenLastUsed updates token last used timestamp
func (r *TunnelRepository) UpdateTokenLastUsed(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE tunnel_tokens
		SET last_used_at = NOW()
		WHERE token_hash = $1
	`

	_, err := r.pool.Exec(ctx, query, tokenHash)
	return err
}

// TunnelSession represents a tunnel session
type TunnelSession struct {
	ID          string
	TunnelID    string
	ClientID    string
	ServerID    string
	ConnectedAt time.Time
	Status      string
}

// CreateTunnelRequest creates a tunnel request log entry
func (r *TunnelRepository) CreateTunnelRequest(ctx context.Context, req *TunnelRequestLog) error {
	query := `
		INSERT INTO tunnel_requests (
			id, tunnel_id, request_id, method, path, query_string,
			request_headers, request_body, status_code, response_headers,
			response_body, latency_ms, request_size, response_size,
			remote_addr, user_agent, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	requestID := uuid.New()
	if req.ID != (uuid.UUID{}) {
		requestID = req.ID
	}

	tunnelUUID, err := uuid.Parse(req.TunnelID)
	if err != nil {
		return err
	}

	// Convert headers to JSONB
	var requestHeadersJSON []byte
	if req.RequestHeaders != nil {
		requestHeadersJSON, _ = json.Marshal(req.RequestHeaders)
	}
	var responseHeadersJSON []byte
	if req.ResponseHeaders != nil {
		responseHeadersJSON, _ = json.Marshal(req.ResponseHeaders)
	}

	_, err = r.pool.Exec(ctx, query,
		requestID,
		tunnelUUID,
		req.RequestID,
		req.Method,
		req.Path,
		req.QueryString,
		requestHeadersJSON,
		req.RequestBody,
		req.StatusCode,
		responseHeadersJSON,
		req.ResponseBody,
		req.LatencyMs,
		req.RequestSize,
		req.ResponseSize,
		req.RemoteAddr,
		req.UserAgent,
		req.CreatedAt,
	)

	return err
}

// GetTunnelRequest retrieves a tunnel request by ID
func (r *TunnelRepository) GetTunnelRequest(ctx context.Context, requestID string) (*TunnelRequestLog, error) {
	query := `
		SELECT id, tunnel_id, request_id, method, path, query_string,
		       request_headers, request_body, status_code, response_headers,
		       response_body, latency_ms, request_size, response_size,
		       remote_addr, user_agent, created_at
		FROM tunnel_requests
		WHERE request_id = $1
		LIMIT 1
	`

	var req TunnelRequestLog
	var tunnelUUID uuid.UUID
	var requestHeadersJSON, responseHeadersJSON []byte

	err := r.pool.QueryRow(ctx, query, requestID).Scan(
		&req.ID,
		&tunnelUUID,
		&req.RequestID,
		&req.Method,
		&req.Path,
		&req.QueryString,
		&requestHeadersJSON,
		&req.RequestBody,
		&req.StatusCode,
		&responseHeadersJSON,
		&req.ResponseBody,
		&req.LatencyMs,
		&req.RequestSize,
		&req.ResponseSize,
		&req.RemoteAddr,
		&req.UserAgent,
		&req.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	req.TunnelID = tunnelUUID.String()

	// Parse headers JSON
	if len(requestHeadersJSON) > 0 {
		json.Unmarshal(requestHeadersJSON, &req.RequestHeaders)
	}
	if len(responseHeadersJSON) > 0 {
		json.Unmarshal(responseHeadersJSON, &req.ResponseHeaders)
	}

	return &req, nil
}

// ListTunnelRequests retrieves tunnel requests with filtering
func (r *TunnelRepository) ListTunnelRequests(ctx context.Context, tunnelID string, limit, offset int, method, pathFilter string) ([]*TunnelRequestLog, error) {
	query := `
		SELECT id, tunnel_id, request_id, method, path, query_string,
		       request_headers, request_body, status_code, response_headers,
		       response_body, latency_ms, request_size, response_size,
		       remote_addr, user_agent, created_at
		FROM tunnel_requests
		WHERE tunnel_id = $1
	`

	args := []interface{}{tunnelID}
	argIndex := 2

	if method != "" {
		query += fmt.Sprintf(" AND method = $%d", argIndex)
		args = append(args, method)
		argIndex++
	}

	if pathFilter != "" {
		query += fmt.Sprintf(" AND path LIKE $%d", argIndex)
		args = append(args, "%"+pathFilter+"%")
		argIndex++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*TunnelRequestLog
	for rows.Next() {
		var req TunnelRequestLog
		var tunnelUUID uuid.UUID
		var requestHeadersJSON, responseHeadersJSON []byte

		err := rows.Scan(
			&req.ID,
			&tunnelUUID,
			&req.RequestID,
			&req.Method,
			&req.Path,
			&req.QueryString,
			&requestHeadersJSON,
			&req.RequestBody,
			&req.StatusCode,
			&responseHeadersJSON,
			&req.ResponseBody,
			&req.LatencyMs,
			&req.RequestSize,
			&req.ResponseSize,
			&req.RemoteAddr,
			&req.UserAgent,
			&req.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		req.TunnelID = tunnelUUID.String()

		// Parse headers JSON
		if len(requestHeadersJSON) > 0 {
			json.Unmarshal(requestHeadersJSON, &req.RequestHeaders)
		}
		if len(responseHeadersJSON) > 0 {
			json.Unmarshal(responseHeadersJSON, &req.ResponseHeaders)
		}

		requests = append(requests, &req)
	}

	return requests, rows.Err()
}

// TunnelRequestLog represents a logged tunnel request
type TunnelRequestLog struct {
	ID              uuid.UUID
	TunnelID        string
	RequestID       string
	Method          string
	Path            string
	QueryString     string
	RequestHeaders  map[string]string
	RequestBody     []byte
	StatusCode      int
	ResponseHeaders map[string]string
	ResponseBody    []byte
	LatencyMs       int
	RequestSize     int
	ResponseSize    int
	RemoteAddr      string
	UserAgent       string
	CreatedAt       time.Time
}
