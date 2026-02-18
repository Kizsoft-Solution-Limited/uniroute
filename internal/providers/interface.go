package providers

import "context"

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Content can be either a string (text-only) or an array of ContentPart (multimodal).
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string for text-only, []ContentPart for multimodal
}

type ContentPart struct {
	Type     string `json:"type"` // "text", "image_url", or "audio_url"
	Text     string `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
	AudioURL *AudioURL `json:"audio_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"` // Can be data URL (base64) or HTTP URL
}

type AudioURL struct {
	URL string `json:"url"` // Can be data URL (base64) or HTTP URL
}

type ChatResponse struct {
	ID        string   `json:"id"`
	Model     string   `json:"model"`
	Provider  string   `json:"provider,omitempty"` // Which provider handled this request
	Choices   []Choice `json:"choices"`
	Usage     Usage    `json:"usage"`
	Cost      float64  `json:"cost,omitempty"`       // Actual cost for this request
	LatencyMs int64    `json:"latency_ms,omitempty"` // Request latency in milliseconds
}

type Choice struct {
	Message Message `json:"message"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type StreamChunk struct {
	ID        string `json:"id,omitempty"`
	Content   string `json:"content"`   // Delta content (incremental text)
	Done      bool   `json:"done"`       // True when stream is complete
	Usage     *Usage `json:"usage,omitempty"` // Final usage stats (only in last chunk)
	Error     string `json:"error,omitempty"` // Error message if any
	Provider  string `json:"provider,omitempty"` // Provider name (for tracking)
}

// Optional interface for providers that support streaming.
type StreamingProvider interface {
	// ChatStream streams chat responses from the provider
	// Returns a channel of StreamChunk and an error channel
	ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, <-chan error)
}

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
