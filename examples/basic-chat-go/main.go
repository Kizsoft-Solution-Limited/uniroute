package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages   []Message `json:"messages"`
	Temperature float64  `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func main() {
	// Get API key from environment
	apiKey := os.Getenv("UNIROUTE_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ Error: UNIROUTE_API_KEY environment variable is required")
		fmt.Println("")
		fmt.Println("Get your API key:")
		fmt.Println("  1. Run: uniroute keys create")
		fmt.Println("  2. Export: export UNIROUTE_API_KEY='ur_your_key_here'")
		os.Exit(1)
	}

	// Get API URL from environment (default: localhost)
	apiURL := os.Getenv("UNIROUTE_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8084"
	}

	// Create chat request
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello! Explain what UniRoute is in one sentence.",
			},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("❌ Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/chat", apiURL), bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("❌ Error creating request: %v\n", err)
		os.Exit(1)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("❌ Error sending request: %v\n", err)
		fmt.Println("")
		fmt.Println("💡 Make sure UniRoute server is running:")
		fmt.Println("  make dev")
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Error: Server returned status %d\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(respBody))
		os.Exit(1)
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		fmt.Printf("❌ Error parsing response: %v\n", err)
		fmt.Printf("Response: %s\n", string(respBody))
		os.Exit(1)
	}

	// Print response
	fmt.Println("✅ Chat Response:")
	fmt.Println("")
	if len(chatResp.Choices) > 0 {
		fmt.Printf("💬 %s\n", chatResp.Choices[0].Message.Content)
	}
	fmt.Println("")
	fmt.Printf("📊 Tokens: %d prompt + %d completion = %d total\n",
		chatResp.Usage.PromptTokens,
		chatResp.Usage.CompletionTokens,
		chatResp.Usage.TotalTokens)
}
