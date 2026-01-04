package gateway

import (
	"testing"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
)

func TestNewCostCalculator(t *testing.T) {
	calculator := NewCostCalculator()
	if calculator == nil {
		t.Fatal("NewCostCalculator returned nil")
	}
}

func TestCostCalculator_EstimateCost(t *testing.T) {
	calculator := NewCostCalculator()
	
	messages := []providers.Message{
		{Role: "user", Content: "Hello, this is a test message with about 50 characters"},
	}

	// Test local (should be free)
	cost := calculator.EstimateCost("local", "llama2", messages)
	if cost != 0.0 {
		t.Errorf("Expected local cost to be 0, got %f", cost)
	}

	// Test OpenAI (should have cost)
	cost = calculator.EstimateCost("openai", "gpt-4", messages)
	if cost <= 0 {
		t.Error("Expected OpenAI cost to be > 0")
	}
}

func TestCostCalculator_CalculateActualCost(t *testing.T) {
	calculator := NewCostCalculator()
	
	usage := providers.Usage{
		PromptTokens:     1000,
		CompletionTokens: 500,
		TotalTokens:      1500,
	}

	// Test local (should be free)
	cost := calculator.CalculateActualCost("local", "llama2", usage)
	if cost != 0.0 {
		t.Errorf("Expected local cost to be 0, got %f", cost)
	}

	// Test OpenAI
	cost = calculator.CalculateActualCost("openai", "gpt-4", usage)
	if cost <= 0 {
		t.Error("Expected OpenAI cost to be > 0")
	}
}

func TestCostCalculator_GetPricing(t *testing.T) {
	calculator := NewCostCalculator()
	
	pricing, exists := calculator.GetPricing("openai", "gpt-4")
	if !exists {
		t.Error("Expected pricing to exist for gpt-4")
	}
	if pricing.InputCost <= 0 {
		t.Error("Expected input cost to be > 0")
	}
}

func TestCostCalculator_UpdatePricing(t *testing.T) {
	calculator := NewCostCalculator()
	
	newPricing := Pricing{
		InputCost:  10.0,
		OutputCost: 20.0,
	}
	
	calculator.UpdatePricing("test-provider", "test-model", newPricing)
	
	pricing, exists := calculator.GetPricing("test-provider", "test-model")
	if !exists {
		t.Error("Expected pricing to exist after update")
	}
	if pricing.InputCost != 10.0 {
		t.Errorf("Expected input cost 10.0, got %f", pricing.InputCost)
	}
}

