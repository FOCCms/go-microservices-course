package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/FOCCms/go-microservices-course/platform/pkg/auth"
)

func AuthOutgoingInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if uuid, ok := auth.SessionUUIDFromContext(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, "session-uuid", uuid)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
