package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	"github.com/FOCCms/go-microservices-course/iam/internal/service/input"
)

type IAMService interface {
	Register(ctx context.Context, input input.RegisterInput) (uuid.UUID, error)
	GetUser(ctx context.Context, uuidStr string) (model.User, error)
}
