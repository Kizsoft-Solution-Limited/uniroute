package handlers

import (
	"net/http"
	"strings"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/gin-gonic/gin"
)

type SwaggerHandler struct {
	jwtService *security.JWTService
}

func NewSwaggerHandler(jwtService *security.JWTService) *SwaggerHandler {
	return &SwaggerHandler{
		jwtService: jwtService,
	}
}

func (h *SwaggerHandler) HandleSwaggerUI(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>UniRoute API Documentation</title>
	<link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui.css" />
	<style>
		html {
			box-sizing: border-box;
			overflow: -moz-scrollbars-vertical;
			overflow-y: scroll;
		}
		*, *:before, *:after {
			box-sizing: inherit;
		}
		body {
			margin: 0;
			background: #fafafa;
		}
		#swagger-ui {
			padding: 20px;
		}
		.loading {
			text-align: center;
			padding: 40px;
			font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
		}
	</style>
</head>
<body>
	<div id="swagger-ui">
		<div class="loading">
			<h2>Loading API Documentation...</h2>
			<p>Please wait while Swagger UI loads.</p>
		</div>
	</div>
	<script src="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui-bundle.js"></script>
	<script src="https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui-standalone-preset.js"></script>
	<script>
		window.onload = function() {
			// Get the base URL from the current location
			const baseUrl = window.location.protocol + "//" + window.location.host;
			// Get access_token from URL query parameter
			const urlParams = new URLSearchParams(window.location.search);
			const accessToken = urlParams.get('access_token');
			let swaggerUrl = baseUrl + "/swagger.json";
			if (accessToken) {
				swaggerUrl += "?access_token=" + encodeURIComponent(accessToken);
			}
			
			try {
				const ui = SwaggerUIBundle({
					url: swaggerUrl,
					dom_id: '#swagger-ui',
					deepLinking: true,
					presets: [
						SwaggerUIBundle.presets.apis,
						SwaggerUIStandalonePreset
					],
					plugins: [
						SwaggerUIBundle.plugins.DownloadUrl
					],
					layout: "StandaloneLayout",
					validatorUrl: null, // Disable validator to avoid external requests
					tryItOutEnabled: true,
					requestInterceptor: function(request) {
						// Add access_token to all requests if available
						if (accessToken) {
							if (!request.headers) {
								request.headers = {};
							}
							request.headers['Authorization'] = 'Bearer ' + accessToken;
						}
						return request;
					},
					onComplete: function() {},
					onFailure: function(data) {
						console.error("Swagger UI failed to load:", data);
						document.getElementById('swagger-ui').innerHTML = 
							'<div style="padding: 20px; text-align: center;">' +
							'<h2>Failed to load API documentation</h2>' +
							'<p>Error: ' + (data.message || 'Unknown error') + '</p>' +
							'<p>Please check the browser console for more details.</p>' +
							'</div>';
					}
				});
			} catch (error) {
				console.error("Error initializing Swagger UI:", error);
				document.getElementById('swagger-ui').innerHTML = 
					'<div style="padding: 20px; text-align: center;">' +
					'<h2>Error loading Swagger UI</h2>' +
					'<p>' + error.message + '</p>' +
					'</div>';
			}
		};
	</script>
