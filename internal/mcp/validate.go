package mcp

import (
	"fmt"
	"net/url"
	"strings"
)

const maxURLLen = 2048

func ValidateServerURL(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("MCP server URL is required")
	}
	if len(s) > maxURLLen {
		return fmt.Errorf("MCP server URL too long")
	}
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid MCP server URL")
	}
	switch u.Scheme {
	case "http", "https":
	default:
		return fmt.Errorf("MCP server URL must use http or https")
	}
	if u.Host == "" {
		return fmt.Errorf("invalid MCP server URL: missing host")
	}
	return nil
}
