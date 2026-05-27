package iam

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	"github.com/FOCCms/go-microservices-course/iam/internal/service/input"
)

const minPasswordLen = 8

func (s *Service) Register(ctx context.Context, input input.RegisterInput) (uuid.UUID, error) {
	if input.Login == "" {
		return uuid.Nil, errs.ErrInvalidLogin
	}

	if len(input.Password) < minPasswordLen {
		return uuid.Nil, errs.ErrWeakPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.bcryptCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("захешировать пароль: %w", err)
	}

	user := model.User{
		UUID:         uuid.New(),
		Login:        input.Login,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	err = s.userRepository.Create(ctx, user)
	if err != nil {
		return uuid.Nil, fmt.Errorf("зарегистрировать пользователя: %w", err)
	}

	return user.UUID, nil
}