</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func (h *SwaggerHandler) HandleSwaggerJSON(c *gin.Context) {
	accessToken := c.Query("access_token")

	var userRoles []string
	var isAdmin bool
	if accessToken != "" && h.jwtService != nil {
		claims, err := h.jwtService.ValidateToken(accessToken)
		if err == nil {
			userRoles = claims.Roles
			for _, role := range userRoles {
				if role == "admin" {
					isAdmin = true
					break
				}
			}
		}
	}

	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "UniRoute API",
			"description": "One unified gateway for every AI model. Route, secure, and manage traffic to any LLM—cloud or local—with one unified platform.",
			"version":     "1.0.0",
			"contact": map[string]interface{}{
				"name":  "UniRoute Support",
				"url":   "https://github.com/Kizsoft-Solution-Limited/uniroute",
				"email": "support@uniroute.co",
			},
		},
		"servers": []map[string]interface{}{
			{
				"url":         "http://localhost:8084",
				"description": "Local development server",
			},
		},
		"tags": []map[string]interface{}{
			{"name": "Health", "description": "Health check endpoints"},
			{"name": "Authentication", "description": "User authentication and registration"},
			{"name": "Chat", "description": "AI chat completion endpoints"},
			{"name": "Conversations", "description": "Conversation management"},
			{"name": "Providers", "description": "LLM provider management"},
			{"name": "Analytics", "description": "Usage analytics and metrics"},
			{"name": "Routing", "description": "Intelligent routing and cost estimation"},
			{"name": "Tunnels", "description": "Tunnel management"},
			{"name": "Domains", "description": "Custom domain management"},
			{"name": "Admin", "description": "Administrative endpoints (admin only)"},
		},
		"paths": map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Health"},
					"summary":     "Health check",
					"description": "Check if the API server is running",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Server is healthy",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"status": map[string]interface{}{
												"type":    "string",
												"example": "ok",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/auth/register": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Register new user",
					"description": "Create a new user account. A verification email will be sent.",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"email", "password", "name"},
									"properties": map[string]interface{}{
										"email": map[string]interface{}{
											"type":    "string",
											"format":  "email",
											"example": "user@example.com",
										},
										"password": map[string]interface{}{
											"type":    "string",
											"format":  "password",
											"example": "securePassword123",
										},
										"name": map[string]interface{}{
											"type":    "string",
											"example": "John Doe",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "User registered successfully",
						},
						"400": map[string]interface{}{
							"description": "Invalid request",
						},
						"409": map[string]interface{}{
							"description": "User already exists",
						},
					},
				},
			},
			"/auth/login": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "User login",
					"description": "Authenticate user and receive JWT token",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"email", "password"},
									"properties": map[string]interface{}{
										"email": map[string]interface{}{
											"type":    "string",
											"format":  "email",
											"example": "user@example.com",
										},
										"password": map[string]interface{}{
											"type":    "string",
											"format":  "password",
											"example": "securePassword123",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Login successful",
						},
						"401": map[string]interface{}{
							"description": "Invalid credentials",
						},
						"403": map[string]interface{}{
							"description": "Email not verified",
						},
					},
				},
			},
			"/auth/logout": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "User logout",
					"description": "Logout current user (client-side token removal)",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Logout successful",
						},
					},
				},
			},
			"/auth/password-reset": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Request password reset",
					"description": "Request a password reset email",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"email"},
									"properties": map[string]interface{}{
										"email": map[string]interface{}{
											"type":    "string",
											"format":  "email",
											"example": "user@example.com",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Password reset email sent (if email exists)",
						},
					},
				},
			},
			"/auth/password-reset/confirm": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Confirm password reset",
					"description": "Reset password using token from email",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"token", "password"},
									"properties": map[string]interface{}{
										"token": map[string]interface{}{
											"type":    "string",
											"example": "reset-token-here",
										},
										"password": map[string]interface{}{
											"type":    "string",
											"format":  "password",
											"example": "newSecurePassword123",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Password reset successful",
						},
						"400": map[string]interface{}{
							"description": "Invalid or expired token",
						},
					},
				},
			},
			"/auth/verify-email": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Verify email",
					"description": "Verify user email with token from verification email",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"token"},
									"properties": map[string]interface{}{
										"token": map[string]interface{}{
											"type":    "string",
											"example": "verification-token-here",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Email verified successfully",
						},
						"400": map[string]interface{}{
							"description": "Invalid or expired token",
						},
					},
				},
			},
			"/auth/resend-verification": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Resend verification email",
					"description": "Resend email verification link",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"email"},
									"properties": map[string]interface{}{
										"email": map[string]interface{}{
											"type":    "string",
											"format":  "email",
											"example": "user@example.com",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Verification email sent",
						},
					},
				},
			},
			"/auth/profile": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Get user profile",
					"description": "Get current authenticated user's profile",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "User profile",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"id": map[string]interface{}{
												"type":    "string",
												"example": "uuid-here",
											},
											"email": map[string]interface{}{
												"type":    "string",
												"example": "user@example.com",
											},
											"name": map[string]interface{}{
												"type":    "string",
												"example": "John Doe",
											},
											"email_verified": map[string]interface{}{
												"type":    "boolean",
												"example": true,
											},
											"roles": map[string]interface{}{
												"type": "array",
												"items": map[string]interface{}{
													"type": "string",
												},
												"example": []string{"user"},
											},
										},
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
				"put": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Update user profile",
					"description": "Update current user's profile (name only - roles cannot be changed)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"name": map[string]interface{}{
											"type":    "string",
											"example": "John Doe",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Profile updated successfully",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			},
			"/auth/refresh": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Refresh token",
					"description": "Refresh JWT authentication token",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "New token generated",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"token": map[string]interface{}{
												"type":    "string",
												"example": "new-jwt-token",
											},
										},
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			},
			"/v1/chat": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Chat"},
					"summary":     "Chat completion",
					"description": "Send a chat completion request to any LLM provider",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"model", "messages"},
									"properties": map[string]interface{}{
										"model": map[string]interface{}{
											"type":        "string",
											"example":     "llama2",
											"description": "Model name (e.g., llama2, gpt-4, claude-3)",
										},
										"messages": map[string]interface{}{
											"type": "array",
											"items": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"role": map[string]interface{}{
														"type":    "string",
														"enum":    []string{"user", "assistant", "system"},
														"example": "user",
													},
													"content": map[string]interface{}{
														"type":    "string",
														"example": "Hello!",
													},
												},
											},
										},
										"temperature": map[string]interface{}{
											"type":    "number",
											"format":  "float",
											"example": 0.7,
										},
										"max_tokens": map[string]interface{}{
											"type":    "integer",
											"example": 1000,
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Chat completion successful",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized - Invalid API key",
						},
						"429": map[string]interface{}{
							"description": "Rate limit exceeded",
						},
					},
				},
			},
			"/v1/providers": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Providers"},
					"summary":     "List providers",
					"description": "Get list of available LLM providers",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of providers",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			},
			"/v1/providers/{name}/health": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Providers"},
					"summary":     "Get provider health",
					"description": "Check health status of a specific provider",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name":        "name",
							"in":          "path",
							"required":    true,
							"description": "Provider name",
							"schema": map[string]interface{}{
								"type":    "string",
								"example": "openai",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Provider health status",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			},
			"/v1/analytics/usage": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Analytics"},
					"summary":     "Get usage statistics",
					"description": "Get usage statistics and metrics",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Usage statistics",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			},
			"/v1/analytics/requests": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Analytics"},
					"summary":     "Get request history",
					"description": "Get history of API requests",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Request history",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			},
			"/v1/routing/estimate-cost": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Routing"},
					"summary":     "Estimate cost",
					"description": "Estimate cost for a chat request",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Cost estimate",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			},
			"/v1/routing/latency": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Routing"},
					"summary":     "Get latency stats",
					"description": "Get latency statistics for providers",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Latency statistics",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			},
			"/auth/api-keys": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Create API key",
					"description": "Create a new API key for the authenticated user",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"name"},
									"properties": map[string]interface{}{
										"name": map[string]interface{}{
											"type":    "string",
											"example": "My API Key",
										},
										"rate_limit_per_minute": map[string]interface{}{
											"type":    "integer",
											"example": 60,
										},
										"rate_limit_per_day": map[string]interface{}{
											"type":    "integer",
											"example": 10000,
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "API key created",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
				"get": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "List API keys",
					"description": "List all API keys for the authenticated user",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of API keys",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			},
			"/auth/api-keys/{id}": map[string]interface{}{
				"delete": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Revoke API key",
					"description": "Revoke an API key belonging to the authenticated user",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"description": "API key ID",
							"schema": map[string]interface{}{
								"type":    "string",
								"example": "uuid-here",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "API key revoked",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"404": map[string]interface{}{
							"description": "API key not found",
						},
					},
				},
			},
			"/auth/provider-keys": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Add provider key",
					"description": "Add a provider API key for the authenticated user (BYOK)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"provider", "api_key"},
									"properties": map[string]interface{}{
										"provider": map[string]interface{}{
											"type":    "string",
											"example": "openai",
										},
										"api_key": map[string]interface{}{
											"type":    "string",
											"example": "sk-...",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Provider key added",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
				"get": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "List provider keys",
					"description": "List all provider keys for the authenticated user (BYOK)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of provider keys",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			},
			"/auth/provider-keys/{provider}": map[string]interface{}{
				"put": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Update provider key",
					"description": "Update a provider API key for the authenticated user (BYOK)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name":        "provider",
							"in":          "path",
							"required":    true,
							"description": "Provider name",
							"schema": map[string]interface{}{
								"type":    "string",
								"example": "openai",
							},
						},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"api_key"},
									"properties": map[string]interface{}{
										"api_key": map[string]interface{}{
											"type":    "string",
											"example": "sk-...",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Provider key updated",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"404": map[string]interface{}{
							"description": "Provider key not found",
						},
					},
				},
				"delete": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Delete provider key",
					"description": "Delete a provider API key for the authenticated user (BYOK)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name":        "provider",
							"in":          "path",
							"required":    true,
							"description": "Provider name",
							"schema": map[string]interface{}{
								"type":    "string",
								"example": "openai",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Provider key deleted",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"404": map[string]interface{}{
							"description": "Provider key not found",
						},
					},
				},
			},
			"/auth/provider-keys/{provider}/test": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Test provider key",
					"description": "Test a provider API key for the authenticated user (BYOK)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name":        "provider",
							"in":          "path",
							"required":    true,
							"description": "Provider name",
							"schema": map[string]interface{}{
								"type":    "string",
								"example": "openai",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Provider key valid (tested with provider API)",
						},
						"400": map[string]interface{}{
							"description": "Provider key test failed (invalid key or provider error)",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"404": map[string]interface{}{
							"description": "Provider key not found",
						},
					},
				},
			},
			"/auth/profile/password": map[string]interface{}{
				"put": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Change password",
					"description": "Change current user password",
					"security":    []map[string]interface{}{{"BearerAuth": []string{}}},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"current_password", "new_password"},
									"properties": map[string]interface{}{
										"current_password": map[string]interface{}{"type": "string", "format": "password"},
										"new_password":     map[string]interface{}{"type": "string", "format": "password"},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Password updated"},
						"400": map[string]interface{}{"description": "Invalid current password"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/auth/conversations": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []string{"Conversations"},
					"summary": "List conversations",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "List of conversations"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
				"post": map[string]interface{}{
					"tags": []string{"Conversations"},
					"summary": "Create conversation",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{"description": "Conversation created"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/auth/conversations/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []string{"Conversations"},
					"summary": "Get conversation",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"parameters": []map[string]interface{}{{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Conversation"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"404": map[string]interface{}{"description": "Not found"},
					},
				},
				"put": map[string]interface{}{
					"tags": []string{"Conversations"},
					"summary": "Update conversation",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"parameters": []map[string]interface{}{{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Updated"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"404": map[string]interface{}{"description": "Not found"},
					},
				},
				"delete": map[string]interface{}{
					"tags": []string{"Conversations"},
					"summary": "Delete conversation",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"parameters": []map[string]interface{}{{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Deleted"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"404": map[string]interface{}{"description": "Not found"},
					},
				},
			},
			"/auth/chat": map[string]interface{}{
				"post": map[string]interface{}{
					"tags": []string{"Chat"},
					"summary": "Chat (JWT)",
					"description": "Chat completion using JWT (same as /v1/chat but with Bearer JWT)",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Chat completion"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/auth/chat/stream": map[string]interface{}{
				"post": map[string]interface{}{
					"tags": []string{"Chat"},
					"summary": "Chat stream (JWT)",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "SSE stream"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/auth/chat/ws": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []string{"Chat"},
					"summary": "Chat WebSocket (JWT)",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"101": map[string]interface{}{"description": "Switching protocols"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/auth/analytics/usage": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []string{"Analytics"},
					"summary": "Get usage (JWT)",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Usage statistics"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/auth/analytics/requests": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []string{"Analytics"},
					"summary": "Get requests (JWT)",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Request history"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/auth/tunnels": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []string{"Tunnels"},
					"summary": "List tunnels",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "List of tunnels"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/auth/tunnels/stats": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []string{"Tunnels"},
					"summary": "Get tunnel stats (user)",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Tunnel statistics"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/auth/tunnels/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []string{"Tunnels"},
					"summary": "Get tunnel",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"parameters": []map[string]interface{}{{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Tunnel details"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"404": map[string]interface{}{"description": "Not found"},
					},
				},
				"put": map[string]interface{}{
					"tags": []string{"Tunnels"},
					"summary": "Set custom domain",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"parameters": []map[string]interface{}{{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Domain set"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"404": map[string]interface{}{"description": "Not found"},
					},
				},
			},
			"/auth/tunnels/{id}/disconnect": map[string]interface{}{
				"post": map[string]interface{}{
					"tags": []string{"Tunnels"},
					"summary": "Disconnect tunnel",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"parameters": []map[string]interface{}{{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Tunnel disconnected"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"404": map[string]interface{}{"description": "Not found"},
					},
				},
			},
			"/auth/tunnels/{id}/associate": map[string]interface{}{
				"post": map[string]interface{}{
					"tags": []string{"Tunnels"},
					"summary": "Associate tunnel",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"parameters": []map[string]interface{}{{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Tunnel associated"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"404": map[string]interface{}{"description": "Not found"},
					},
				},
			},
			"/auth/tunnels/{id}/domain": map[string]interface{}{
				"post": map[string]interface{}{
					"tags": []string{"Tunnels"},
					"summary": "Set custom domain (POST)",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"parameters": []map[string]interface{}{{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Domain set"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"404": map[string]interface{}{"description": "Not found"},
					},
				},
				"put": map[string]interface{}{
					"tags": []string{"Tunnels"},
					"summary": "Set custom domain (PUT)",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"parameters": []map[string]interface{}{{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Domain set"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"404": map[string]interface{}{"description": "Not found"},
					},
				},
			},
			"/auth/domains": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []string{"Domains"},
					"summary": "List custom domains",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "List of domains"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
				"post": map[string]interface{}{
					"tags": []string{"Domains"},
					"summary": "Create custom domain",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{"description": "Domain created"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/auth/domains/{id}": map[string]interface{}{
				"delete": map[string]interface{}{
					"tags": []string{"Domains"},
					"summary": "Delete domain",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"parameters": []map[string]interface{}{{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Deleted"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"404": map[string]interface{}{"description": "Not found"},
					},
				},
			},
			"/auth/domains/{id}/verify": map[string]interface{}{
				"post": map[string]interface{}{
					"tags": []string{"Domains"},
					"summary": "Verify domain",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"parameters": []map[string]interface{}{{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Verification result"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"404": map[string]interface{}{"description": "Not found"},
					},
				},
			},
			"/auth/routing/strategy": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []string{"Routing"},
					"summary": "Get user routing strategy",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Current strategy"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
				"put": map[string]interface{}{
					"tags": []string{"Routing"},
					"summary": "Set user routing strategy",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Strategy updated"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
				"delete": map[string]interface{}{
					"tags": []string{"Routing"},
					"summary": "Clear user routing strategy",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Strategy cleared"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/auth/routing/custom-rules": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []string{"Routing"},
					"summary": "Get user custom rules",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Custom routing rules"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
				"post": map[string]interface{}{
					"tags": []string{"Routing"},
					"summary": "Set user custom rules",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Rules updated"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/v1/user": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []string{"Authentication"},
					"summary": "Get user (API key)",
					"description": "Get user info for the API key holder",
					"security": []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "User info"},
						"401": map[string]interface{}{"description": "Unauthorized"},
					},
				},
			},
			"/admin/routing/strategy/lock": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "Lock routing strategy",
					"description": "Lock or unlock default routing strategy (admin only)",
					"security":    []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Lock state updated"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"403": map[string]interface{}{"description": "Forbidden - Admin required"},
					},
				},
			},
			"/admin/routing/custom-rules": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "Get custom routing rules",
					"security":    []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Custom rules"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"403": map[string]interface{}{"description": "Forbidden - Admin required"},
					},
				},
				"post": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "Set custom routing rules",
					"security":    []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Rules updated"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"403": map[string]interface{}{"description": "Forbidden - Admin required"},
					},
				},
			},
			"/admin/tunnels/stats": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "Get tunnel statistics",
					"description": "Get tunnel stats across all users (admin only)",
					"security":    []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Tunnel statistics"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"403": map[string]interface{}{"description": "Forbidden - Admin required"},
					},
				},
			},
			"/admin/users": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "List users",
					"description": "List all users (admin only)",
					"security":    []map[string]interface{}{{"BearerAuth": []string{}}},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "List of users"},
						"401": map[string]interface{}{"description": "Unauthorized"},
						"403": map[string]interface{}{"description": "Forbidden - Admin required"},
					},
				},
			},
			"/admin/routing/strategy": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "Set routing strategy",
					"description": "Set routing strategy (admin only)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Routing strategy updated",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin access required",
						},
					},
				},
				"get": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "Get routing strategy",
					"description": "Get current routing strategy (admin only)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Current routing strategy",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin access required",
						},
					},
				},
			},
			"/admin/errors": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "Get error logs",
					"description": "Get application error logs (admin only)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of error logs",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin access required",
						},
					},
				},
			},
			"/admin/errors/{id}/resolve": map[string]interface{}{
				"patch": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "Mark error as resolved",
					"description": "Mark an error log as resolved (admin only)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"description": "Error log ID",
							"schema": map[string]interface{}{
								"type":    "string",
								"example": "uuid-here",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Error marked as resolved",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin access required",
						},
					},
				},
			},
			"/admin/email/config": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "Get email config",
					"description": "Get SMTP configuration (admin only)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "SMTP configuration",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin access required",
						},
					},
				},
			},
			"/admin/email/test": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "Test email",
					"description": "Send a test email (admin only)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Test email sent",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin access required",
						},
					},
				},
			},
			"/admin/users/{id}/roles": map[string]interface{}{
				"put": map[string]interface{}{
					"tags":        []string{"Admin"},
					"summary":     "Update user roles",
					"description": "Update user roles (admin only)",
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"description": "User ID",
							"schema": map[string]interface{}{
								"type":    "string",
								"example": "uuid-here",
							},
						},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"roles"},
									"properties": map[string]interface{}{
										"roles": map[string]interface{}{
											"type": "array",
											"items": map[string]interface{}{
												"type": "string",
												"enum": []string{"user", "admin"},
											},
											"example": []string{"user", "admin"},
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "User roles updated",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin access required",
						},
					},
				},
			},
		},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"BearerAuth": map[string]interface{}{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
					"description":  "JWT token or API key",
				},
			},
		},
	}

	if !isAdmin {
		paths := spec["paths"].(map[string]interface{})
		filteredPaths := make(map[string]interface{})

		for path, pathSpec := range paths {
			if strings.HasPrefix(path, "/admin/") {
				continue
			}
			filteredPaths[path] = pathSpec
		}
		spec["paths"] = filteredPaths

		tags := spec["tags"].([]map[string]interface{})
		filteredTags := []map[string]interface{}{}
		for _, tag := range tags {
			if tag["name"] != "Admin" {
				filteredTags = append(filteredTags, tag)
			}
		}
		spec["tags"] = filteredTags
	}

	c.JSON(http.StatusOK, spec)
}
