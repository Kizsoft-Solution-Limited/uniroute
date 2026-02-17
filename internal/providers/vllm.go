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

// VLLMProvider implements the Provider interface for vLLM (OpenAI-compatible API)
type VLLMProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
	logger  zerolog.Logger
}

// NewVLLMProvider creates a new vLLM provider. baseURL should include /v1 (e.g. http://localhost:8000/v1). apiKey is optional.
func NewVLLMProvider(baseURL, apiKey string, logger zerolog.Logger) *VLLMProvider {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &VLLMProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		logger: logger,
	}
}

// Name returns the provider name
func (p *VLLMProvider) Name() string {
	return "vllm"
}

// Chat sends a chat request to vLLM (OpenAI-compatible endpoint)
func (p *VLLMProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": convertMessagesToOpenAI(req.Messages),
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	p.logger.Debug().
		Str("provider", "vllm").
		Str("model", req.Model).
		Str("url", url).
		Msg("sending request to vLLM")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error.Message != "" {
			// Models without a chat template (e.g. facebook/opt-125m) return 400; fall back to completions
			if resp.StatusCode == http.StatusBadRequest && strings.Contains(errResp.Error.Message, "chat template") {
				return p.chatViaCompletions(ctx, req)
			}
			return nil, fmt.Errorf("vLLM API error: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("vLLM API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var openAIResp struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Role    string      `json:"role"`
				Content interface{} `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	choices := make([]Choice, 0, len(openAIResp.Choices))
	for _, c := range openAIResp.Choices {
		var contentStr string
		switch v := c.Message.Content.(type) {
		case string:
			contentStr = v
		case []interface{}:
			for _, part := range v {
				if m, ok := part.(map[string]interface{}); ok {
					if t, _ := m["type"].(string); t == "text" {
						if text, _ := m["text"].(string); text != "" {
							contentStr = text
							break
						}
					}
				}
			}
		default:
			contentStr = fmt.Sprintf("%v", v)
		}

		choices = append(choices, Choice{
			Message: Message{Role: c.Message.Role, Content: contentStr},
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

// chatViaCompletions uses /completions for models that have no chat template (e.g. facebook/opt-125m).
func (p *VLLMProvider) chatViaCompletions(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	prompt := buildPromptFromMessages(req.Messages)
	body := map[string]interface{}{
		"model":  req.Model,
		"prompt": prompt,
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal completions request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create completions request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("vLLM completions request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read completions response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vLLM completions returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var compResp struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &compResp); err != nil {
		return nil, fmt.Errorf("failed to decode completions response: %w", err)
	}

	choices := make([]Choice, 0, len(compResp.Choices))
	for _, c := range compResp.Choices {
		choices = append(choices, Choice{
			Message: Message{Role: "assistant", Content: strings.TrimSpace(c.Text)},
		})
	}

	return &ChatResponse{
		ID:      compResp.ID,
		Model:   compResp.Model,
		Choices: choices,
		Usage: Usage{
			PromptTokens:     compResp.Usage.PromptTokens,
			CompletionTokens: compResp.Usage.CompletionTokens,
			TotalTokens:      compResp.Usage.TotalTokens,
		},
	}, nil
}

// buildPromptFromMessages formats messages for completion-style models (no chat template).
func buildPromptFromMessages(messages []Message) string {
	var b strings.Builder
	for _, m := range messages {
		role := m.Role
		if role == "" {
			role = "user"
		}
		b.WriteString(strings.ToUpper(role[:1]) + role[1:] + ": ")
		b.WriteString(strings.TrimSpace(messageContentString(m.Content)))
		b.WriteString("\n\n")
	}
	b.WriteString("Assistant: ")
	return b.String()
}

func messageContentString(c interface{}) string {
	if c == nil {
		return ""
	}
	if s, ok := c.(string); ok {
		return s
	}
	if parts, ok := c.([]ContentPart); ok {
		for _, p := range parts {
			if p.Type == "text" && p.Text != "" {
				return p.Text
			}
		}
	}
	return fmt.Sprintf("%v", c)
}

// HealthCheck verifies vLLM is reachable
func (p *VLLMProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
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

// ChatStream streams chat responses from vLLM (SSE, OpenAI-compatible)
func (p *VLLMProvider) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, <-chan error) {
	chunkChan := make(chan StreamChunk, 10)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		body := map[string]interface{}{
			"model":    req.Model,
			"messages": convertMessagesToOpenAI(req.Messages),
			"stream":   true,
		}
		if req.Temperature > 0 {
			body["temperature"] = req.Temperature
		}
		if req.MaxTokens > 0 {
			body["max_tokens"] = req.MaxTokens
		}

		reqBody, err := json.Marshal(body)
		if err != nil {
			errChan <- err
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
		if err != nil {
			errChan <- err
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		if p.apiKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
		}

		resp, err := p.client.Do(httpReq)
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("vLLM API returned status %d: %s", resp.StatusCode, string(b))
			return
		}

		var responseID string
		var finalUsage *Usage
		scanner := bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				chunkChan <- StreamChunk{ID: responseID, Content: "", Done: true, Usage: finalUsage}
				return
			}

			var chunk struct {
				ID      string `json:"id"`
				Choices []struct {
					Delta        struct{ Content string `json:"content"` }
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
				Usage struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
					TotalTokens      int `json:"total_tokens"`
				} `json:"usage"`
			}

			if json.Unmarshal([]byte(data), &chunk) != nil {
				continue
			}

			if responseID == "" && chunk.ID != "" {
				responseID = chunk.ID
			}

			if len(chunk.Choices) > 0 {
				if delta := chunk.Choices[0].Delta.Content; delta != "" {
					chunkChan <- StreamChunk{ID: responseID, Content: delta, Done: false}
				}
				if chunk.Choices[0].FinishReason != "" {
					finalUsage = &Usage{
						PromptTokens:     chunk.Usage.PromptTokens,
						CompletionTokens: chunk.Usage.CompletionTokens,
						TotalTokens:      chunk.Usage.TotalTokens,
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- err
		}
	}()

	return chunkChan, errChan
}

// GetModels returns models from vLLM /v1/models (OpenAI-compatible)
func (p *VLLMProvider) GetModels() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		p.logger.Debug().Err(err).Msg("vLLM GetModels request failed")
		return nil
	}
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Debug().Err(err).Msg("vLLM GetModels failed")
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var list struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if json.NewDecoder(resp.Body).Decode(&list) != nil {
		return nil
	}

	names := make([]string, 0, len(list.Data))
	for _, m := range list.Data {
		if m.ID != "" {
			names = append(names, m.ID)
		}
	}
	return names
}
