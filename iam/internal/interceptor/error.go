package interceptor

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
)

func UnaryErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("перехвачена паника",
				"panic", r,
				"method", info.FullMethod,
				"stack", string(debug.Stack()))
			err = status.Error(codes.Internal, "внутренняя ошибка")
		}
	}()

	resp, err = handler(ctx, req)
	if err != nil {
		slog.Error("ошибка в методе", "method", info.FullMethod, "err", err)
		return nil, mapToGRPCError(err)
	}
	return resp, nil
}

func mapToGRPCError(err error) error {
	if nil == err {
		return nil
	}

	if _, ok := status.FromError(err); ok {
		return err
	}

	switch {
	case errors.Is(err, errs.ErrInvalidLogin),
		errors.Is(err, errs.ErrWeakPassword),
		errors.Is(err, errs.ErrEmptyCredential),
		errors.Is(err, errs.ErrEmptySessionID),
		errors.Is(err, errs.ErrInvalidUUID):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, errs.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, errs.ErrInvalidCredentials),
		errors.Is(err, errs.ErrSessionNotFound):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, errs.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "внутренняя ошибка")
	}
}
