package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleBindingError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "Invalid request body",
		"details": err.Error(),
	})
}

func HandleServiceError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})

	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})

	case codes.AlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": "Resource already exists"})

	case codes.PermissionDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})

	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})

	case codes.FailedPrecondition:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Precondition failed"})

	case codes.Unavailable:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service temporarily unavailable, please try again later"})

	case codes.DeadlineExceeded:
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Request timed out"})

	case codes.Internal, codes.Unknown, codes.DataLoss:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}
}
