package interceptor

import (
	"context"

	commonv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/common/v1"
)

type IAMClient interface {
	Whoami(ctx context.Context, sessionUuid string) (*commonv1.Session, *commonv1.User, error)
}
