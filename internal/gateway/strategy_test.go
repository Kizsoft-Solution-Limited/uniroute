package gateway

import (
	"context"
	"testing"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
)

// Helper to create ChatRequest for testing
func createTestChatRequest(model string) providers.ChatRequest {
	return providers.ChatRequest{
		Model: model,
		Messages: []providers.Message{
			{Role: "user", Content: "test"},
		},
	}
}

func TestModelBasedStrategy(t *testing.T) {
	strategy := &ModelBasedStrategy{}
	
	providers := []providers.Provider{
		&mockProvider{name: "provider1", models: []string{"model1"}},
		&mockProvider{name: "provider2", models: []string{"model2"}},
	}

	req := createTestChatRequest("model1")

	selected, err := strategy.SelectProvider(context.Background(), req, providers)
	if err != nil {
		t.Fatalf("SelectProvider returned error: %v", err)
	}
	if selected.Name() != "provider1" {
		t.Errorf("Expected provider1, got %s", selected.Name())
	}
}

func TestCostBasedStrategy(t *testing.T) {
	calculator := NewCostCalculator()
	strategy := NewCostBasedStrategy(calculator)
	
	providers := []providers.Provider{
		&mockProvider{name: "local", models: []string{"model1"}}, // Free
		&mockProvider{name: "openai", models: []string{"model1"}}, // Paid
	}

	req := createTestChatRequest("model1")

	selected, err := strategy.SelectProvider(context.Background(), req, providers)
	if err != nil {
		t.Fatalf("SelectProvider returned error: %v", err)
	}
	// Should select local (cheapest/free)
	if selected.Name() != "local" {
		t.Errorf("Expected local (cheapest), got %s", selected.Name())
	}
}

func TestLoadBalancedStrategy(t *testing.T) {
	strategy := NewLoadBalancedStrategy()
	
	providers := []providers.Provider{
		&mockProvider{name: "provider1", models: []string{"model1"}},
		&mockProvider{name: "provider2", models: []string{"model1"}},
		&mockProvider{name: "provider3", models: []string{"model1"}},
	}

	req := createTestChatRequest("model1")

	// First request
	selected1, err := strategy.SelectProvider(context.Background(), req, providers)
	if err != nil {
		t.Fatalf("SelectProvider returned error: %v", err)
	}

	// Second request (should be different)
	selected2, err := strategy.SelectProvider(context.Background(), req, providers)
	if err != nil {
		t.Fatalf("SelectProvider returned error: %v", err)
	}

	// Third request
	selected3, err := strategy.SelectProvider(context.Background(), req, providers)
	if err != nil {
		t.Fatalf("SelectProvider returned error: %v", err)
	}

	// Should cycle through providers
	if selected1.Name() == selected2.Name() && selected2.Name() == selected3.Name() {
		t.Error("Load balanced strategy should cycle through providers")
	}
}

func TestLatencyBasedStrategy(t *testing.T) {
	tracker := NewLatencyTracker(100)
	tracker.RecordLatency("fast", 100*time.Millisecond)
	tracker.RecordLatency("slow", 1000*time.Millisecond)
	
	strategy := NewLatencyBasedStrategy(tracker)
	
	providers := []providers.Provider{
		&mockProvider{name: "fast", models: []string{"model1"}},
		&mockProvider{name: "slow", models: []string{"model1"}},
	}

	req := createTestChatRequest("model1")

	selected, err := strategy.SelectProvider(context.Background(), req, providers)
	if err != nil {
		t.Fatalf("SelectProvider returned error: %v", err)
	}
	// Should select fast provider
	if selected.Name() != "fast" {
		t.Errorf("Expected fast provider, got %s", selected.Name())
	}
}

