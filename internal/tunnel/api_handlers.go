package tunnel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// handleListTunnels returns list of active tunnels
func (ts *TunnelServer) handleListTunnels(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	ts.security.AddSecurityHeaders(w, r)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tunnels := ts.ListTunnels()

	type TunnelInfo struct {
		ID           string `json:"id"`
		Subdomain    string `json:"subdomain"`
		LocalURL     string `json:"local_url"`
		PublicURL    string `json:"public_url"`
		RequestCount int64  `json:"request_count"`
		CreatedAt    string `json:"created_at"`
		LastActive   string `json:"last_active"`
	}

	infos := make([]TunnelInfo, 0, len(tunnels))
	for _, tunnel := range tunnels {
		tunnel.mu.RLock()
		infos = append(infos, TunnelInfo{
			ID:           tunnel.ID,
			Subdomain:    tunnel.Subdomain,
			LocalURL:     tunnel.LocalURL,
			PublicURL:    "http://" + tunnel.Subdomain + ".localhost:" + string(rune(ts.port)),
			RequestCount: tunnel.RequestCount,
			CreatedAt:    tunnel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastActive:   tunnel.LastActive.Format("2006-01-02T15:04:05Z07:00"),
		})
		tunnel.mu.RUnlock()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tunnels": infos,
		"count":   len(infos),
	})
}

