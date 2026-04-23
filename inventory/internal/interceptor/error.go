package interceptor

import (
	"context"
	"log/slog"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("перехвачена паника",
				"panic", r,
				"method", info.FullMethod,
				"stack", string(debug.Stack()))
		}
	}()

	resp, err := handler(ctx, req)
	if err != nil {
		slog.Error("ошибка в методе", "method", info.FullMethod, "err", err)
		return nil, status.Error(codes.Internal, "внутренняя ошибка inventory сервиса")
	}
	return resp, err
}
