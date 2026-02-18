package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/Kizsoft-Solution-Limited/uniroute/internal/security"
)

func JWTAuthMiddleware(jwtService *security.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			c.Abort()
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format, expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token is empty",
			})
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			if err == security.ErrExpiredToken {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "token expired",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "invalid token",
				})
			}
		c.Abort()
		return
	}

	c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_roles", claims.Roles)

		c.Next()
	}
}

