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

type GoogleProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

func NewGoogleProvider(apiKey, baseURL string, logger zerolog.Logger) *GoogleProvider {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1"
	}
	return &GoogleProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger,
	}
}

func (p *GoogleProvider) Name() string {
	return "google"
}

func (p *GoogleProvider) streamClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Minute}
}

func (p *GoogleProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("Google API key not configured")
	}

	// Convert to Google Gemini format
	googleReq := map[string]interface{}{
		"contents": convertMessagesToGoogle(req.Messages),
	}

	if req.Temperature > 0 {
		googleReq["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		googleReq["maxOutputTokens"] = req.MaxTokens
	}

	reqBody, err := json.Marshal(googleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, req.Model, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	p.logger.Debug().
		Str("provider", "google").
		Str("model", req.Model).
		Msg("sending request to Google")

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
				Status  string `json:"status"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("Google API error: %s", errorResp.Error.Message)
		}
		return nil, fmt.Errorf("Google API returned status %d: %s", resp.StatusCode, string(body))
	}

	var googleResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &googleResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract content from first candidate
	content := ""
	if len(googleResp.Candidates) > 0 && len(googleResp.Candidates[0].Content.Parts) > 0 {
		content = googleResp.Candidates[0].Content.Parts[0].Text
	}

	// Convert to UniRoute format
	return &ChatResponse{
		ID:    fmt.Sprintf("chat-%d", time.Now().Unix()),
		Model: req.Model,
		Choices: []Choice{
			{
				Message: Message{
					Role:    "assistant",
					Content: content,
				},
			},
		},
		Usage: Usage{
			PromptTokens:     googleResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: googleResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      googleResp.UsageMetadata.TotalTokenCount,
		},
	}, nil
}

// HealthCheck verifies Google API is accessible
func (p *GoogleProvider) HealthCheck(ctx context.Context) error {
	if p.apiKey == "" {
		return fmt.Errorf("Google API key not configured")
	}

	// Simple health check: list models
	url := fmt.Sprintf("%s/models?key=%s", p.baseURL, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

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

func (p *GoogleProvider) GetModels() []string {
	return []string{
		"gemini-3-pro-preview", "gemini-3-flash-preview",
		"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.5-flash-lite",
		"gemini-2.0-flash-exp",
		"gemini-1.5-pro-latest", "gemini-1.5-pro", "gemini-1.5-flash-8b", "gemini-1.5-flash",
		"gemini-pro", "gemini-pro-vision",
	}
}

func convertMessagesToGoogle(messages []Message) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		parts := make([]map[string]interface{}, 0)

		switch content := msg.Content.(type) {
		case string:
			parts = append(parts, map[string]interface{}{
				"text": content,
			})
		case []ContentPart:
			for _, part := range content {
				if part.Type == "text" {
					parts = append(parts, map[string]interface{}{
						"text": part.Text,
					})
				} else if part.Type == "image_url" && part.ImageURL != nil {
					imageURL := part.ImageURL.URL
					if strings.HasPrefix(imageURL, "data:image/") {
						urlParts := strings.SplitN(imageURL, ",", 2)
						if len(urlParts) == 2 {
							parts = append(parts, map[string]interface{}{
								"inline_data": map[string]interface{}{
									"mime_type": extractMediaType(imageURL),
									"data":      urlParts[1],
								},
							})
						}
					} else {
						parts = append(parts, map[string]interface{}{
							"file_data": map[string]interface{}{
								"mime_type": "image/jpeg",
								"file_uri":  imageURL,
							},
						})
					}
				} else if part.Type == "audio_url" && part.AudioURL != nil {
					audioURL := part.AudioURL.URL
					if strings.HasPrefix(audioURL, "data:audio/") {
						urlParts := strings.SplitN(audioURL, ",", 2)
						if len(urlParts) == 2 {
							parts = append(parts, map[string]interface{}{
								"inline_data": map[string]interface{}{
									"mime_type": extractMediaType(audioURL),
									"data":      urlParts[1],
								},
							})
						}
					}
				}
			}
		default:
			// Fallback: try to convert to string
			parts = append(parts, map[string]interface{}{
				"text": fmt.Sprintf("%v", content),
			})
		}

		result = append(result, map[string]interface{}{
			"role": msg.Role,
			"parts": parts,
		})
	}
	return result
}

// ChatStream streams chat responses from Google Gemini
func (p *GoogleProvider) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, <-chan error) {
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
			errChan <- fmt.Errorf("Google API key not configured")
			return
		}

		googleReq := map[string]interface{}{
			"contents": convertMessagesToGoogle(req.Messages),
		}
		genConfig := make(map[string]interface{})
		if req.Temperature > 0 {
			genConfig["temperature"] = req.Temperature
		}
		if req.MaxTokens > 0 {
			genConfig["maxOutputTokens"] = req.MaxTokens
		}
		if len(genConfig) > 0 {
			googleReq["generationConfig"] = genConfig
		}

		reqBody, err := json.Marshal(googleReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		// Use streamGenerateContent endpoint with alt=sse parameter
		url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s&alt=sse", p.baseURL, req.Model, p.apiKey)
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
			var errorResp struct {
				Error struct {
					Message string `json:"message"`
					Status  string `json:"status"`
				} `json:"error"`
			}
			var apiErr string
			if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
				apiErr = fmt.Sprintf("Google API error: %s", errorResp.Error.Message)
			} else {
				apiErr = fmt.Sprintf("Google API returned status %d: %s", resp.StatusCode, string(body))
			}
			chunkChan <- StreamChunk{Content: "", Done: true, Error: apiErr}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		const maxLineSize = 1024 * 1024
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, maxLineSize)
		var responseID string
		var previousText string
		var finalUsage *Usage
		var isDone bool

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
			data = strings.TrimSpace(data)
			if data == "" || data == "[DONE]" {
				continue
			}

			var geminiResp struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
						Role string `json:"role,omitempty"`
					} `json:"content"`
					FinishReason string `json:"finishReason"`
				} `json:"candidates"`
				UsageMetadata struct {
					PromptTokenCount     int `json:"promptTokenCount"`
					CandidatesTokenCount int `json:"candidatesTokenCount"`
					TotalTokenCount      int `json:"totalTokenCount"`
				} `json:"usageMetadata"`
			}

			if err := json.Unmarshal([]byte(data), &geminiResp); err != nil {
				var apiErr struct {
					Error struct {
						Message string `json:"message"`
						Status  string `json:"status"`
					} `json:"error"`
				}
				if json.Unmarshal([]byte(data), &apiErr) == nil && apiErr.Error.Message != "" {
					errChan <- fmt.Errorf("Google API: %s", apiErr.Error.Message)
					return
				}
				continue
			}

			var checkErr struct {
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			_ = json.Unmarshal([]byte(data), &checkErr)
			if checkErr.Error.Message != "" {
				errChan <- fmt.Errorf("Google API: %s", checkErr.Error.Message)
				return
			}

			if len(geminiResp.Candidates) > 0 {
				candidate := geminiResp.Candidates[0]
				if len(candidate.Content.Parts) > 0 {
					// Gemini sends full text in each chunk, calculate delta
					var currentText strings.Builder
					for _, part := range candidate.Content.Parts {
						if part.Text != "" {
							currentText.WriteString(part.Text)
						}
					}

					fullText := currentText.String()
					// Calculate delta (new content since last chunk)
					var delta string
					if strings.HasPrefix(fullText, previousText) {
						delta = fullText[len(previousText):]
					} else {
						// Content changed (shouldn't happen, but handle it)
						delta = fullText
					}

					if delta != "" {
						chunkChan <- StreamChunk{
							ID:      responseID,
							Content: delta,
							Done:    candidate.FinishReason != "",
						}
						previousText = fullText
					}

					if candidate.FinishReason != "" {
						isDone = true
					}
				}
			}

			if geminiResp.UsageMetadata.TotalTokenCount > 0 {
				finalUsage = &Usage{
					PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
					CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
					TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
				}
			}

			if isDone {
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

		if !isDone {
			chunkChan <- StreamChunk{
				ID:      responseID,
				Content: previousText,
				Done:    true,
				Usage:   finalUsage,
			}
		}
	}()

	return chunkChan, errChan
}

func extractMediaType(dataURL string) string {
	if strings.HasPrefix(dataURL, "data:") {
		parts := strings.SplitN(dataURL, ";", 2)
		if len(parts) > 0 {
			return strings.TrimPrefix(parts[0], "data:")
		}
	}
	return "image/png" // Default
}
