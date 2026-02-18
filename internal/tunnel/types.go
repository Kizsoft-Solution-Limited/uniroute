package tunnel

import "time"

// Tunnel protocols
const (
	ProtocolHTTP = "http"
	ProtocolTCP  = "tcp"
	ProtocolTLS  = "tls"
	ProtocolUDP  = "udp"
)

// Message types for WebSocket protocol
const (
	MsgTypeInit          = "init"
	MsgTypeTunnelCreated = "tunnel_created"
	MsgTypePing          = "ping"
	MsgTypePong          = "pong"
	MsgTypeHTTPRequest   = "http_request"
	MsgTypeHTTPResponse  = "http_response"
	MsgTypeHTTPError     = "http_error"
	MsgTypeTCPData       = "tcp_data"
	MsgTypeTCPError      = "tcp_error"
	MsgTypeTLSData       = "tls_data"
	MsgTypeTLSError      = "tls_error"
	MsgTypeUDPData       = "udp_data"
	MsgTypeUDPError      = "udp_error"
	MsgTypeUpdateTunnel  = "update_tunnel"
	MsgTypeTunnelStatus  = "tunnel_status"
)

type InitMessage struct {
	Type      string                 `json:"type,omitempty"`
	Version   string                 `json:"version,omitempty"`
	Protocol  string                 `json:"protocol,omitempty"` // http, tcp, tls, udp
	LocalURL  string                 `json:"local_url"`          // For HTTP: http://localhost:port, For TCP/TLS/UDP: host:port
	Host      string                 `json:"host,omitempty"`     // Optional: specific host/subdomain to request
	Token     string                 `json:"token,omitempty"`
	Subdomain string                 `json:"subdomain,omitempty"` // For resuming existing tunnel
	TunnelID  string                 `json:"tunnel_id,omitempty"` // For resuming existing tunnel
	ForceNew  bool                   `json:"force_new,omitempty"` // If true, force creating a new tunnel (don't auto-find or resume)
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type InitResponse struct {
	Type      string `json:"type"`
	TunnelID  string `json:"tunnel_id"`
	Subdomain string `json:"subdomain"`
	PublicURL string `json:"public_url"`
	Status    string `json:"status"`
}

type TunnelMessage struct {
	Type      string        `json:"type"`
	RequestID string        `json:"request_id,omitempty"`
	Request   *HTTPRequest  `json:"request,omitempty"`
	Response  *HTTPResponse `json:"response,omitempty"`
	Error     *HTTPError    `json:"error,omitempty"`
	Data     []byte       `json:"data,omitempty"` // For TCP/TLS raw data
}

type HTTPRequest struct {
	RequestID string            `json:"request_id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Query     string            `json:"query,omitempty"`
	Headers   map[string]string `json:"headers"`
	Body      []byte            `json:"body,omitempty"`
}

type HTTPResponse struct {
	RequestID string            `json:"request_id"`
	Status    int               `json:"status"`
	Headers   map[string]string `json:"headers"`
	Body      []byte            `json:"body,omitempty"`
}

type HTTPError struct {
	RequestID string `json:"request_id"`
	Error     string `json:"error"`
	Message   string `json:"message"`
}

type Tunnel struct {
	ID           string
	UserID       string
	Subdomain    string
	CustomDomain string
	LocalURL     string
	PublicURL    string
	Protocol     string // http, tcp, tls, udp
	Status       string
	Region       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastActive   time.Time
	RequestCount int64
}

type TunnelInfo struct {
	ID        string
	Subdomain string
	PublicURL string
	Status    string
}

type TunnelStatus struct {
	TunnelID   string    `json:"tunnel_id"`
	Status     string    `json:"status"`
	Requests   int64     `json:"requests"`
	LatencyMs  int64     `json:"latency_ms"`
	UptimeSec  int64     `json:"uptime_seconds"`
	LastActive time.Time `json:"last_active"`
}

type ConnectionStats struct {
	Total int64   `json:"total"` // Total connections/requests
	Open  int64   `json:"open"`  // Open/active connections
	RT1   float64 `json:"rt1"`   // Response time 1-minute average (seconds)
	RT5   float64 `json:"rt5"`   // Response time 5-minute average (seconds)
	P50   float64 `json:"p50"`   // 50th percentile latency (seconds)
	P90   float64 `json:"p90"`   // 90th percentile latency (seconds)
}

type TunnelConfig struct {
	Name      string `json:"name"`                // Tunnel name/identifier
	Protocol  string `json:"protocol"`            // http, tcp, tls, udp
	LocalAddr string `json:"local_addr"`          // Local address (e.g., "localhost:8080" or "127.0.0.1:3306")
	Host      string `json:"host,omitempty"`      // Optional: specific host/subdomain
	ServerURL string `json:"server_url,omitempty"` // Optional: override default server URL
	Enabled   bool   `json:"enabled"`             // Whether this tunnel should be started with --all
}

type TunnelConfigFile struct {
	Version  string          `json:"version"`  // Config file version
	Tunnels  []TunnelConfig  `json:"tunnels"`  // List of tunnel configurations
	Defaults *TunnelDefaults `json:"defaults,omitempty"` // Default values
}

type TunnelDefaults struct {
	ServerURL string `json:"server_url,omitempty"` // Default tunnel server URL
}
