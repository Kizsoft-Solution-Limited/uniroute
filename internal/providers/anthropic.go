package providers

import (
	"bufio"
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

// AnthropicProvider implements the Provider interface for Anthropic (Claude)
type AnthropicProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, baseURL string, logger zerolog.Logger) *AnthropicProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	return &AnthropicProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger,
	}
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// Chat sends a chat request to Anthropic
func (p *AnthropicProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key not configured")
	}

	// Convert to Anthropic format
	anthropicReq := map[string]interface{}{
		"model":      req.Model,
		"messages":   convertMessagesToAnthropic(req.Messages),
		"max_tokens": 4096, // Anthropic requires max_tokens
	}

	if req.Temperature > 0 {
		anthropicReq["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		anthropicReq["max_tokens"] = req.MaxTokens
	}

	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/messages", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	p.logger.Debug().
		Str("provider", "anthropic").
		Str("model", req.Model).
		Msg("sending request to Anthropic")

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
			return nil, fmt.Errorf("Anthropic API error: %s", errorResp.Error.Message)
		}
		return nil, fmt.Errorf("Anthropic API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Anthropic response
	var anthropicResp struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Combine content blocks
	content := ""
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	// Convert to UniRoute format
	return &ChatResponse{
		ID:    anthropicResp.ID,
		Model: anthropicResp.Model,
		Choices: []Choice{
			{
				Message: Message{
					Role:    "assistant",
					Content: content,
				},
			},
		},
		Usage: Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}, nil
}

// HealthCheck verifies Anthropic API is accessible
func (p *AnthropicProvider) HealthCheck(ctx context.Context) error {
	if p.apiKey == "" {
		return fmt.Errorf("Anthropic API key not configured")
	}

	// Simple health check: try a minimal request
	// Anthropic doesn't have a simple health endpoint, so we'll just check API key validity
	url := fmt.Sprintf("%s/messages", p.baseURL)
	reqBody := map[string]interface{}{
		"model":      "claude-3-haiku-20240307",
		"max_tokens": 10,
		"messages": []map[string]string{
			{"role": "user", "content": "hi"},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal health check request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// ChatStream streams chat responses from Anthropic
func (p *AnthropicProvider) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, <-chan error) {
	chunkChan := make(chan StreamChunk, 10)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		if p.apiKey == "" {
			errChan <- fmt.Errorf("Anthropic API key not configured")
			return
		}

		// Convert to Anthropic format with stream=true
		anthropicReq := map[string]interface{}{
			"model":      req.Model,
			"messages":   convertMessagesToAnthropic(req.Messages),
			"max_tokens": 4096,
			"stream":     true,
		}

		if req.Temperature > 0 {
			anthropicReq["temperature"] = req.Temperature
		}
		if req.MaxTokens > 0 {
			anthropicReq["max_tokens"] = req.MaxTokens
		}

		reqBody, err := json.Marshal(anthropicReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		url := fmt.Sprintf("%s/messages", p.baseURL)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("x-api-key", p.apiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")

		resp, err := p.client.Do(httpReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			var errorResp struct {
				Error struct {
					Message string `json:"message"`
					Type    string `json:"type"`
				} `json:"error"`
			}
			if err := json.Unmarshal(body, &errorResp); err == nil {
				errChan <- fmt.Errorf("Anthropic API error: %s", errorResp.Error.Message)
			} else {
				errChan <- fmt.Errorf("Anthropic API returned status %d: %s", resp.StatusCode, string(body))
			}
			return
		}

		// Read streaming SSE response
		scanner := bufio.NewScanner(resp.Body)
		var responseID string
		var fullContent strings.Builder
		var finalUsage *Usage

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			// SSE format: "event: <type>" and "data: {...}"
			if strings.HasPrefix(line, "event: ") {
				// Skip event type line, read data on next line
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Parse SSE event
			var event struct {
				Type    string `json:"type"`
				Message struct {
					ID string `json:"id"`
				} `json:"message"`
				ContentBlock struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content_block"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
				Index int `json:"index"`
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
					TotalTokens  int `json:"total_tokens"`
				} `json:"usage"`
			}

			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			switch event.Type {
			case "message_start":
				if event.Message.ID != "" {
					responseID = event.Message.ID
				}

			case "content_block_delta":
				// Only process text deltas for the current content block
				if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
					fullContent.WriteString(event.Delta.Text)
					chunkChan <- StreamChunk{
						ID:      responseID,
						Content: event.Delta.Text,
						Done:    false,
					}
				}

			case "message_delta":
				// Update usage information
				if event.Usage.TotalTokens > 0 {
					finalUsage = &Usage{
						PromptTokens:     event.Usage.InputTokens,
						CompletionTokens: event.Usage.OutputTokens,
						TotalTokens:      event.Usage.TotalTokens,
					}
				}

			case "message_stop":
				// Send final chunk with usage
				chunkChan <- StreamChunk{
					ID:      responseID,
					Content: "",
					Done:    true,
					Usage:   finalUsage,
				}
				return

			case "error":
				errChan <- fmt.Errorf("Anthropic streaming error: %s", data)
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("failed to read stream: %w", err)
			return
		}
	}()

	return chunkChan, errChan
}

// GetModels returns list of available Anthropic models
func (p *AnthropicProvider) GetModels() []string {
	return []string{
		// Claude 3.5 series (latest, 2024)
		"claude-3-5-sonnet-20241022", // Claude 3.5 Sonnet (latest, Oct 2024)
		"claude-3-5-sonnet-20240620", // Claude 3.5 Sonnet (June 2024)
		"claude-3-5-haiku-20241022",  // Claude 3.5 Haiku (latest, Oct 2024)
		// Claude 3.0 series
		"claude-3-opus-20240229",   // Claude 3 Opus
		"claude-3-sonnet-20240229", // Claude 3 Sonnet
		"claude-3-haiku-20240307",  // Claude 3 Haiku
		// Note: When Anthropic releases Claude Sonnet 4.5, Opus 4.5, Haiku 4.5 via API,
		// their exact model IDs will be added here. Check Anthropic API documentation.
	}
}

// convertMessagesToAnthropic converts UniRoute messages to Anthropic format
// Anthropic supports multimodal content with images
func convertMessagesToAnthropic(messages []Message) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(messages))
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
			// Multimodal message - Anthropic uses array of content blocks
			contentArray := make([]map[string]interface{}, 0, len(content))
			for _, part := range content {
				if part.Type == "text" {
					contentArray = append(contentArray, map[string]interface{}{
						"type": "text",
						"text": part.Text,
					})
				} else if part.Type == "image_url" && part.ImageURL != nil {
					// Extract base64 data from data URL if present
					imageURL := part.ImageURL.URL
					if strings.HasPrefix(imageURL, "data:image/") {
						// Data URL format: data:image/png;base64,<data>
						parts := strings.SplitN(imageURL, ",", 2)
						if len(parts) == 2 {
							contentArray = append(contentArray, map[string]interface{}{
								"type": "image",
								"source": map[string]interface{}{
									"type":       "base64",
									"media_type": extractMediaType(imageURL),
									"data":       parts[1],
								},
							})
						}
					}
					// Note: HTTP URLs would need to be fetched and converted to base64
				}
			}
			messageMap["content"] = contentArray
		default:
			// Fallback: try to convert to string
			messageMap["content"] = fmt.Sprintf("%v", content)
		}

		result = append(result, messageMap)
	}
	return result
}

// extractMediaType is defined in google.go and shared across the package
