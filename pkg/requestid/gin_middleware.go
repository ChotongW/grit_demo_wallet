package requestid

import (
	"github.com/gin-gonic/gin"
)

// GinMiddleware returns a Gin middleware that extracts or generates a request ID
// and adds it to the context and response headers.
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get request ID from incoming header
		requestID := c.GetHeader(HeaderKey)

		// Generate a new one if not provided
		if requestID == "" {
			requestID = Generate()
		}

		// Add to context
		ctx := ToContext(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// Set response header
		c.Header(HeaderKey, requestID)

		// Store in Gin context for easy access
		c.Set("request_id", requestID)

		c.Next()
	}
}
