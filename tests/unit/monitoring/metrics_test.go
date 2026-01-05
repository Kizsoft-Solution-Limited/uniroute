package monitoring_test

import (
	"testing"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/monitoring"
)

func TestRecordRequest(t *testing.T) {
	// Test that metrics can be recorded without errors
	monitoring.RecordRequest("test-provider", "test-model", "success", 0.5)
	monitoring.RecordRequest("test-provider", "test-model", "error", 1.0)
}

func TestRecordTokens(t *testing.T) {
	monitoring.RecordTokens("test-provider", "test-model", "input", 100)
	monitoring.RecordTokens("test-provider", "test-model", "output", 50)
}

func TestRecordCost(t *testing.T) {
	monitoring.RecordCost("test-provider", "test-model", 0.001)
}

func TestSetProviderHealth(t *testing.T) {
	monitoring.SetProviderHealth("test-provider", true)
	monitoring.SetProviderHealth("test-provider", false)
}

func TestRecordRateLimitHit(t *testing.T) {
	monitoring.RecordRateLimitHit("test-key-id", "per_minute")
	monitoring.RecordRateLimitHit("test-key-id", "per_day")
}
