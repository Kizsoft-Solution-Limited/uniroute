package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	ProtocolVersion = "2024-11-05"
	ClientName      = "uniroute"
	ClientVersion   = "1.0.0"
)

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema,omitempty"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type Client struct {
	serverURL  string
	httpClient *http.Client
	sessionID  string
	mu         sync.Mutex
	nextID     int
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL:  serverURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		nextID:     1,
	}
}

func (c *Client) do(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	c.mu.Lock()
	id := c.nextID
	c.nextID++
	c.mu.Unlock()

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", ProtocolVersion)
	if c.sessionID != "" {
		req.Header.Set("MCP-Session-Id", c.sessionID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Header.Get("MCP-Session-Id") != "" {
		c.mu.Lock()
		c.sessionID = resp.Header.Get("MCP-Session-Id")
		c.mu.Unlock()
	}

	if resp.StatusCode == http.StatusAccepted {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mcp server returned %d", resp.StatusCode)
	}

	const maxResponseSize = 1 << 20
	var rpcResp JSONRPCResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseSize)).Decode(&rpcResp); err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("mcp error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}

func (c *Client) Initialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: ProtocolVersion,
		Capabilities: map[string]interface{}{
			"roots": map[string]bool{"listChanged": true},
		},
	}
	params.ClientInfo.Name = ClientName
	params.ClientInfo.Version = ClientVersion

	_, err := c.do(ctx, "initialize", params)
	if err != nil {
		return err
	}

	notif := JSONRPCRequest{JSONRPC: "2.0", Method: "notifications/initialized"}
	notifBody, _ := json.Marshal(notif)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL, bytes.NewReader(notifBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", ProtocolVersion)
	if c.sessionID != "" {
		req.Header.Set("MCP-Session-Id", c.sessionID)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("initialized notification: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	result, err := c.do(ctx, "tools/list", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	var list ToolsListResult
	if err := json.Unmarshal(result, &list); err != nil {
		return nil, err
	}
	return list.Tools, nil
}

func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (json.RawMessage, error) {
	params := CallToolParams{Name: name, Arguments: arguments}
	return c.do(ctx, "tools/call", params)
}

func (c *Client) Connect(ctx context.Context) error {
	return c.Initialize(ctx)
}
