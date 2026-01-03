package requestid

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct{}

const MetadataKey = "x-request-id"
const HeaderKey = "X-Request-ID"

func Generate() string {
	return uuid.New().String()
}

func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(ctxKey{}).(string); ok {
		return id
	}
	return ""
}

func ToContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, requestID)
}
func GetOrGenerate(ctx context.Context) (context.Context, string) {
	if id := FromContext(ctx); id != "" {
		return ctx, id
	}
	id := Generate()
	return ToContext(ctx, id), id
}
