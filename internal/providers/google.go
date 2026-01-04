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

// GoogleProvider implements the Provider interface for Google (Gemini)
type GoogleProvider struct {
	apiKey string
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
		"gemini-pro",
		"gemini-pro-vision",
		"gemini-1.5-pro",
		"gemini-1.5-flash",
	}
}

// convertMessagesToGoogle converts UniRoute messages to Google format
func convertMessagesToGoogle(messages []Message) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		result = append(result, map[string]interface{}{
			"role": msg.Role,
			"parts": []map[string]string{
				{"text": msg.Content},
			},
		})
	}
	return result
}

