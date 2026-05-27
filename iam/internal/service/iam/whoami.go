package iam

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/model"
)

func (s *Service) Whoami(ctx context.Context, sessionUuid string) (model.Session, model.User, error) {
	if sessionUuid == "" {
		return model.Session{}, model.User{}, fmt.Errorf("получить сессию и пользователя: %w", errs.ErrEmptySessionID)
	}
	if err := uuid.Validate(sessionUuid); err != nil {
		return model.Session{}, model.User{}, fmt.Errorf("получить сессию и пользователя: %w", errs.ErrInvalidUUID)
	}

	session, err := s.sessionRepository.Get(ctx, sessionUuid)
	if err != nil {
		return model.Session{}, model.User{}, fmt.Errorf("получить сессию и пользователя: %w", err)
	}

	user, err := s.userRepository.GetByUUID(ctx, session.UserUUID)
	if err != nil {
		return model.Session{}, model.User{}, fmt.Errorf("получить сессию и пользователя: %w", err)
	}

	return session, user, nil
}
