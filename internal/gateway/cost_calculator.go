package gateway

import (
	"math"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/providers"
)

type CostCalculator struct {
	pricing map[string]map[string]Pricing
}

type Pricing struct {
	InputCost  float64 // Cost per 1M input tokens
	OutputCost float64 // Cost per 1M output tokens
}

func NewCostCalculator() *CostCalculator {
	pricing := make(map[string]map[string]Pricing)

	pricing["openai"] = map[string]Pricing{
		"gpt-4o":                  {InputCost: 5.0, OutputCost: 15.0},
		"gpt-4o-mini":             {InputCost: 0.15, OutputCost: 0.60},
		"gpt-4":                   {InputCost: 30.0, OutputCost: 60.0},
		"gpt-4-turbo":             {InputCost: 10.0, OutputCost: 30.0},
		"gpt-4-turbo-preview":     {InputCost: 10.0, OutputCost: 30.0},
		"gpt-3.5-turbo":           {InputCost: 0.5, OutputCost: 1.5},
		"gpt-3.5-turbo-0125":      {InputCost: 0.5, OutputCost: 1.5},
	}

	pricing["anthropic"] = map[string]Pricing{
		"claude-3-5-sonnet-20241022": {InputCost: 3.0, OutputCost: 15.0},
		"claude-3-5-haiku-20241022":  {InputCost: 0.80, OutputCost: 4.0},
		"claude-3-opus-20240229":     {InputCost: 15.0, OutputCost: 75.0},
		"claude-3-sonnet-20240229":   {InputCost: 3.0, OutputCost: 15.0},
		"claude-3-haiku-20240307":   {InputCost: 0.25, OutputCost: 1.25},
	}

	pricing["google"] = map[string]Pricing{
		"gemini-pro":             {InputCost: 0.0, OutputCost: 0.0},
		"gemini-1.5-pro":         {InputCost: 1.25, OutputCost: 5.0},
		"gemini-1.5-pro-latest":  {InputCost: 1.25, OutputCost: 5.0},
		"gemini-1.5-flash":       {InputCost: 0.075, OutputCost: 0.30},
		"gemini-1.5-flash-8b":    {InputCost: 0.0375, OutputCost: 0.15},
		"gemini-2.5-pro":         {InputCost: 1.25, OutputCost: 5.0},
		"gemini-2.5-flash":       {InputCost: 0.15, OutputCost: 0.60},
		"gemini-2.0-flash-exp":   {InputCost: 0.10, OutputCost: 0.40},
	}

	pricing["local"] = map[string]Pricing{
		"llama2":    {InputCost: 0.0, OutputCost: 0.0},
		"mistral":   {InputCost: 0.0, OutputCost: 0.0},
		"codellama": {InputCost: 0.0, OutputCost: 0.0},
	}

	return &CostCalculator{
		pricing: pricing,
	}
}

func (c *CostCalculator) EstimateCost(providerName, model string, messages []providers.Message) float64 {
	providerPricing, exists := c.pricing[providerName]
	if !exists {
		return 999999.0
	}

	modelPricing, exists := providerPricing[model]
	if !exists {
		for _, p := range providerPricing {
			modelPricing = p
			break
		}
		if modelPricing.InputCost == 0 && modelPricing.OutputCost == 0 {
			return 0.0 // Free if no pricing found
		}
	}

	totalChars := 0
	for _, msg := range messages {
		switch content := msg.Content.(type) {
		case string:
			totalChars += len(content)
		case []providers.ContentPart:
			for _, part := range content {
				if part.Type == "text" {
					totalChars += len(part.Text)
				}
			}
		}
	}
	estimatedInputTokens := float64(totalChars) / 4.0
	estimatedOutputTokens := estimatedInputTokens * 0.5

	inputCost := (estimatedInputTokens / 1_000_000.0) * modelPricing.InputCost
	outputCost := (estimatedOutputTokens / 1_000_000.0) * modelPricing.OutputCost

	return inputCost + outputCost
}

func (c *CostCalculator) GetPricing(providerName, model string) (Pricing, bool) {
	providerPricing, exists := c.pricing[providerName]
	if !exists {
		return Pricing{}, false
	}

	pricing, exists := providerPricing[model]
	if !exists {
		for _, p := range providerPricing {
			return p, true
		}
		return Pricing{}, false
	}

	return pricing, true
}

func (c *CostCalculator) UpdatePricing(providerName, model string, pricing Pricing) {
	if c.pricing[providerName] == nil {
		c.pricing[providerName] = make(map[string]Pricing)
	}
	c.pricing[providerName][model] = pricing
}

func (c *CostCalculator) CalculateActualCost(providerName, model string, usage providers.Usage) float64 {
	pricing, exists := c.GetPricing(providerName, model)
	if !exists {
		return 0.0
	}

	inputCost := (float64(usage.PromptTokens) / 1_000_000.0) * pricing.InputCost
	outputCost := (float64(usage.CompletionTokens) / 1_000_000.0) * pricing.OutputCost

	return math.Round((inputCost+outputCost)*10000) / 10000 // Round to 4 decimal places
}
