package middleware

import (
	"net/http"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates middleware for API key authentication (database-backed)
func AuthMiddleware(apiKeyService *security.APIKeyServiceV2) gin.HandlerFunc {
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
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errors.ErrUnauthorized.Error(),
			})
			c.Abort()
			return
		}

		// Validate API key against database
		keyRecord, err := apiKeyService.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil || keyRecord == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errors.ErrInvalidAPIKey.Error(),
			})
			c.Abort()
			return
		}

		// Store API key info in context
		c.Set("api_key", apiKey)
		c.Set("api_key_id", keyRecord.ID.String())
		c.Set("api_key_record", keyRecord)
		c.Set("user_id", keyRecord.UserID.String())

		c.Next()
	}
}
