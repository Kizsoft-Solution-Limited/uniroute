package tunnel

import "time"

// Message types for WebSocket protocol
const (
	MsgTypeInit          = "init"
	MsgTypeTunnelCreated = "tunnel_created"
	MsgTypePing          = "ping"
	MsgTypePong          = "pong"
	MsgTypeHTTPRequest   = "http_request"
	MsgTypeHTTPResponse  = "http_response"
	MsgTypeHTTPError     = "http_error"
	MsgTypeUpdateTunnel  = "update_tunnel"
	MsgTypeTunnelStatus  = "tunnel_status"
)

// InitMessage is sent by client to initialize tunnel
type InitMessage struct {
	Type      string                 `json:"type,omitempty"`
	Version   string                 `json:"version,omitempty"`
	LocalURL  string                 `json:"local_url"`
	Token     string                 `json:"token,omitempty"`
	Subdomain string                 `json:"subdomain,omitempty"` // For resuming existing tunnel
	TunnelID  string                 `json:"tunnel_id,omitempty"` // For resuming existing tunnel
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// InitResponse is sent by server after tunnel creation
type InitResponse struct {
	Type      string `json:"type"`
	TunnelID  string `json:"tunnel_id"`
	Subdomain string `json:"subdomain"`
	PublicURL string `json:"public_url"`
	Status    string `json:"status"`
}

// TunnelMessage is the base message type
type TunnelMessage struct {
	Type      string        `json:"type"`
	RequestID string        `json:"request_id,omitempty"`
	Request   *HTTPRequest  `json:"request,omitempty"`
	Response  *HTTPResponse `json:"response,omitempty"`
	Error     *HTTPError    `json:"error,omitempty"`
}

// HTTPRequest represents an HTTP request to forward
type HTTPRequest struct {
	RequestID string            `json:"request_id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Query     string            `json:"query,omitempty"`
	Headers   map[string]string `json:"headers"`
	Body      []byte            `json:"body,omitempty"`
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	RequestID string            `json:"request_id"`
	Status    int               `json:"status"`
	Headers   map[string]string `json:"headers"`
	Body      []byte            `json:"body,omitempty"`
}

// HTTPError represents an HTTP error
type HTTPError struct {
	RequestID string `json:"request_id"`
	Error     string `json:"error"`
	Message   string `json:"message"`
}

// Tunnel represents tunnel metadata (database model)
type Tunnel struct {
	ID           string
	UserID       string
	Subdomain    string
	CustomDomain string
	LocalURL     string
	PublicURL    string
	Status       string
	Region       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastActive   time.Time
	RequestCount int64
}

// TunnelInfo contains tunnel connection information
type TunnelInfo struct {
	ID        string
	Subdomain string
	PublicURL string
	Status    string
}

// TunnelStatus contains tunnel statistics
type TunnelStatus struct {
	TunnelID   string    `json:"tunnel_id"`
	Status     string    `json:"status"`
	Requests   int64     `json:"requests"`
	LatencyMs  int64     `json:"latency_ms"`
	UptimeSec  int64     `json:"uptime_seconds"`
	LastActive time.Time `json:"last_active"`
}

// ConnectionStats represents connection statistics in ngrok-style format
type ConnectionStats struct {
	Total int64   `json:"total"` // Total connections/requests
	Open  int64   `json:"open"`  // Open/active connections
	RT1   float64 `json:"rt1"`   // Response time 1-minute average (seconds)
	RT5   float64 `json:"rt5"`   // Response time 5-minute average (seconds)
	P50   float64 `json:"p50"`   // 50th percentile latency (seconds)
	P90   float64 `json:"p90"`   // 90th percentile latency (seconds)
}
