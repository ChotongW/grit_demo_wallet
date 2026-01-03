package requestid

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		requestID := extractRequestIDFromMetadata(ctx)

		if requestID == "" {
			requestID = Generate()
		}
		ctx = ToContext(ctx, requestID)

		return handler(ctx, req)
	}
}

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		requestID := FromContext(ctx)

		if requestID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, MetadataKey, requestID)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

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

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		requestID := extractRequestIDFromMetadata(ctx)

		if requestID == "" {
			requestID = Generate()
		}

		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ToContext(ctx, requestID),
		}

		return handler(srv, wrapped)
	}
}

func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		requestID := FromContext(ctx)

		if requestID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, MetadataKey, requestID)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
