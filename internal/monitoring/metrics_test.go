package monitoring

import (
	"testing"
)

func TestRecordRequest(t *testing.T) {
	// Test that metrics can be recorded without errors
	RecordRequest("test-provider", "test-model", "success", 0.5)
	RecordRequest("test-provider", "test-model", "error", 1.0)
}

func TestRecordTokens(t *testing.T) {
	RecordTokens("test-provider", "test-model", "input", 100)
	RecordTokens("test-provider", "test-model", "output", 50)
}

func TestRecordCost(t *testing.T) {
	RecordCost("test-provider", "test-model", 0.001)
}

func TestSetProviderHealth(t *testing.T) {
	SetProviderHealth("test-provider", true)
	SetProviderHealth("test-provider", false)
}

func TestRecordRateLimitHit(t *testing.T) {
	RecordRateLimitHit("test-key-id", "per_minute")
	RecordRateLimitHit("test-key-id", "per_day")
}

