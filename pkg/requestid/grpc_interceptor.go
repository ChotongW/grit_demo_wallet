package requestid

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor returns a gRPC server interceptor that extracts
// the request ID from incoming metadata and adds it to the context.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract request ID from metadata
		requestID := extractRequestIDFromMetadata(ctx)

		// Generate new one if not present
		if requestID == "" {
			requestID = Generate()
		}

		// Add to context
		ctx = ToContext(ctx, requestID)

		return handler(ctx, req)
	}
}

// UnaryClientInterceptor returns a gRPC client interceptor that propagates
// the request ID from context to outgoing gRPC metadata.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Get request ID from context
		requestID := FromContext(ctx)

		// If we have a request ID, add it to outgoing metadata
		if requestID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, MetadataKey, requestID)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// extractRequestIDFromMetadata extracts the request ID from gRPC incoming metadata
func extractRequestIDFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get(MetadataKey)
	if len(values) > 0 {
		return values[0]
	}

	return ""
}

// StreamServerInterceptor returns a gRPC server interceptor for streaming RPCs
// that extracts the request ID from incoming metadata and adds it to the context.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		// Extract request ID from metadata
		requestID := extractRequestIDFromMetadata(ctx)

		// Generate new one if not present
		if requestID == "" {
			requestID = Generate()
		}

		// Create wrapped stream with new context
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ToContext(ctx, requestID),
		}

		return handler(srv, wrapped)
	}
}

// StreamClientInterceptor returns a gRPC client interceptor for streaming RPCs
// that propagates the request ID from context to outgoing gRPC metadata.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		// Get request ID from context
		requestID := FromContext(ctx)

		// If we have a request ID, add it to outgoing metadata
		if requestID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, MetadataKey, requestID)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}

// wrappedServerStream wraps a grpc.ServerStream with a custom context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
