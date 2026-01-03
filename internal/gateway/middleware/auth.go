package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasSuffix(c.Request.URL.Path, "/health") {
			c.Next()
			return
		}

		clientAPIKey := c.GetHeader("X-API-KEY")
		if clientAPIKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API Key is missing"})
			c.Abort()
			return
		}

		if clientAPIKey != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
			c.Abort()
			return
		}

		c.Next()
	}
}
