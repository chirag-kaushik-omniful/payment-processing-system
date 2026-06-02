package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	allowed := os.Getenv("CORS_ORIGIN")
	if allowed == "" {
		allowed = "http://localhost:5173,http://127.0.0.1:5173"
	}
	origins := strings.Split(allowed, ",")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		for _, o := range origins {
			if strings.TrimSpace(o) == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		if c.GetHeader("Access-Control-Allow-Origin") == "" && len(origins) > 0 {
			c.Header("Access-Control-Allow-Origin", strings.TrimSpace(origins[0]))
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Correlation-ID")
		c.Header("Access-Control-Expose-Headers", "X-Correlation-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
