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

type OpenAIProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

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

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) streamClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Minute}
}

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

	choices := make([]Choice, 0, len(openAIResp.Choices))
	for _, choice := range openAIResp.Choices {
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

// ChatStream streams chat responses from OpenAI
func (p *OpenAIProvider) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, <-chan error) {
	chunkChan := make(chan StreamChunk, 10)
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				select {
				case errChan <- fmt.Errorf("stream panic: %v", r):
				default:
				}
			}
		}()
		defer close(chunkChan)
		defer close(errChan)

		if p.apiKey == "" {
			errChan <- fmt.Errorf("OpenAI API key not configured")
			return
		}

		// Convert to OpenAI format with stream=true
		openAIReq := map[string]interface{}{
			"model":    req.Model,
			"messages": convertMessagesToOpenAI(req.Messages),
			"stream":   true,
		}

		if req.Temperature > 0 {
			openAIReq["temperature"] = req.Temperature
		}
		if req.MaxTokens > 0 {
			openAIReq["max_tokens"] = req.MaxTokens
		}

		reqBody, err := json.Marshal(openAIReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		url := fmt.Sprintf("%s/chat/completions", p.baseURL)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

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
			if err := json.Unmarshal(body, &errorResp); err == nil {
				errChan <- fmt.Errorf("OpenAI API error: %s", errorResp.Error.Message)
			} else {
				errChan <- fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
			}
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

			// SSE format: "data: {...}"
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				chunkChan <- StreamChunk{
					ID:      responseID,
					Content: "",
					Done:    true,
					Usage:   finalUsage,
				}
				return
			}

			var streamResp struct {
				ID      string `json:"id"`
				Model   string `json:"model"`
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
						Role    string `json:"role"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
				Usage struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
					TotalTokens      int `json:"total_tokens"`
				} `json:"usage"`
			}

			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				p.logger.Debug().Err(err).Str("data", data).Msg("Failed to parse stream chunk")
				continue
			}

			if responseID == "" && streamResp.ID != "" {
				responseID = streamResp.ID
			}

			if len(streamResp.Choices) > 0 {
				delta := streamResp.Choices[0].Delta.Content
				if delta != "" {
					fullContent.WriteString(delta)
					chunkChan <- StreamChunk{
						ID:      responseID,
						Content: delta,
						Done:    false,
					}
				}

				if streamResp.Choices[0].FinishReason != "" {
					finalUsage = &Usage{
						PromptTokens:     streamResp.Usage.PromptTokens,
						CompletionTokens: streamResp.Usage.CompletionTokens,
						TotalTokens:      streamResp.Usage.TotalTokens,
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("failed to read stream: %w", err)
			return
		}
	}()

	return chunkChan, errChan
}

func (p *OpenAIProvider) GetModels() []string {
	return []string{
		"gpt-5.2", "gpt-5.2-pro", "gpt-5.1", "gpt-5", "gpt-5-mini", "gpt-5-nano",
		"gpt-4.1", "gpt-4.1-mini", "gpt-4.1-nano",
		"o3", "o3-mini", "o4-mini", "o1", "o1-pro",
		"gpt-4o", "gpt-4o-2024-08-06", "gpt-4o-mini", "gpt-4o-mini-2024-07-18",
		"gpt-4-turbo", "gpt-4-turbo-preview", "gpt-4", "gpt-3.5-turbo",
	}
}

func convertMessagesToOpenAI(messages []Message) []interface{} {
	result := make([]interface{}, 0, len(messages))
	for _, msg := range messages {
		messageMap := map[string]interface{}{
			"role": msg.Role,
		}

		switch content := msg.Content.(type) {
		case string:
			messageMap["content"] = content
		case []ContentPart:
			contentArray := make([]interface{}, 0, len(content))
			hasText := false
			for _, part := range content {
				if part.Type == "text" {
					hasText = true
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
				} else if part.Type == "audio_url" && part.AudioURL != nil {
					contentArray = append(contentArray, map[string]interface{}{
						"type": "audio_url",
						"audio_url": map[string]interface{}{
							"url": part.AudioURL.URL,
						},
					})
				}
			}
			// Some backends (e.g. vLLM) expect at least one text part when image/audio are present
			if len(contentArray) > 0 && !hasText {
				contentArray = append([]interface{}{map[string]interface{}{
					"type": "text",
					"text": "Describe this.",
				}}, contentArray...)
			}
			messageMap["content"] = contentArray
		case []interface{}:
			messageMap["content"] = content
		default:
			// Fallback: try to convert to string
			messageMap["content"] = fmt.Sprintf("%v", content)
		}

		result = append(result, messageMap)
	}
	return result
}
