package tunnel

import (
	"context"
	"sync"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/rs/zerolog"
)

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

type StatsCollector struct {
	stats       map[string]*TunnelStats
	mu          sync.RWMutex
	redisClient *storage.RedisClient
	redisMu     sync.RWMutex
	logger      zerolog.Logger
}

func NewStatsCollector(logger zerolog.Logger) *StatsCollector {
	return &StatsCollector{
		stats:  make(map[string]*TunnelStats),
		logger: logger,
	}
}

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

	if stats.TotalRequests == 1 {
		stats.AvgLatencyMs = float64(latencyMs)
	} else {
		stats.AvgLatencyMs = (stats.AvgLatencyMs*float64(stats.TotalRequests-1) + float64(latencyMs)) / float64(stats.TotalRequests)
	}

	stats.latencyHistory = append(stats.latencyHistory, int64(latencyMs))
	if len(stats.latencyHistory) > 1000 {
		stats.latencyHistory = stats.latencyHistory[len(stats.latencyHistory)-1000:]
	}

	if isError {
		stats.ErrorCount++
	}
	stats.LastRequestAt = time.Now()
	stats.mu.Unlock()

	sc.recordRequestRedis(context.Background(), tunnelID, latencyMs, requestSize, responseSize, isError)
}

func (sc *StatsCollector) GetStats(tunnelID string) *TunnelStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	stats, exists := sc.stats[tunnelID]
	if !exists {
		return &TunnelStats{TunnelID: tunnelID}
	}

	stats.mu.RLock()
	defer stats.mu.RUnlock()

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

func (sc *StatsCollector) GetAllStats() map[string]*TunnelStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	result := make(map[string]*TunnelStats)
	for id, stats := range sc.stats {
		stats.mu.RLock()
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

func (sc *StatsCollector) ResetStats(tunnelID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.stats, tunnelID)
}

func (sc *StatsCollector) GetConnectionStats(tunnelID string, openConnections int64) ConnectionStats {
	stats := sc.GetStats(tunnelID)
	if stats == nil {
		return ConnectionStats{}
	}
	return GetConnectionStatsFromTunnelStats(stats, openConnections)
}

// GetConnectionStatsFromTunnelStats builds ConnectionStats from a TunnelStats (e.g. from Redis).
func GetConnectionStatsFromTunnelStats(stats *TunnelStats, openConnections int64) ConnectionStats {
	if stats == nil {
		return ConnectionStats{}
	}
	stats.mu.RLock()
	defer stats.mu.RUnlock()
	p50 := calculatePercentile(stats.latencyHistory, 50)
	p90 := calculatePercentile(stats.latencyHistory, 90)
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

func calculatePercentile(latencies []int64, percentile int) float64 {
	if len(latencies) == 0 {
		return 0.0
	}

	sorted := make([]int64, len(latencies))
	copy(sorted, latencies)

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

	return float64(sorted[index]) / 1000.0
}

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

	return float64(sum) / float64(len(latencies)-start) / 1000.0
}
