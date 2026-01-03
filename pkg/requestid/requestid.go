package requestid

import (
	"context"

	"github.com/google/uuid"
)

// Key is the context key for request ID
type ctxKey struct{}

// MetadataKey is the key used in gRPC metadata for request ID
const MetadataKey = "x-request-id"

// HeaderKey is the header key for request ID in HTTP requests
const HeaderKey = "X-Request-ID"

// Generate creates a new unique request ID
func Generate() string {
	return uuid.New().String()
}

// FromContext extracts the request ID from context
func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(ctxKey{}).(string); ok {
		return id
	}
	return ""
}

// ToContext adds the request ID to context
func ToContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, requestID)
}

// GetOrGenerate returns the request ID from context or generates a new one
func GetOrGenerate(ctx context.Context) (context.Context, string) {
	if id := FromContext(ctx); id != "" {
		return ctx, id
	}
	id := Generate()
	return ToContext(ctx, id), id
}
