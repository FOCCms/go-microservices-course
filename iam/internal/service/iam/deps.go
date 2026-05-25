package iam

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/iam/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user model.User) error
	GetByLogin(ctx context.Context, login string) (model.User, error)
	GetByUUID(ctx context.Context, uuid uuid.UUID) (model.User, error)
}

type SessionRepository interface {
	Get(ctx context.Context, uuid string) (model.Session, error)
	Set(ctx context.Context, session model.Session, ttl time.Duration) error
	Delete(ctx context.Context, uuid string) error
}
