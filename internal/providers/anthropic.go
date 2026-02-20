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

type AnthropicProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

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

func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

func (p *AnthropicProvider) streamClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Minute}
}

func (p *AnthropicProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key not configured")
	}

	anthropicReq := map[string]interface{}{
		"model":      req.Model,
		"messages":   convertMessagesToAnthropic(req.Messages),
		"max_tokens": 4096,
	}

	if req.Temperature > 0 {
		anthropicReq["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		anthropicReq["max_tokens"] = req.MaxTokens
	}
	if req.WebSearch {
		anthropicReq["tools"] = []map[string]string{
			{"type": "web_search_20250305", "name": "web_search"},
		}
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

	content := ""
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

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

func (p *AnthropicProvider) HealthCheck(ctx context.Context) error {
	if p.apiKey == "" {
		return fmt.Errorf("Anthropic API key not configured")
	}

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
		if req.WebSearch {
			anthropicReq["tools"] = []map[string]string{
				{"type": "web_search_20250305", "name": "web_search"},
			}
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

		resp, err := p.streamClient().Do(httpReq)
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
			var apiErr string
			if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
				apiErr = fmt.Sprintf("Anthropic API error: %s", errorResp.Error.Message)
			} else {
				apiErr = fmt.Sprintf("Anthropic API returned status %d: %s", resp.StatusCode, string(body))
			}
			chunkChan <- StreamChunk{Content: "", Done: true, Error: apiErr}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		const maxLineSize = 1024 * 1024
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, maxLineSize)
		var responseID string
		var fullContent strings.Builder
		var finalUsage *Usage

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "event: ") {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

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
				if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
					fullContent.WriteString(event.Delta.Text)
					chunkChan <- StreamChunk{
						ID:      responseID,
						Content: event.Delta.Text,
						Done:    false,
					}
				}

			case "message_delta":
				if event.Usage.TotalTokens > 0 {
					finalUsage = &Usage{
						PromptTokens:     event.Usage.InputTokens,
						CompletionTokens: event.Usage.OutputTokens,
						TotalTokens:      event.Usage.TotalTokens,
					}
				}

			case "message_stop":
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

func (p *AnthropicProvider) GetModels() []string {
	return []string{
		"claude-opus-4-6", "claude-sonnet-4-6", "claude-haiku-4-5-20251001",
		"claude-sonnet-4-5-20250929", "claude-opus-4-5-20251101",
		"claude-opus-4-1-20250805", "claude-sonnet-4-20250514", "claude-opus-4-20250514",
		"claude-3-5-sonnet-20241022", "claude-3-5-sonnet-20240620", "claude-3-5-haiku-20241022",
		"claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307",
	}
}

func convertMessagesToAnthropic(messages []Message) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		messageMap := map[string]interface{}{
			"role": msg.Role,
		}

		switch content := msg.Content.(type) {
		case string:
			messageMap["content"] = content
		case []ContentPart:
			contentArray := make([]map[string]interface{}, 0, len(content))
			for _, part := range content {
				if part.Type == "text" {
					contentArray = append(contentArray, map[string]interface{}{
						"type": "text",
						"text": part.Text,
					})
				} else if part.Type == "image_url" && part.ImageURL != nil {
					imageURL := part.ImageURL.URL
					if strings.HasPrefix(imageURL, "data:image/") {
						urlParts := strings.SplitN(imageURL, ",", 2)
						if len(urlParts) == 2 {
							b64 := strings.ReplaceAll(strings.TrimSpace(urlParts[1]), "\n", "")
							b64 = strings.ReplaceAll(b64, "\r", "")
							contentArray = append(contentArray, map[string]interface{}{
								"type": "image",
								"source": map[string]interface{}{
									"type":       "base64",
									"media_type": extractMediaType(imageURL),
									"data":       b64,
								},
							})
						}
					}
				} else if part.Type == "audio_url" && part.AudioURL != nil {
					audioURL := part.AudioURL.URL
					if strings.HasPrefix(audioURL, "data:audio/") {
						urlParts := strings.SplitN(audioURL, ",", 2)
						if len(urlParts) == 2 {
							b64 := strings.ReplaceAll(strings.TrimSpace(urlParts[1]), "\n", "")
							b64 = strings.ReplaceAll(b64, "\r", "")
							contentArray = append(contentArray, map[string]interface{}{
								"type": "audio",
								"source": map[string]interface{}{
									"type":       "base64",
									"media_type": extractMediaType(audioURL),
									"data":       b64,
								},
							})
						}
					}
				}
			}
			messageMap["content"] = contentArray
		default:
			messageMap["content"] = fmt.Sprintf("%v", content)
		}

		result = append(result, messageMap)
	}
	return result
}

