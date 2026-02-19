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

type LocalProvider struct {
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

func NewLocalProvider(baseURL string, logger zerolog.Logger) *LocalProvider {
	return &LocalProvider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (p *LocalProvider) streamClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Minute}
}

func (p *LocalProvider) Name() string {
	return "local"
}

func convertToOllamaMessages(messages []Message) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		om := map[string]interface{}{"role": msg.Role}
		switch content := msg.Content.(type) {
		case string:
			om["content"] = content
		case []ContentPart:
			var textParts []string
			var images []string
			hasAudio := false
			for _, part := range content {
				if part.Type == "text" && part.Text != "" {
					textParts = append(textParts, part.Text)
				} else if part.Type == "image_url" && part.ImageURL != nil {
					u := part.ImageURL.URL
					if strings.HasPrefix(strings.ToLower(u), "data:image/") {
						parts := strings.SplitN(u, ",", 2)
						if len(parts) == 2 {
							images = append(images, parts[1])
						}
					}
				} else if part.Type == "audio_url" {
					hasAudio = true
					textParts = append(textParts, "[Audio attached]")
				}
			}
			textContent := strings.TrimSpace(strings.Join(textParts, " "))
			if textContent == "" && len(images) > 0 {
				textContent = "Describe this image."
			}
			if textContent == "" && hasAudio {
				textContent = "Transcribe or describe this audio."
			}
			om["content"] = textContent
			if len(images) > 0 {
				om["images"] = images
			}
		default:
			om["content"] = fmt.Sprintf("%v", content)
		}
		out = append(out, om)
	}
	return out
}

func (p *LocalProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	ollamaMessages := convertToOllamaMessages(req.Messages)

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

// ChatStream streams chat responses from Ollama
func (p *LocalProvider) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, <-chan error) {
	chunkChan := make(chan StreamChunk, 10)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		ollamaMessages := convertToOllamaMessages(req.Messages)
		ollamaReq := map[string]interface{}{
			"model":    req.Model,
			"messages": ollamaMessages,
			"stream":   true,
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
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		url := fmt.Sprintf("%s/api/chat", p.baseURL)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := p.streamClient().Do(httpReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("provider returned status %d: %s", resp.StatusCode, string(body))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		const maxLineSize = 1024 * 1024
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, maxLineSize)
		var responseID string
		var previousContent string
		var finalUsage *Usage

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			var ollamaResp struct {
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				Done            bool  `json:"done"`
				TotalDuration   int64 `json:"total_duration,omitempty"`
				PromptEvalCount int   `json:"prompt_eval_count,omitempty"`
				EvalCount       int   `json:"eval_count,omitempty"`
			}

			if err := json.Unmarshal([]byte(line), &ollamaResp); err != nil {
				continue
			}

			// Extract incremental content (delta)
			// Ollama sends the full content in each message, so we need to calculate the delta
			currentContent := ollamaResp.Message.Content
			if currentContent != "" && currentContent != previousContent {
				// Calculate delta: new content since last message
				var delta string
				if strings.HasPrefix(currentContent, previousContent) {
					delta = currentContent[len(previousContent):]
				} else {
					// Content changed completely (shouldn't happen, but handle it)
					delta = currentContent
				}

				if delta != "" {
					chunkChan <- StreamChunk{
						ID:      responseID,
						Content: delta,
						Done:    ollamaResp.Done,
					}
				}

				previousContent = currentContent
			}

			if ollamaResp.Done {
				if ollamaResp.PromptEvalCount > 0 || ollamaResp.EvalCount > 0 {
					finalUsage = &Usage{
						PromptTokens:     ollamaResp.PromptEvalCount,
						CompletionTokens: ollamaResp.EvalCount,
						TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
					}
				}

				// Send final chunk
				chunkChan <- StreamChunk{
					ID:      responseID,
					Content: "",
					Done:    true,
					Usage:   finalUsage,
				}
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

