package tunnel

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// TunnelStats tracks statistics for tunnels
type TunnelStats struct {
	TunnelID       string
	TotalRequests  int64
	TotalBytes     int64
	AvgLatencyMs   float64
	ErrorCount     int64
	LastRequestAt  time.Time
	latencyHistory []int64 // Recent latency measurements for percentile calculation
	mu             sync.RWMutex
}

// StatsCollector collects and aggregates tunnel statistics
type StatsCollector struct {
	stats  map[string]*TunnelStats
	mu     sync.RWMutex
	logger zerolog.Logger
}

// NewStatsCollector creates a new stats collector
func NewStatsCollector(logger zerolog.Logger) *StatsCollector {
	return &StatsCollector{
		stats:  make(map[string]*TunnelStats),
		logger: logger,
	}
}

// RecordRequest records a request in statistics
func (sc *StatsCollector) RecordRequest(tunnelID string, latencyMs int, requestSize, responseSize int, isError bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	stats, exists := sc.stats[tunnelID]
	if !exists {
		stats = &TunnelStats{
			TunnelID: tunnelID,
		}
		sc.stats[tunnelID] = stats
	}

	stats.mu.Lock()
	stats.TotalRequests++
	stats.TotalBytes += int64(requestSize + responseSize)

	// Update average latency (simple moving average)
	if stats.TotalRequests == 1 {
		stats.AvgLatencyMs = float64(latencyMs)
	} else {
		stats.AvgLatencyMs = (stats.AvgLatencyMs*float64(stats.TotalRequests-1) + float64(latencyMs)) / float64(stats.TotalRequests)
	}

	// Track latency history for percentile calculation (keep last 1000 measurements)
	stats.latencyHistory = append(stats.latencyHistory, int64(latencyMs))
	if len(stats.latencyHistory) > 1000 {
		stats.latencyHistory = stats.latencyHistory[len(stats.latencyHistory)-1000:]
	}

	if isError {
		stats.ErrorCount++
	}
	stats.LastRequestAt = time.Now()
	stats.mu.Unlock()
}

// GetStats retrieves statistics for a tunnel
func (sc *StatsCollector) GetStats(tunnelID string) *TunnelStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	stats, exists := sc.stats[tunnelID]
	if !exists {
		return &TunnelStats{TunnelID: tunnelID}
	}

	// Return a copy to avoid race conditions
	stats.mu.RLock()
	defer stats.mu.RUnlock()

	// Copy latency history
	latencyHistory := make([]int64, len(stats.latencyHistory))
	copy(latencyHistory, stats.latencyHistory)

	return &TunnelStats{
		TunnelID:       stats.TunnelID,
		TotalRequests:  stats.TotalRequests,
		TotalBytes:     stats.TotalBytes,
		AvgLatencyMs:   stats.AvgLatencyMs,
		ErrorCount:     stats.ErrorCount,
		LastRequestAt:  stats.LastRequestAt,
		latencyHistory: latencyHistory,
	}
}

// GetAllStats retrieves statistics for all tunnels
func (sc *StatsCollector) GetAllStats() map[string]*TunnelStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	result := make(map[string]*TunnelStats)
	for id, stats := range sc.stats {
		stats.mu.RLock()
		// Copy latency history
		latencyHistory := make([]int64, len(stats.latencyHistory))
		copy(latencyHistory, stats.latencyHistory)

		result[id] = &TunnelStats{
			TunnelID:       stats.TunnelID,
			TotalRequests:  stats.TotalRequests,
			TotalBytes:     stats.TotalBytes,
			AvgLatencyMs:   stats.AvgLatencyMs,
			ErrorCount:     stats.ErrorCount,
			LastRequestAt:  stats.LastRequestAt,
			latencyHistory: latencyHistory,
		}
		stats.mu.RUnlock()
	}

	return result
}

// ResetStats resets statistics for a tunnel
func (sc *StatsCollector) ResetStats(tunnelID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.stats, tunnelID)
}

// GetConnectionStats calculates connection statistics for a tunnel
func (sc *StatsCollector) GetConnectionStats(tunnelID string, openConnections int64) ConnectionStats {
	stats := sc.GetStats(tunnelID)
	if stats == nil {
		return ConnectionStats{}
	}

	stats.mu.RLock()
	defer stats.mu.RUnlock()

	// Calculate percentiles
	p50 := calculatePercentile(stats.latencyHistory, 50)
	p90 := calculatePercentile(stats.latencyHistory, 90)

	// Calculate 1-minute and 5-minute averages
	// For simplicity, we'll use recent history (last 60 and 300 measurements)
	rt1 := calculateRecentAverage(stats.latencyHistory, 60)
	rt5 := calculateRecentAverage(stats.latencyHistory, 300)

	return ConnectionStats{
		Total: stats.TotalRequests,
		Open:  openConnections,
		RT1:   rt1,
		RT5:   rt5,
		P50:   p50,
		P90:   p90,
	}
}

// calculatePercentile calculates the nth percentile from latency history
func calculatePercentile(latencies []int64, percentile int) float64 {
	if len(latencies) == 0 {
		return 0.0
	}

	// Create a copy and sort
	sorted := make([]int64, len(latencies))
	copy(sorted, latencies)

	// Simple bubble sort (fine for small arrays)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	index := (percentile * len(sorted)) / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	// Convert to seconds
	return float64(sorted[index]) / 1000.0
}

// calculateRecentAverage calculates average of recent latencies
func calculateRecentAverage(latencies []int64, count int) float64 {
	if len(latencies) == 0 {
		return 0.0
	}

	start := 0
	if len(latencies) > count {
		start = len(latencies) - count
	}

	sum := int64(0)
	for i := start; i < len(latencies); i++ {
		sum += latencies[i]
	}

	// Convert to seconds and return average
	return float64(sum) / float64(len(latencies)-start) / 1000.0
}
