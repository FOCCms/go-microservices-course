package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	"github.com/FOCCms/go-microservices-course/iam/internal/service/input"
)

type IAMService interface {
	Login(ctx context.Context, input input.LoginInput) (uuid.UUID, error)
	Logout(ctx context.Context, sessionUuid string) error
	Whoami(ctx context.Context, sessionUuid string) (model.Session, model.User, error)
}