// handleTunnelStats returns statistics for a specific tunnel
func (ts *TunnelServer) handleTunnelStats(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	ts.security.AddSecurityHeaders(w, r)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract tunnel ID from path: /api/tunnels/{tunnel_id}
	path := strings.TrimPrefix(r.URL.Path, "/api/tunnels/")
	tunnelID := strings.TrimSuffix(path, "/stats")

	if tunnelID == "" {
		http.Error(w, "Tunnel ID required", http.StatusBadRequest)
		return
	}

	// Get tunnel connection
	ts.tunnelsMu.RLock()
	var tunnel *TunnelConnection
	for _, t := range ts.tunnels {
		if t.ID == tunnelID {
			tunnel = t
			break
		}
	}
	ts.tunnelsMu.RUnlock()

	if tunnel == nil {
		http.Error(w, "Tunnel not found", http.StatusNotFound)
		return
	}

	// Get statistics
	stats := ts.statsCollector.GetStats(tunnelID)

	// Get tunnel info
	tunnel.mu.RLock()
	tunnelInfo := map[string]interface{}{
		"id":            tunnel.ID,
		"subdomain":     tunnel.Subdomain,
		"local_url":     tunnel.LocalURL,
		"public_url":    "http://" + tunnel.Subdomain + ".localhost:" + fmt.Sprintf("%d", ts.port),
		"request_count": tunnel.RequestCount,
		"created_at":    tunnel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"last_active":   tunnel.LastActive.Format("2006-01-02T15:04:05Z07:00"),
	}
	tunnel.mu.RUnlock()

	// Calculate actual number of open connections
	// Count active TCP/UDP connections for this tunnel
	ts.tcpConnMu.RLock()
	openConnections := int64(0)
	for _, conn := range ts.tcpConnections {
		if conn.TunnelID == tunnelID {
			openConnections++
		}
	}
	ts.tcpConnMu.RUnlock()
	
	ts.udpConnMu.RLock()
	for _, conn := range ts.udpConnections {
		if conn.TunnelID == tunnelID {
			openConnections++
		}
	}
	ts.udpConnMu.RUnlock()
	
	// Get connection stats with actual open connections count
	connStats := ts.statsCollector.GetConnectionStats(tunnelID, openConnections)

	// Combine tunnel info and statistics
	response := map[string]interface{}{
		"tunnel": tunnelInfo,
		"stats": map[string]interface{}{
			"total_requests":  stats.TotalRequests,
			"total_bytes":     stats.TotalBytes,
			"avg_latency_ms":  stats.AvgLatencyMs,
			"error_count":     stats.ErrorCount,
			"last_request_at": stats.LastRequestAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		"connections": map[string]interface{}{
			"total": connStats.Total,
			"open":  connStats.Open,
			"rt1":   connStats.RT1,
			"rt5":   connStats.RT5,
			"p50":   connStats.P50,
			"p90":   connStats.P90,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListTunnelRequests returns list of requests for a tunnel
func (ts *TunnelServer) handleListTunnelRequests(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	ts.security.AddSecurityHeaders(w, r)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract tunnel ID from path: /api/tunnels/{tunnel_id}/requests
	path := strings.TrimPrefix(r.URL.Path, "/api/tunnels/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "requests" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	tunnelID := parts[0]

	if ts.repository == nil {
		http.Error(w, "Request logging not enabled", http.StatusServiceUnavailable)
		return
	}

	// Parse query parameters
	method := r.URL.Query().Get("method")
	pathFilter := r.URL.Query().Get("path")
	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	requests, err := ts.repository.ListTunnelRequests(r.Context(), tunnelID, limit, offset, method, pathFilter)
	if err != nil {
		ts.logger.Error().Err(err).Str("tunnel_id", tunnelID).Msg("Failed to list tunnel requests")
		http.Error(w, "Failed to retrieve requests", http.StatusInternalServerError)
		return
	}

	// Convert to response format (exclude large bodies for list view)
	type RequestSummary struct {
		ID           string `json:"id"`
		RequestID    string `json:"request_id"`
		Method       string `json:"method"`
		Path         string `json:"path"`
		QueryString  string `json:"query_string"`
		StatusCode   int    `json:"status_code"`
		LatencyMs    int    `json:"latency_ms"`
		RequestSize  int    `json:"request_size"`
		ResponseSize int    `json:"response_size"`
		RemoteAddr   string `json:"remote_addr"`
		UserAgent    string `json:"user_agent"`
		CreatedAt    string `json:"created_at"`
	}

	summaries := make([]RequestSummary, len(requests))
	for i, req := range requests {
		summaries[i] = RequestSummary{
			ID:           req.ID.String(),
			RequestID:    req.RequestID,
			Method:       req.Method,
			Path:         req.Path,
			QueryString:  req.QueryString,
			StatusCode:   req.StatusCode,
			LatencyMs:    req.LatencyMs,
			RequestSize:  req.RequestSize,
			ResponseSize: req.ResponseSize,
			RemoteAddr:   req.RemoteAddr,
			UserAgent:    req.UserAgent,
			CreatedAt:    req.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": summaries,
		"count":    len(summaries),
		"limit":    limit,
		"offset":   offset,
	})
}

// handleGetTunnelRequest returns a single request with full details
func (ts *TunnelServer) handleGetTunnelRequest(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	ts.security.AddSecurityHeaders(w, r)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract tunnel ID and request ID from path: /api/tunnels/{tunnel_id}/requests/{request_id}
	path := strings.TrimPrefix(r.URL.Path, "/api/tunnels/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "requests" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	tunnelID := parts[0]
	requestID := parts[2]

	if ts.repository == nil {
		http.Error(w, "Request logging not enabled", http.StatusServiceUnavailable)
		return
	}

	req, err := ts.repository.GetTunnelRequest(r.Context(), requestID)
	if err != nil {
		ts.logger.Error().Err(err).Str("request_id", requestID).Msg("Failed to get tunnel request")
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	// Verify request belongs to tunnel
	if req.TunnelID != tunnelID {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	// Convert to response format
	type RequestDetail struct {
		ID              string            `json:"id"`
		RequestID       string            `json:"request_id"`
		Method          string            `json:"method"`
		Path            string            `json:"path"`
		QueryString     string            `json:"query_string"`
		RequestHeaders  map[string]string `json:"request_headers"`
		RequestBody     string            `json:"request_body"` // Base64 encoded
		StatusCode      int               `json:"status_code"`
		ResponseHeaders map[string]string `json:"response_headers"`
		ResponseBody    string            `json:"response_body"` // Base64 encoded
		LatencyMs       int               `json:"latency_ms"`
		RequestSize     int               `json:"request_size"`
		ResponseSize    int               `json:"response_size"`
		RemoteAddr      string            `json:"remote_addr"`
		UserAgent       string            `json:"user_agent"`
		CreatedAt       string            `json:"created_at"`
	}

	detail := RequestDetail{
		ID:              req.ID.String(),
		RequestID:       req.RequestID,
		Method:          req.Method,
		Path:            req.Path,
		QueryString:     req.QueryString,
		RequestHeaders:  req.RequestHeaders,
		RequestBody:     string(req.RequestBody), // Could base64 encode if binary
		StatusCode:      req.StatusCode,
		ResponseHeaders: req.ResponseHeaders,
		ResponseBody:    string(req.ResponseBody), // Could base64 encode if binary
		LatencyMs:       req.LatencyMs,
		RequestSize:     req.RequestSize,
		ResponseSize:    req.ResponseSize,
		RemoteAddr:      req.RemoteAddr,
		UserAgent:       req.UserAgent,
		CreatedAt:       req.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

// handleReplayTunnelRequest replays a request through the tunnel
func (ts *TunnelServer) handleReplayTunnelRequest(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	ts.security.AddSecurityHeaders(w, r)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract tunnel ID and request ID from path: /api/tunnels/{tunnel_id}/requests/{request_id}/replay
	path := strings.TrimPrefix(r.URL.Path, "/api/tunnels/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[1] != "requests" || parts[3] != "replay" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	tunnelID := parts[0]
	requestID := parts[2]

	if ts.repository == nil {
		http.Error(w, "Request logging not enabled", http.StatusServiceUnavailable)
		return
	}

	// Get original request
	originalReq, err := ts.repository.GetTunnelRequest(r.Context(), requestID)
	if err != nil {
		ts.logger.Error().Err(err).Str("request_id", requestID).Msg("Failed to get tunnel request for replay")
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	// Verify request belongs to tunnel
	if originalReq.TunnelID != tunnelID {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	// Get tunnel connection
	ts.tunnelsMu.RLock()
	tunnel, exists := ts.tunnels[tunnelID]
	ts.tunnelsMu.RUnlock()

	if !exists {
		http.Error(w, "Tunnel not found or not connected", http.StatusNotFound)
		return
	}

	// Create new request ID for replay
	newRequestID := generateID()

	// Create HTTP request message
	reqData := &HTTPRequest{
		RequestID: newRequestID,
		Method:    originalReq.Method,
		Path:      originalReq.Path,
		Query:     originalReq.QueryString,
		Headers:   originalReq.RequestHeaders,
		Body:      originalReq.RequestBody,
	}

	// Send request through tunnel
	msg := TunnelMessage{
		Type:      MsgTypeHTTPRequest,
		RequestID: newRequestID,
		Request:   reqData,
	}

	// Register pending request
	pendingReq, err := ts.requestTracker.RegisterRequest(newRequestID)
	if err != nil {
		http.Error(w, "Failed to register replay request", http.StatusInternalServerError)
		return
	}

	// Send to tunnel
	if err := tunnel.WSConn.WriteJSON(msg); err != nil {
		ts.requestTracker.FailRequest(newRequestID, err)
		http.Error(w, "Failed to send replay request", http.StatusBadGateway)
		return
	}

	// Wait for response
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	response, err := pendingReq.WaitForResponse(ctx)
	if err != nil {
		ts.logger.Error().Err(err).Str("request_id", newRequestID).Msg("Replay request failed")
		http.Error(w, "Replay request failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Return replay result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"request_id":    newRequestID,
		"original_id":   requestID,
		"status_code":   response.Status,
		"response_size": len(response.Body),
		"replayed_at":   time.Now().Format("2006-01-02T15:04:05Z07:00"),
	})
}
