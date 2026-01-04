package middleware

import (
	"net/http"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates middleware for API key authentication
func AuthMiddleware(apiKeyService *security.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errors.ErrUnauthorized.Error(),
			})
			c.Abort()
			return
		}

		apiKey := security.ExtractAPIKey(authHeader)
		if !apiKeyService.ValidateAPIKey(apiKey) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errors.ErrInvalidAPIKey.Error(),
			})
			c.Abort()
			return
		}

		// Store API key in context for later use
		c.Set("api_key", apiKey)
		c.Next()
	}
}
