package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// LocalProvider implements the Provider interface for local LLM servers (Ollama)
type LocalProvider struct {
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

// NewLocalProvider creates a new local LLM provider
func NewLocalProvider(baseURL string, logger zerolog.Logger) *LocalProvider {
	return &LocalProvider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// Name returns the provider name
func (p *LocalProvider) Name() string {
	return "local"
}

// Chat sends a chat request to the local LLM server
func (p *LocalProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Convert messages - Ollama supports multimodal but we'll convert to text for now
	ollamaMessages := make([]Message, 0, len(req.Messages))
	for _, msg := range req.Messages {
		// Convert multimodal content to text for Ollama (or pass through if supported)
		var content interface{} = msg.Content
		if contentParts, ok := msg.Content.([]ContentPart); ok {
			// Extract text from multimodal content for Ollama
			textParts := make([]string, 0)
			for _, part := range contentParts {
				if part.Type == "text" {
					textParts = append(textParts, part.Text)
				} else if part.Type == "image_url" {
					// Ollama supports images, but for now we'll note it
					textParts = append(textParts, "[Image attached - Ollama vision models may support this]")
				}
			}
			content = strings.Join(textParts, " ")
		}
		ollamaMessages = append(ollamaMessages, Message{
			Role:    msg.Role,
			Content: content,
		})
	}

	// Convert to Ollama format
	ollamaReq := map[string]interface{}{
		"model":    req.Model,
		"messages": ollamaMessages,
	}

	if req.Temperature > 0 {
		ollamaReq["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		ollamaReq["options"] = map[string]interface{}{
			"num_predict": req.MaxTokens,
		}
	}

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	p.logger.Debug().
		Str("provider", "local").
		Str("url", url).
		Str("model", req.Model).
		Msg("sending request to local LLM")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("provider returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Ollama response
	var ollamaResp struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Done bool `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to UniRoute format
	// Ollama returns content as string, convert to our format
	var content interface{} = ollamaResp.Message.Content
	if contentStr, ok := content.(string); ok {
		content = contentStr
	}

	return &ChatResponse{
		ID:    fmt.Sprintf("chat-%d", time.Now().Unix()),
		Model: req.Model,
		Choices: []Choice{
			{
				Message: Message{
					Role:    ollamaResp.Message.Role,
					Content: content,
				},
			},
		},
		Usage: Usage{
			PromptTokens:     0, // Ollama doesn't always return this
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}, nil
}

// HealthCheck verifies the local LLM server is available
func (p *LocalProvider) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/tags", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// GetModels returns list of available models from Ollama
func (p *LocalProvider) GetModels() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/api/tags", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		p.logger.Error().Err(err).Msg("failed to create models request")
		return []string{}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Error().Err(err).Msg("failed to fetch models")
		return []string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}
	}

	var modelsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		p.logger.Error().Err(err).Msg("failed to decode models response")
		return []string{}
	}

	modelNames := make([]string, 0, len(modelsResp.Models))
	for _, model := range modelsResp.Models {
		modelNames = append(modelNames, model.Name)
	}

	return modelNames
}

