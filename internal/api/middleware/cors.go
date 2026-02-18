package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(customOrigins []string) gin.HandlerFunc {
	allowOrigins := []string{
		"http://localhost:3000",
		"http://localhost:3002", // Frontend dev server
		"http://localhost:5173", // Vite default port
		"http://127.0.0.1:3000",
		"http://127.0.0.1:3002",
		"http://127.0.0.1:5173",
		"https://uniroute.co",
		"https://www.uniroute.co",
		"https://app.uniroute.co",
	}

	if len(customOrigins) > 0 {
		allowOrigins = customOrigins
	}
	
	return cors.New(cors.Config{
		AllowOrigins: allowOrigins,
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
