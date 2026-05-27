package iam

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/model"
)

func (s *Service) GetUser(ctx context.Context, uuidStr string) (model.User, error) {
	uuid, err := uuid.Parse(uuidStr)
	if err != nil {
		return model.User{}, fmt.Errorf("получить пользователя: %w", errs.ErrInvalidUUID)
	}

	user, err := s.userRepository.GetByUUID(ctx, uuid)
	if err != nil {
		return model.User{}, fmt.Errorf("получить пользователя: %w", err)
	}

	return user, nil
}
