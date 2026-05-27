package iam

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	"github.com/FOCCms/go-microservices-course/iam/internal/service/input"
)

func (s *Service) Login(ctx context.Context, input input.LoginInput) (uuid.UUID, error) {
	if input.Login == "" || input.Password == "" {
		return uuid.Nil, errs.ErrEmptyCredential
	}

	user, err := s.userRepository.GetByLogin(ctx, input.Login)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return uuid.Nil, fmt.Errorf("залогинить пользователя: %w", errs.ErrInvalidCredentials)
		}
		return uuid.Nil, fmt.Errorf("залогинить пользователя: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return uuid.Nil, fmt.Errorf("залогинить пользователя: %w", errs.ErrInvalidCredentials)
	}

	session := model.Session{
		UUID:      uuid.New(),
		UserUUID:  user.UUID,
		Login:     user.Login,
		CreatedAt: time.Now(),
		UpdatedAt: nil,
		ExpiresAt: time.Now().Add(s.sessionTTL),
	}

	err = s.sessionRepository.Set(ctx, session, s.sessionTTL)
	if err != nil {
		return uuid.Nil, fmt.Errorf("залогинить пользователя: %w", err)
	}

	return session.UUID, nil
}
