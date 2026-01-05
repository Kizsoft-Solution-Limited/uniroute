package middleware

import (
	"net/http"

	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/errors"
	"github.com/gin-gonic/gin"
)

// contains checks if a string slice contains a specific value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// AdminMiddleware ensures only users with admin role can access the route
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user roles from context (set by JWT middleware)
		roles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errors.ErrUnauthorized.Error(),
			})
			c.Abort()
			return
		}

		// Check if user has admin role
		rolesSlice, ok := roles.([]string)
		if !ok || !contains(rolesSlice, "admin") {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
