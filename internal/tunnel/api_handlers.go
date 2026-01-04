package tunnel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// handleListTunnels returns list of active tunnels
func (ts *TunnelServer) handleListTunnels(w http.ResponseWriter, r *http.Request) {
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

	// Get connection stats
	connStats := ts.statsCollector.GetConnectionStats(tunnelID, 1) // 1 open connection for this tunnel

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
