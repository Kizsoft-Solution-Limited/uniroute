package tunnel

import (
	"context"
	"strconv"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
)

const (
	redisStatsKeyPrefix     = "tunnel:stats:"
	redisStatsLatencySuffix = ":latencies"
	redisStatsTTL           = 1 * time.Hour
	redisLatencyMaxLen      = 1000
)

func (sc *StatsCollector) SetRedisClient(client *storage.RedisClient) {
	sc.redisMu.Lock()
	defer sc.redisMu.Unlock()
	sc.redisClient = client
}

func (sc *StatsCollector) recordRequestRedis(ctx context.Context, tunnelID string, latencyMs int, requestSize, responseSize int, isError bool) {
	sc.redisMu.RLock()
	client := sc.redisClient
	sc.redisMu.RUnlock()
	if client == nil {
		return
	}
	key := redisStatsKeyPrefix + tunnelID
	pipe := client.Client().Pipeline()
	pipe.HIncrBy(ctx, key, "total_requests", 1)
	pipe.HIncrBy(ctx, key, "total_bytes", int64(requestSize+responseSize))
	if isError {
		pipe.HIncrBy(ctx, key, "error_count", 1)
	}
	pipe.HIncrBy(ctx, key, "sum_latency_ms", int64(latencyMs))
	pipe.HSet(ctx, key, "last_request_at", time.Now().Format(time.RFC3339))
	pipe.Expire(ctx, key, redisStatsTTL)
	latKey := key + redisStatsLatencySuffix
	pipe.LPush(ctx, latKey, latencyMs)
	pipe.LTrim(ctx, latKey, 0, redisLatencyMaxLen-1)
	pipe.Expire(ctx, latKey, redisStatsTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		sc.logger.Debug().Err(err).Str("tunnel_id", tunnelID).Msg("Failed to write stats to Redis")
	}
}

func (sc *StatsCollector) GetStatsFromRedis(ctx context.Context, tunnelID string) (*TunnelStats, error) {
	sc.redisMu.RLock()
	client := sc.redisClient
	sc.redisMu.RUnlock()
	if client == nil {
		return nil, nil
	}
	key := redisStatsKeyPrefix + tunnelID
	m, err := client.Client().HGetAll(ctx, key).Result()
	if err != nil || len(m) == 0 {
		return nil, nil
	}
	totalReq, _ := strconv.ParseInt(m["total_requests"], 10, 64)
	totalBytes, _ := strconv.ParseInt(m["total_bytes"], 10, 64)
	errorCount, _ := strconv.ParseInt(m["error_count"], 10, 64)
	sumLatency, _ := strconv.ParseInt(m["sum_latency_ms"], 10, 64)
	lastAt, _ := time.Parse(time.RFC3339, m["last_request_at"])
	avgMs := float64(0)
	if totalReq > 0 {
		avgMs = float64(sumLatency) / float64(totalReq)
	}
	latKey := key + redisStatsLatencySuffix
	latStrs, err := client.Client().LRange(ctx, latKey, 0, -1).Result()
	if err != nil {
		latStrs = nil
	}
	latencies := make([]int64, 0, len(latStrs))
	for _, s := range latStrs {
		v, _ := strconv.ParseInt(s, 10, 64)
		latencies = append(latencies, v)
	}
	// LPUSH stores newest first; we want oldest-first to match in-memory and calculateRecentAverage(..., 60/300)
	for i, j := 0, len(latencies)-1; i < j; i, j = i+1, j-1 {
		latencies[i], latencies[j] = latencies[j], latencies[i]
	}
	return &TunnelStats{
		TunnelID:       tunnelID,
		TotalRequests:  totalReq,
		TotalBytes:     totalBytes,
		AvgLatencyMs:   avgMs,
		ErrorCount:     errorCount,
		LastRequestAt:  lastAt,
		latencyHistory: latencies,
	}, nil
}
