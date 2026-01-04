package providers

import "context"

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID        string   `json:"id"`
	Model     string   `json:"model"`
	Provider  string   `json:"provider,omitempty"` // Phase 4: Which provider handled this
	Choices   []Choice `json:"choices"`
	Usage     Usage    `json:"usage"`
	Cost      float64  `json:"cost,omitempty"`       // Phase 4: Actual cost
	LatencyMs int64    `json:"latency_ms,omitempty"` // Phase 4: Request latency
}

// Choice represents a chat choice
type Choice struct {
	Message Message `json:"message"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Provider defines the interface for all LLM providers
type Provider interface {
	// Name returns the provider's name (e.g., "local", "openai", "anthropic")
	Name() string

	// Chat sends a chat request to the provider
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// HealthCheck verifies the provider is available
	HealthCheck(ctx context.Context) error

	// GetModels returns list of available models
	GetModels() []string
}
