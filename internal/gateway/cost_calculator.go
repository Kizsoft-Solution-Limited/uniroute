package gateway

import (
	"math"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
)

// CostCalculator calculates estimated costs for provider requests
type CostCalculator struct {
	// Pricing per 1M tokens (input/output)
	pricing map[string]map[string]Pricing
}

// Pricing represents cost per million tokens
type Pricing struct {
	InputCost  float64 // Cost per 1M input tokens
	OutputCost float64 // Cost per 1M output tokens
}

// NewCostCalculator creates a new cost calculator with default pricing
func NewCostCalculator() *CostCalculator {
	// Default pricing (as of 2024, update as needed)
	pricing := make(map[string]map[string]Pricing)

	// OpenAI pricing
	pricing["openai"] = map[string]Pricing{
		"gpt-4":               {InputCost: 30.0, OutputCost: 60.0},
		"gpt-4-turbo-preview": {InputCost: 10.0, OutputCost: 30.0},
		"gpt-3.5-turbo":       {InputCost: 0.5, OutputCost: 1.5},
		"gpt-3.5-turbo-0125":  {InputCost: 0.5, OutputCost: 1.5},
	}

	// Anthropic pricing
	pricing["anthropic"] = map[string]Pricing{
		"claude-3-5-sonnet-20241022": {InputCost: 3.0, OutputCost: 15.0},
		"claude-3-opus-20240229":     {InputCost: 15.0, OutputCost: 75.0},
		"claude-3-sonnet-20240229":   {InputCost: 3.0, OutputCost: 15.0},
		"claude-3-haiku-20240307":    {InputCost: 0.25, OutputCost: 1.25},
	}

	// Google pricing
	pricing["google"] = map[string]Pricing{
		"gemini-pro":       {InputCost: 0.0, OutputCost: 0.0}, // Free tier
		"gemini-1.5-pro":   {InputCost: 1.25, OutputCost: 5.0},
		"gemini-1.5-flash": {InputCost: 0.075, OutputCost: 0.30},
	}

	// Local provider (free)
	pricing["local"] = map[string]Pricing{
		"llama2":    {InputCost: 0.0, OutputCost: 0.0},
		"mistral":   {InputCost: 0.0, OutputCost: 0.0},
		"codellama": {InputCost: 0.0, OutputCost: 0.0},
	}

	return &CostCalculator{
		pricing: pricing,
	}
}

// EstimateCost estimates the cost for a request
func (c *CostCalculator) EstimateCost(providerName, model string, messages []providers.Message) float64 {
	// Get pricing for provider and model
	providerPricing, exists := c.pricing[providerName]
	if !exists {
		return 999999.0 // Unknown provider = high cost
	}

	modelPricing, exists := providerPricing[model]
	if !exists {
		// Use default pricing for provider
		// Try to find any model pricing for this provider
		for _, p := range providerPricing {
			modelPricing = p
			break
		}
		if modelPricing.InputCost == 0 && modelPricing.OutputCost == 0 {
			return 0.0 // Free if no pricing found
		}
	}

	// Estimate tokens (rough approximation: 1 token ≈ 4 characters)
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Content)
	}
	estimatedInputTokens := float64(totalChars) / 4.0

	// Estimate output tokens (assume 50% of input for now)
	estimatedOutputTokens := estimatedInputTokens * 0.5

	// Calculate cost
	inputCost := (estimatedInputTokens / 1_000_000.0) * modelPricing.InputCost
	outputCost := (estimatedOutputTokens / 1_000_000.0) * modelPricing.OutputCost

	return inputCost + outputCost
}

// GetPricing returns pricing for a provider/model
func (c *CostCalculator) GetPricing(providerName, model string) (Pricing, bool) {
	providerPricing, exists := c.pricing[providerName]
	if !exists {
		return Pricing{}, false
	}

	pricing, exists := providerPricing[model]
	if !exists {
		// Return default for provider
		for _, p := range providerPricing {
			return p, true
		}
		return Pricing{}, false
	}

	return pricing, true
}

// UpdatePricing updates pricing for a provider/model
func (c *CostCalculator) UpdatePricing(providerName, model string, pricing Pricing) {
	if c.pricing[providerName] == nil {
		c.pricing[providerName] = make(map[string]Pricing)
	}
	c.pricing[providerName][model] = pricing
}

// CalculateActualCost calculates actual cost from usage
func (c *CostCalculator) CalculateActualCost(providerName, model string, usage providers.Usage) float64 {
	pricing, exists := c.GetPricing(providerName, model)
	if !exists {
		return 0.0
	}

	inputCost := (float64(usage.PromptTokens) / 1_000_000.0) * pricing.InputCost
	outputCost := (float64(usage.CompletionTokens) / 1_000_000.0) * pricing.OutputCost

	return math.Round((inputCost+outputCost)*10000) / 10000 // Round to 4 decimal places
}
