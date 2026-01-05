package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware creates middleware for CORS configuration
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173", // Vite default port
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
			"https://uniroute.dev",
			"https://www.uniroute.dev",
			"https://app.uniroute.dev",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"HEAD",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Accept",
			"X-API-Key",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	})
}
