package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/mcp"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type MCPHandler struct {
	service *mcp.Service
	logger  zerolog.Logger
}

func NewMCPHandler(service *mcp.Service, logger zerolog.Logger) *MCPHandler {
	return &MCPHandler{service: service, logger: logger}
}

func (h *MCPHandler) ListServers(c *gin.Context) {
	urls := h.service.ServerURLs()
	c.JSON(http.StatusOK, gin.H{"servers": urls})
}

func (h *MCPHandler) ListTools(c *gin.Context) {
	serverURL := c.Query("server_url")
	if serverURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server_url is required"})
		return
	}
	decoded, err := url.QueryUnescape(serverURL)
	if err != nil {
		decoded = serverURL
	}
	if err := mcp.ValidateServerURL(decoded); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server_url"})
		return
	}
	tools, err := h.service.ListTools(c.Request.Context(), decoded)
	if err != nil {
		h.logger.Debug().Err(err).Str("server", decoded).Msg("MCP list tools failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "MCP request failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tools": tools})
}

type MCPCallRequest struct {
	ServerURL string                 `json:"server_url"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

const maxToolNameLen = 256

func (h *MCPHandler) CallTool(c *gin.Context) {
	var req MCPCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.ServerURL == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server_url and name are required"})
		return
	}
	if len(req.Name) > maxToolNameLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tool name too long"})
		return
	}
	if err := mcp.ValidateServerURL(req.ServerURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server_url"})
		return
	}
	result, err := h.service.CallTool(c.Request.Context(), req.ServerURL, req.Name, req.Arguments)
	if err != nil {
		h.logger.Debug().Err(err).Str("server", req.ServerURL).Str("tool", req.Name).Msg("MCP call tool failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "MCP request failed"})
		return
	}
	if result == nil {
		c.JSON(http.StatusOK, gin.H{"result": nil})
		return
	}
	var raw interface{}
	if err := json.Unmarshal(result, &raw); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": string(result)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": raw})
}
