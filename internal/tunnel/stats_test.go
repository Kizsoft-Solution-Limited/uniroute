package tunnel

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewStatsCollector(t *testing.T) {
	logger := zerolog.Nop()
	collector := NewStatsCollector(logger)
	
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.stats)
}

func TestStatsCollector_RecordRequest(t *testing.T) {
	logger := zerolog.Nop()
	collector := NewStatsCollector(logger)
	
	tunnelID := "test-tunnel-1"
	
	// Record first request
	collector.RecordRequest(tunnelID, 100, 500, 1000, false)
	
	stats := collector.GetStats(tunnelID)
	assert.Equal(t, int64(1), stats.TotalRequests)
	assert.Equal(t, int64(1500), stats.TotalBytes)
	assert.Equal(t, 100.0, stats.AvgLatencyMs)
	assert.Equal(t, int64(0), stats.ErrorCount)
	
	// Record second request
	collector.RecordRequest(tunnelID, 200, 300, 700, true)
	
	stats = collector.GetStats(tunnelID)
	assert.Equal(t, int64(2), stats.TotalRequests)
	assert.Equal(t, int64(2500), stats.TotalBytes)
	assert.Equal(t, 150.0, stats.AvgLatencyMs) // (100 + 200) / 2
	assert.Equal(t, int64(1), stats.ErrorCount)
}

func TestStatsCollector_GetStats(t *testing.T) {
	logger := zerolog.Nop()
	collector := NewStatsCollector(logger)
	
	// Get stats for non-existent tunnel
	stats := collector.GetStats("non-existent")
	assert.NotNil(t, stats)
	assert.Equal(t, "non-existent", stats.TunnelID)
	assert.Equal(t, int64(0), stats.TotalRequests)
}

func TestStatsCollector_GetAllStats(t *testing.T) {
	logger := zerolog.Nop()
	collector := NewStatsCollector(logger)
	
	collector.RecordRequest("tunnel-1", 100, 100, 200, false)
	collector.RecordRequest("tunnel-2", 200, 200, 400, false)
	
	allStats := collector.GetAllStats()
	assert.Len(t, allStats, 2)
	assert.Equal(t, int64(1), allStats["tunnel-1"].TotalRequests)
	assert.Equal(t, int64(1), allStats["tunnel-2"].TotalRequests)
}

func TestStatsCollector_ResetStats(t *testing.T) {
	logger := zerolog.Nop()
	collector := NewStatsCollector(logger)
	
	collector.RecordRequest("tunnel-1", 100, 100, 200, false)
	
	stats := collector.GetStats("tunnel-1")
	assert.Equal(t, int64(1), stats.TotalRequests)
	
	collector.ResetStats("tunnel-1")
	
	stats = collector.GetStats("tunnel-1")
	assert.Equal(t, int64(0), stats.TotalRequests)
}

