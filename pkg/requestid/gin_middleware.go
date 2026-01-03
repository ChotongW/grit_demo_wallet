package requestid

import (
	"github.com/gin-gonic/gin"
)

func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(HeaderKey)

		if requestID == "" {
			requestID = Generate()
		}

		ctx := ToContext(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Header(HeaderKey, requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}
