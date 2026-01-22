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
// If a tunnel with the same subdomain already exists, it updates the LocalURL, PublicURL, and status
func (r *TunnelRepository) CreateTunnel(ctx context.Context, tunnel *Tunnel) error {
	// Use UPSERT (INSERT ... ON CONFLICT UPDATE) to handle case where subdomain already exists
	// This ensures LocalURL is always updated when creating a tunnel with an existing subdomain
	// When conflict occurs, we update the existing record and return its ID (not the new one we tried to insert)
	query := `
		INSERT INTO tunnels (
			id, user_id, subdomain, custom_domain, local_url, public_url,
			status, region, created_at, updated_at, last_active_at, request_count, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (subdomain) DO UPDATE SET
			local_url = EXCLUDED.local_url,
			public_url = EXCLUDED.public_url,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at,
			last_active_at = EXCLUDED.last_active_at,
			user_id = COALESCE(EXCLUDED.user_id, tunnels.user_id),
			request_count = CASE WHEN EXCLUDED.request_count = 0 THEN tunnels.request_count ELSE EXCLUDED.request_count END
		RETURNING id
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
			// Tunnel ID is not a valid UUID - log warning and generate new one
			// This can happen if old hex-format IDs are still being used
			r.logger.Warn().
				Err(err).
				Str("invalid_tunnel_id", tunnel.ID).
				Str("tunnel_id_length", fmt.Sprintf("%d", len(tunnel.ID))).
				Msg("Tunnel ID is not a valid UUID format - generating new UUID. This may cause tunnel ID mismatch.")
			tunnelID = uuid.New()
			// Update the tunnel.ID to the new UUID so it matches what's saved
			tunnel.ID = tunnelID.String()
		}
	} else {
		tunnelID = uuid.New()
		tunnel.ID = tunnelID.String()
	}

	now := time.Now()
	var returnedID uuid.UUID
	err := r.pool.QueryRow(ctx, query,
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
	).Scan(&returnedID)

	if err != nil {
		r.logger.Error().
			Err(err).
			Str("tunnel_id", tunnelID.String()).
			Str("subdomain", tunnel.Subdomain).
			Str("local_url", tunnel.LocalURL).
			Msg("Failed to create/update tunnel in database")
		return err
	}

	// Use the returned ID (which might be the existing tunnel's ID if there was a conflict)
	// If returnedID != tunnelID, it means we updated an existing tunnel
	if returnedID != tunnelID {
		r.logger.Debug().
			Str("subdomain", tunnel.Subdomain).
			Str("local_url", tunnel.LocalURL).
			Msg("Updated existing tunnel LocalURL")
	}
	
	tunnel.ID = returnedID.String()
	return nil
}

// GetTunnelByCustomDomain retrieves a tunnel by custom domain
func (r *TunnelRepository) GetTunnelByCustomDomain(ctx context.Context, domain string) (*Tunnel, error) {
	query := `
		SELECT id, user_id, subdomain, custom_domain, local_url, public_url,
		       status, region, created_at, updated_at, last_active_at, request_count
		FROM tunnels
		WHERE custom_domain = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var tunnel Tunnel
	var userID sql.NullString
	var customDomain sql.NullString
	var lastActive sql.NullTime

	err := r.pool.QueryRow(ctx, query, domain).Scan(
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

// GetTunnelBySubdomain retrieves a tunnel by subdomain (regardless of status)
// This allows resuming inactive tunnels
func (r *TunnelRepository) GetTunnelBySubdomain(ctx context.Context, subdomain string) (*Tunnel, error) {
	query := `
		SELECT id, user_id, subdomain, custom_domain, local_url, public_url,
		       status, region, created_at, updated_at, last_active_at, request_count
		FROM tunnels
		WHERE subdomain = $1
		ORDER BY created_at DESC
		LIMIT 1
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

// ListAllTunnels retrieves all tunnels (admin view)
func (r *TunnelRepository) ListAllTunnels(ctx context.Context) ([]*Tunnel, error) {
	query := `
		SELECT id, user_id, subdomain, custom_domain, local_url, public_url,
		       status, region, created_at, updated_at, last_active_at, request_count
		FROM tunnels
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tunnels, nil
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
// If requestCount is 0, it only updates last_active_at without changing request_count
// If requestCount > 0, it increments the request_count by 1
func (r *TunnelRepository) UpdateTunnelActivity(ctx context.Context, tunnelID string, requestCount int64) error {
	var query string
	if requestCount == 0 {
		// Only update last_active_at, don't touch request_count
		query = `
			UPDATE tunnels
			SET last_active_at = NOW(), updated_at = NOW()
			WHERE id = $1
		`
		_, err := r.pool.Exec(ctx, query, tunnelID)
		return err
	} else {
		// Increment request_count by 1 (to avoid double counting)
		query = `
			UPDATE tunnels
			SET last_active_at = NOW(), request_count = request_count + 1, updated_at = NOW()
			WHERE id = $1
		`
		_, err := r.pool.Exec(ctx, query, tunnelID)
		return err
	}
}

// UpdateTunnelLocalURL updates tunnel LocalURL
func (r *TunnelRepository) UpdateTunnelLocalURL(ctx context.Context, tunnelID, localURL string) error {
	query := `
		UPDATE tunnels
		SET local_url = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, localURL, tunnelID)
	return err
}

// UpdateTunnelCustomDomain updates tunnel custom domain
func (r *TunnelRepository) UpdateTunnelCustomDomain(ctx context.Context, tunnelID uuid.UUID, customDomain string) error {
	query := `
		UPDATE tunnels
		SET custom_domain = $1, updated_at = NOW()
		WHERE id = $2
	`

	var domain *string
	if customDomain != "" {
		domain = &customDomain
	}

	_, err := r.pool.Exec(ctx, query, domain, tunnelID)
	return err
}

// AssociateTunnelWithUser associates a tunnel with a user (for CLI-created tunnels)
func (r *TunnelRepository) AssociateTunnelWithUser(ctx context.Context, tunnelID uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE tunnels
		SET user_id = $1, updated_at = NOW()
		WHERE id = $2 AND (user_id IS NULL OR user_id = $1)
	`

	result, err := r.pool.Exec(ctx, query, userID, tunnelID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrTunnelNotFound
	}

	return nil
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

// TunnelStatsPoint represents a single data point in tunnel statistics
type TunnelStatsPoint struct {
	Time           time.Time `json:"time"`
	ActiveTunnels  int       `json:"active_tunnels"`
	TotalTunnels   int       `json:"total_tunnels"`
}

// GetTunnelStatsOverTime returns tunnel statistics grouped by time intervals
// For 6h/24h: Returns ONE aggregated point with total tunnels created in that period
// For 7d: Returns one point per day with total tunnels created on that day
// If userID is uuid.Nil, returns stats for ALL tunnels (admin view)
// intervalHours can be fractional (e.g., 0.25 for 15 minutes)
func (r *TunnelRepository) GetTunnelStatsOverTime(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, intervalHours float64) ([]TunnelStatsPoint, error) {
	// Get tunnels - all tunnels if userID is Nil (admin), otherwise user's tunnels
	var tunnels []*Tunnel
	var err error
	
	if userID == uuid.Nil {
		// Admin view: get all tunnels
		r.logger.Debug().Msg("GetTunnelStatsOverTime: fetching all tunnels (admin view)")
		tunnels, err = r.ListAllTunnels(ctx)
	} else {
		// User view: get user's tunnels
		r.logger.Debug().Str("user_id", userID.String()).Msg("GetTunnelStatsOverTime: fetching user tunnels")
		tunnels, err = r.ListTunnelsByUser(ctx, userID)
	}
	
	if err != nil {
		return nil, err
	}

	r.logger.Debug().Int("tunnel_count", len(tunnels)).Msg("GetTunnelStatsOverTime: tunnels fetched")

	// If no tunnels, return empty array (not nil to avoid frontend issues)
	if len(tunnels) == 0 {
		r.logger.Debug().Msg("GetTunnelStatsOverTime: no tunnels found, returning empty array")
		return []TunnelStatsPoint{}, nil
	}
	
	// For daily intervals (7d view): aggregate by day
	// For 6h/24h views: return time-series data points to show trend/slope
	if intervalHours >= 24.0 {
		// 7d view: aggregate by day - count tunnels created on each day
		r.logger.Debug().
			Float64("interval_hours", intervalHours).
			Msg("GetTunnelStatsOverTime: using daily aggregation (7d view)")
		return r.aggregateByDay(ctx, tunnels, startTime, endTime)
	} else {
		// 6h/24h view: return time-series data points to show trend over time
		r.logger.Debug().
			Float64("interval_hours", intervalHours).
			Time("start_time", startTime).
			Time("end_time", endTime).
			Msg("GetTunnelStatsOverTime: using time-series aggregation (multiple points for trend)")
		return r.generateTimeSeries(ctx, tunnels, startTime, endTime, intervalHours)
	}
}

// aggregateByPeriod returns a single data point with current tunnel counts
// For 6h/24h views, this shows the CURRENT state (all active tunnels now, not just created in period)
// CRITICAL: This function MUST return exactly ONE point, never multiple points
func (r *TunnelRepository) aggregateByPeriod(ctx context.Context, tunnels []*Tunnel, startTime, endTime time.Time) ([]TunnelStatsPoint, error) {
	activeCount := 0
	totalCount := 0
	
	// Count ALL tunnels (current state), not just those created in the period
	// This represents the current state: how many tunnels exist and are active RIGHT NOW
	// This includes tunnels created 3 hours ago, 10 hours ago, etc. - all currently existing tunnels
	for _, tunnel := range tunnels {
		totalCount++
		r.logger.Debug().
			Str("tunnel_id", tunnel.ID).
			Str("status", tunnel.Status).
			Time("created_at", tunnel.CreatedAt).
			Msg("aggregateByPeriod: processing tunnel")
		if tunnel.Status == "active" {
			activeCount++
			r.logger.Debug().
				Str("tunnel_id", tunnel.ID).
				Msg("aggregateByPeriod: tunnel is active")
		}
	}
	
	// Return single aggregated point at the middle of the period
	midTime := startTime.Add(endTime.Sub(startTime) / 2)
	
	r.logger.Debug().
		Time("start_time", startTime).
		Time("end_time", endTime).
		Int("total_tunnels", totalCount).
		Int("active_tunnels", activeCount).
		Int("all_tunnels_checked", len(tunnels)).
		Msg("GetTunnelStatsOverTime: aggregated by period - returning SINGLE point with current state")
	
	// CRITICAL: Return exactly ONE aggregated point - never multiple points
	// This is the key difference from time-series data
	result := []TunnelStatsPoint{
		{
			Time:          midTime,
			ActiveTunnels: activeCount,
			TotalTunnels:  totalCount,
		},
	}
	
	// Double-check we're returning exactly one point
	if len(result) != 1 {
		r.logger.Error().
			Int("result_points", len(result)).
			Msg("GetTunnelStatsOverTime: ERROR - aggregateByPeriod returned multiple points, this should never happen!")
		// Force single point
		result = result[:1]
	}
	
	r.logger.Debug().
		Int("result_points", len(result)).
		Int("active_tunnels", result[0].ActiveTunnels).
		Int("total_tunnels", result[0].TotalTunnels).
		Msg("GetTunnelStatsOverTime: aggregateByPeriod result - SINGLE POINT")
	
	return result, nil
}

// generateTimeSeries returns multiple data points over time to show trend/slope
// For 6h/24h views, this creates time-series data showing cumulative tunnel counts at each interval
// Each point shows the TOTAL count of tunnels that existed at that time (cumulative)
func (r *TunnelRepository) generateTimeSeries(ctx context.Context, tunnels []*Tunnel, startTime, endTime time.Time, intervalHours float64) ([]TunnelStatsPoint, error) {
	var stats []TunnelStatsPoint
	now := time.Now()
	interval := time.Duration(intervalHours * float64(time.Hour))
	currentTime := startTime
	
	r.logger.Debug().
		Time("start_time", startTime).
		Time("end_time", endTime).
		Float64("interval_hours", intervalHours).
		Dur("interval", interval).
		Int("total_tunnels_available", len(tunnels)).
		Msg("generateTimeSeries: generating time-series data points")
	
	// Log all tunnels for debugging
	for _, tunnel := range tunnels {
		r.logger.Debug().
			Str("tunnel_id", tunnel.ID).
			Str("status", tunnel.Status).
			Time("created_at", tunnel.CreatedAt).
			Msg("generateTimeSeries: available tunnel")
	}
	
	// Generate data points for each time interval
	// Stop before endTime to avoid going past it, we'll add a "now" point at the end
	for currentTime.Before(endTime) {
		activeCount := 0
		totalCount := 0
		
		// Count tunnels that existed at this point in time (cumulative)
		// A tunnel exists at time T if it was created before or at T
		// This creates a cumulative trend showing tunnel growth over time
		for _, tunnel := range tunnels {
			// Tunnel was created before or at this time point
			// Use Before with a small buffer to handle timezone/rounding issues
			if tunnel.CreatedAt.Before(currentTime.Add(time.Second)) || tunnel.CreatedAt.Equal(currentTime) {
				totalCount++
				// For active count, use current status
				// This shows how many of the existing tunnels are currently active
				if tunnel.Status == "active" {
					activeCount++
				}
			}
		}
		
		stats = append(stats, TunnelStatsPoint{
			Time:          currentTime,
			ActiveTunnels: activeCount,
			TotalTunnels:  totalCount,
		})
		
		r.logger.Debug().
			Time("point_time", currentTime).
			Int("active", activeCount).
			Int("total", totalCount).
			Msg("generateTimeSeries: added data point")
		
		// Move to next interval
		currentTime = currentTime.Add(interval)
		
		// Prevent infinite loop - max 50 points
		if len(stats) >= 50 {
			break
		}
	}
	
	// Always add a "now" point with current state for real-time accuracy
	// Count ALL tunnels that exist NOW (all tunnels in the database exist now)
	currentActiveCount := 0
	currentTotalCount := len(tunnels) // All tunnels exist now
	for _, tunnel := range tunnels {
		if tunnel.Status == "active" {
			currentActiveCount++
		}
	}
	
	// Add or update the "now" point
	if len(stats) > 0 {
		lastPoint := &stats[len(stats)-1]
		timeDiff := now.Sub(lastPoint.Time)
		// If last point is very close to now (within half interval), update it
		if timeDiff < interval/2 {
			lastPoint.ActiveTunnels = currentActiveCount
			lastPoint.TotalTunnels = currentTotalCount
			lastPoint.Time = now
		} else {
			// Add new point for current time
			stats = append(stats, TunnelStatsPoint{
				Time:          now,
				ActiveTunnels: currentActiveCount,
				TotalTunnels:  currentTotalCount,
			})
		}
	} else {
		// No points generated, add at least the "now" point"
		stats = append(stats, TunnelStatsPoint{
			Time:          now,
			ActiveTunnels: currentActiveCount,
			TotalTunnels:  currentTotalCount,
		})
	}
	
	r.logger.Debug().
		Int("total_points", len(stats)).
		Int("final_active", currentActiveCount).
		Int("final_total", currentTotalCount).
		Msg("generateTimeSeries: time-series generation complete")
	
	return stats, nil
}

// aggregateByDay returns one data point per day with total tunnels created on that day
// For 7d view: shows data from the last 7 days from NOW (rolling window)
func (r *TunnelRepository) aggregateByDay(ctx context.Context, tunnels []*Tunnel, startTime, endTime time.Time) ([]TunnelStatsPoint, error) {
	var stats []TunnelStatsPoint
	now := time.Now()
	
	// Align startTime to day boundary (midnight)
	year, month, day := startTime.Date()
	currentDay := time.Date(year, month, day, 0, 0, 0, 0, startTime.Location())
	
	// Track seen dates to avoid duplicates
	seenDates := make(map[string]bool)
	
	// Generate one point per day from startTime to endTime (inclusive of today)
	for currentDay.Before(endTime) || currentDay.Equal(endTime) {
		dateKey := currentDay.Format("2006-01-02")
		
		// Skip if we've already processed this date
		if seenDates[dateKey] {
			currentDay = currentDay.AddDate(0, 0, 1)
			continue
		}
		seenDates[dateKey] = true
		
		// Calculate next day (midnight of next day)
		nextDay := currentDay.AddDate(0, 0, 1)
		
		activeCount := 0
		totalCount := 0
		
		// Count tunnels created on this day (from midnight to next midnight)
		for _, tunnel := range tunnels {
			// Tunnel was created on this day (inclusive of start, exclusive of next day)
			// Use a small buffer to handle timezone/rounding issues
			if (tunnel.CreatedAt.After(currentDay.Add(-time.Second)) || tunnel.CreatedAt.Equal(currentDay)) &&
			   tunnel.CreatedAt.Before(nextDay) {
				totalCount++
				r.logger.Debug().
					Str("tunnel_id", tunnel.ID).
					Str("status", tunnel.Status).
					Time("created_at", tunnel.CreatedAt).
					Time("day_start", currentDay).
					Time("day_end", nextDay).
					Msg("aggregateByDay: counting tunnel for day")
				// Count active tunnels (using current status)
				if tunnel.Status == "active" {
					activeCount++
				}
			}
		}
		
		stats = append(stats, TunnelStatsPoint{
			Time:          currentDay,
			ActiveTunnels: activeCount,
			TotalTunnels:  totalCount,
		})
		
		r.logger.Debug().
			Time("day", currentDay).
			Int("total_tunnels", totalCount).
			Int("active_tunnels", activeCount).
			Str("date_key", dateKey).
			Msg("GetTunnelStatsOverTime: aggregated by day")
		
		currentDay = nextDay
		
		// Prevent infinite loop - max 10 days
		if len(stats) > 10 {
			break
		}
	}
	
	// Always update today's point with current state (includes tunnels created just now)
	if len(stats) > 0 {
		lastPoint := &stats[len(stats)-1]
		// If the last point is today, update it with ALL tunnels created today (including just now)
		year, month, day := now.Date()
		todayStart := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
		todayEnd := todayStart.AddDate(0, 0, 1) // Midnight of next day
		
		if lastPoint.Time.Equal(todayStart) {
			// Recalculate for today - count ALL tunnels created today (from midnight to end of day)
			// This includes tunnels created just now (even if CreatedAt is slightly after now due to DB timing)
			currentActiveCount := 0
			currentTotalCount := 0
			for _, tunnel := range tunnels {
				// Tunnel created today (from midnight to end of day)
				// Use a small buffer (2 seconds) to handle timing/rounding/timezone issues
				// This ensures we catch tunnels created "just now" even if there's a slight time difference
				tunnelTime := tunnel.CreatedAt.In(now.Location()) // Convert to same timezone
				if (tunnelTime.After(todayStart.Add(-2*time.Second)) || tunnelTime.Equal(todayStart)) &&
				   tunnelTime.Before(todayEnd) {
					currentTotalCount++
					r.logger.Debug().
						Str("tunnel_id", tunnel.ID).
						Str("status", tunnel.Status).
						Time("created_at", tunnel.CreatedAt).
						Time("created_at_local", tunnelTime).
						Time("today_start", todayStart).
						Time("today_end", todayEnd).
						Time("now", now).
						Msg("aggregateByDay: counting tunnel for today")
					if tunnel.Status == "active" {
						currentActiveCount++
					}
				} else {
					r.logger.Debug().
						Str("tunnel_id", tunnel.ID).
						Time("created_at", tunnel.CreatedAt).
						Time("created_at_local", tunnelTime).
						Time("today_start", todayStart).
						Time("today_end", todayEnd).
						Msg("aggregateByDay: tunnel NOT counted for today (outside date range)")
				}
			}
			lastPoint.ActiveTunnels = currentActiveCount
			lastPoint.TotalTunnels = currentTotalCount
			
			r.logger.Debug().
				Time("today", todayStart).
				Int("total_tunnels", currentTotalCount).
				Int("active_tunnels", currentActiveCount).
				Int("all_tunnels_checked", len(tunnels)).
				Msg("GetTunnelStatsOverTime: updated today's point with current state")
		}
	}
	
	return stats, nil
}
