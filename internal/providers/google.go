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

// GoogleProvider implements the Provider interface for Google (Gemini)
type GoogleProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

// NewGoogleProvider creates a new Google provider
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

// Name returns the provider name
func (p *GoogleProvider) Name() string {
	return "google"
}

// Chat sends a chat request to Google Gemini
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

	// Parse Google response
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

// GetModels returns list of available Google models
func (p *GoogleProvider) GetModels() []string {
	return []string{
		// Gemini 2.0 series (latest, 2025)
		"gemini-2.0-flash-exp", // Gemini 2.0 Flash (experimental)
		// Gemini 1.5 series (2024-2025)
		"gemini-1.5-pro-latest", // Gemini 1.5 Pro (latest)
		"gemini-1.5-pro",        // Gemini 1.5 Pro
		"gemini-1.5-flash-8b",   // Gemini 1.5 Flash 8B
		"gemini-1.5-flash",      // Gemini 1.5 Flash
		// Gemini 1.0 series
		"gemini-pro",        // Gemini Pro (original)
		"gemini-pro-vision", // Gemini Pro Vision
		// Note: When Google releases Gemini 3 Pro, Gemini 3 Deep Think via API,
		// their model IDs will be added here. Check Google AI Studio for latest models.
	}
}

// convertMessagesToGoogle converts UniRoute messages to Google Gemini format
// Google Gemini supports multimodal content with images
func convertMessagesToGoogle(messages []Message) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		parts := make([]map[string]interface{}, 0)

		// Handle content: can be string (text-only) or []ContentPart (multimodal)
		switch content := msg.Content.(type) {
		case string:
			// Text-only message (backward compatible)
			parts = append(parts, map[string]interface{}{
				"text": content,
			})
		case []ContentPart:
			// Multimodal message - Google Gemini uses parts array
			for _, part := range content {
				if part.Type == "text" {
					parts = append(parts, map[string]interface{}{
						"text": part.Text,
					})
				} else if part.Type == "image_url" && part.ImageURL != nil {
					// Extract base64 data from data URL if present
					imageURL := part.ImageURL.URL
					if strings.HasPrefix(imageURL, "data:image/") {
						// Data URL format: data:image/png;base64,<data>
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
						// HTTP URL - Google Gemini supports URLs directly
						parts = append(parts, map[string]interface{}{
							"file_data": map[string]interface{}{
								"mime_type": "image/jpeg", // Default, could detect from URL
								"file_uri":  imageURL,
							},
						})
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
		defer close(chunkChan)
		defer close(errChan)

		if p.apiKey == "" {
			errChan <- fmt.Errorf("Google API key not configured")
			return
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
					Status  string `json:"status"`
				} `json:"error"`
			}
			if err := json.Unmarshal(body, &errorResp); err == nil {
				errChan <- fmt.Errorf("Google API error: %s", errorResp.Error.Message)
			} else {
				errChan <- fmt.Errorf("Google API returned status %d: %s", resp.StatusCode, string(body))
			}
			return
		}

		// Read streaming SSE response
		scanner := bufio.NewScanner(resp.Body)
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

			// Parse Gemini response
			var geminiResp struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
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
				continue
			}

			// Extract text from candidates
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

					// Check if finished
					if candidate.FinishReason != "" {
						isDone = true
					}
				}
			}

			// Update usage if available
			if geminiResp.UsageMetadata.TotalTokenCount > 0 {
				finalUsage = &Usage{
					PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
					CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
					TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
				}
			}

			// Send final chunk if done
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
	}()

	return chunkChan, errChan
}

// extractMediaType extracts media type from data URL
func extractMediaType(dataURL string) string {
	if strings.HasPrefix(dataURL, "data:") {
		parts := strings.SplitN(dataURL, ";", 2)
		if len(parts) > 0 {
			return strings.TrimPrefix(parts[0], "data:")
		}
	}
	return "image/png" // Default
}
