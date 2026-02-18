package middleware

import (
	"net/http"
	"strings"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(apiKeyService *security.APIKeyServiceV2) gin.HandlerFunc {
	return func(c *gin.Context) {
		var apiKey string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			apiKey = security.ExtractAPIKey(authHeader)
		}

		if apiKey == "" {
			isWebSocket := c.GetHeader("Upgrade") == "websocket" ||
				strings.ToLower(c.GetHeader("Connection")) == "upgrade" ||
				c.Query("token") != ""

			if isWebSocket {
				token := c.Query("token")
				if token != "" && strings.HasPrefix(token, "ur_") {
					apiKey = token
				}
			}
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errors.ErrUnauthorized.Error(),
			})
			c.Abort()
			return
		}

		keyRecord, err := apiKeyService.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil || keyRecord == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errors.ErrInvalidAPIKey.Error(),
			})
			c.Abort()
			return
		}

		c.Set("api_key", apiKey)
		c.Set("api_key_id", keyRecord.ID.String())
		c.Set("api_key_record", keyRecord)
		c.Set("user_id", keyRecord.UserID.String())

		c.Next()
	}
}
