package interceptor

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/FOCCms/go-microservices-course/platform/pkg/auth"
)

var publicMethods = map[string]bool{
	"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo":      true,
	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo": true,
}

const SessionMetadataKey = "session-uuid"

func AuthIncomingInterceptor(iamClient IAMClient) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Пропускаем аутентификацию для публичных методов
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "отсутствует metadata")
		}

		sessionUuids := md.Get(SessionMetadataKey)
		if len(sessionUuids) == 0 || sessionUuids[0] == "" {
			return nil, status.Error(codes.Unauthenticated, "отсутствует session-uuid")
		}
		sessionUuid := sessionUuids[0]

		_, user, err := iamClient.Whoami(ctx, sessionUuid)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "недействительная сессия")
		}

		ctx = auth.WithUserUUID(ctx, uuid.MustParse(user.Uuid))

		return handler(ctx, req)
	}
}
