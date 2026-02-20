package providers

import "context"

type ChatRequest struct {
	Model                 string    `json:"model"`
	Messages              []Message `json:"messages"`
	Temperature           float64   `json:"temperature,omitempty"`
	MaxTokens             int       `json:"max_tokens,omitempty"`
	GoogleSearchGrounding bool `json:"google_search_grounding,omitempty"`
	WebSearch             bool `json:"web_search,omitempty"`
}

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type ContentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
	AudioURL *AudioURL `json:"audio_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type AudioURL struct {
	URL string `json:"url"`
}

type ChatResponse struct {
	ID        string   `json:"id"`
	Model     string   `json:"model"`
	Provider  string   `json:"provider,omitempty"`
	Choices   []Choice `json:"choices"`
	Usage     Usage    `json:"usage"`
	Cost      float64  `json:"cost,omitempty"`
	LatencyMs int64    `json:"latency_ms,omitempty"`
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
	ID       string `json:"id,omitempty"`
	Content  string `json:"content"`
	Done     bool   `json:"done"`
	Usage    *Usage `json:"usage,omitempty"`
	Error    string `json:"error,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type StreamingProvider interface {
	ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, <-chan error)
}

type Provider interface {
	Name() string
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	HealthCheck(ctx context.Context) error
	GetModels() []string
}
