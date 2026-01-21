package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, baseURL string, logger zerolog.Logger) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger,
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Chat sends a chat request to OpenAI
func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	// Convert to OpenAI format
	openAIReq := map[string]interface{}{
		"model":    req.Model,
		"messages": convertMessagesToOpenAI(req.Messages),
	}

	if req.Temperature > 0 {
		openAIReq["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		openAIReq["max_tokens"] = req.MaxTokens
	}

	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	p.logger.Debug().
		Str("provider", "openai").
		Str("model", req.Model).
		Msg("sending request to OpenAI")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("OpenAI API error: %s", errorResp.Error.Message)
		}
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse OpenAI response
	var openAIResp struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Role    string      `json:"role"`
				Content interface{} `json:"content"` // Can be string or array
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to UniRoute format
	choices := make([]Choice, 0, len(openAIResp.Choices))
	for _, choice := range openAIResp.Choices {
		// Convert content back to string (OpenAI always returns text in responses)
		var contentStr string
		switch c := choice.Message.Content.(type) {
		case string:
			contentStr = c
		case []interface{}:
			// Extract text from multimodal response
			for _, part := range c {
				if partMap, ok := part.(map[string]interface{}); ok {
					if partType, _ := partMap["type"].(string); partType == "text" {
						if text, _ := partMap["text"].(string); text != "" {
							contentStr = text
							break
						}
					}
				}
			}
		default:
			contentStr = fmt.Sprintf("%v", c)
		}

		choices = append(choices, Choice{
			Message: Message{
				Role:    choice.Message.Role,
				Content: contentStr, // Responses are always text
			},
		})
	}

	return &ChatResponse{
		ID:      openAIResp.ID,
		Model:   openAIResp.Model,
		Choices: choices,
		Usage: Usage{
			PromptTokens:     openAIResp.Usage.PromptTokens,
			CompletionTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:      openAIResp.Usage.TotalTokens,
		},
	}, nil
}

// HealthCheck verifies OpenAI API is accessible
func (p *OpenAIProvider) HealthCheck(ctx context.Context) error {
	if p.apiKey == "" {
		return fmt.Errorf("OpenAI API key not configured")
	}

	// Simple health check: list models
	url := fmt.Sprintf("%s/models", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

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

// GetModels returns list of available OpenAI models
func (p *OpenAIProvider) GetModels() []string {
	// Latest OpenAI models (as of 2025)
	return []string{
		// GPT-4o series (latest, 2024-2025)
		"gpt-4o",                 // GPT-4 Optimized (latest, May 2024)
		"gpt-4o-2024-08-06",      // GPT-4o (August 2024)
		"gpt-4o-mini",            // GPT-4o Mini (latest)
		"gpt-4o-mini-2024-07-18", // GPT-4o Mini (July 2024)
		// GPT-4 Turbo series
		"gpt-4-turbo", // GPT-4 Turbo
		"gpt-4-turbo-preview",
		"gpt-4-0125-preview",
		"gpt-4-1106-preview",
		// GPT-4 (original)
		"gpt-4",
		// GPT-3.5 series
		"gpt-3.5-turbo", // GPT-3.5 Turbo (latest)
		"gpt-3.5-turbo-0125",
		"gpt-3.5-turbo-1106",
		// Note: When OpenAI releases GPT-5.2, GPT-5.2 Pro, o4-mini via API,
		// their exact model IDs will be added here. Check OpenAI API documentation.
	}
}

// convertMessagesToOpenAI converts UniRoute messages to OpenAI format
// Supports both text-only (string) and multimodal ([]ContentPart) content
func convertMessagesToOpenAI(messages []Message) []interface{} {
	result := make([]interface{}, 0, len(messages))
	for _, msg := range messages {
		messageMap := map[string]interface{}{
			"role": msg.Role,
		}

		// Handle content: can be string (text-only) or []ContentPart (multimodal)
		switch content := msg.Content.(type) {
		case string:
			// Text-only message (backward compatible)
			messageMap["content"] = content
		case []ContentPart:
			// Multimodal message
			contentArray := make([]interface{}, 0, len(content))
			for _, part := range content {
				if part.Type == "text" {
					contentArray = append(contentArray, map[string]interface{}{
						"type": "text",
						"text": part.Text,
					})
				} else if part.Type == "image_url" && part.ImageURL != nil {
					contentArray = append(contentArray, map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]interface{}{
							"url": part.ImageURL.URL,
						},
					})
				}
			}
			messageMap["content"] = contentArray
		case []interface{}:
			// Already in OpenAI format (for direct passthrough)
			messageMap["content"] = content
		default:
			// Fallback: try to convert to string
			messageMap["content"] = fmt.Sprintf("%v", content)
		}

		result = append(result, messageMap)
	}
	return result
}
