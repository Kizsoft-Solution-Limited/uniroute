package gateway

import (
	"sync"
	"time"
)

type LatencyTracker struct {
	mu       sync.RWMutex
	latencies map[string][]time.Duration // provider -> list of latencies
	maxSamples int // Maximum samples to keep per provider
}

func NewLatencyTracker(maxSamples int) *LatencyTracker {
	if maxSamples <= 0 {
		maxSamples = 100 // Default: keep last 100 samples
	}
	return &LatencyTracker{
		latencies:  make(map[string][]time.Duration),
		maxSamples: maxSamples,
	}
}

func (lt *LatencyTracker) RecordLatency(providerName string, latency time.Duration) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if lt.latencies[providerName] == nil {
		lt.latencies[providerName] = make([]time.Duration, 0, lt.maxSamples)
	}

	lt.latencies[providerName] = append(lt.latencies[providerName], latency)

	if len(lt.latencies[providerName]) > lt.maxSamples {
		lt.latencies[providerName] = lt.latencies[providerName][len(lt.latencies[providerName])-lt.maxSamples:]
	}
}

func (lt *LatencyTracker) GetAverageLatency(providerName string) time.Duration {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	latencies, exists := lt.latencies[providerName]
	if !exists || len(latencies) == 0 {
		return 1 * time.Second // Default latency if no data
	}

	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}

	return total / time.Duration(len(latencies))
}

func (lt *LatencyTracker) GetRecentLatency(providerName string) (time.Duration, bool) {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	latencies, exists := lt.latencies[providerName]
	if !exists || len(latencies) == 0 {
		return 0, false
	}

	return latencies[len(latencies)-1], true
}

func (lt *LatencyTracker) GetLatencyStats(providerName string) (avg, min, max time.Duration, count int) {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	latencies, exists := lt.latencies[providerName]
	if !exists || len(latencies) == 0 {
		return 1 * time.Second, 1 * time.Second, 1 * time.Second, 0
	}

	var total time.Duration
	min = latencies[0]
	max = latencies[0]

	for _, latency := range latencies {
		total += latency
		if latency < min {
			min = latency
		}
		if latency > max {
			max = latency
		}
	}

	avg = total / time.Duration(len(latencies))
	count = len(latencies)

	return avg, min, max, count
}

func (lt *LatencyTracker) Reset() {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.latencies = make(map[string][]time.Duration)
}

