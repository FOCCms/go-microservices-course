package iam

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
			slog.Warn("неудачная попытка входа: пользователь не найден",
				slog.String("login", input.Login),
			)
			return uuid.Nil, fmt.Errorf("залогинить пользователя: %w", errs.ErrInvalidCredentials)
		}
		return uuid.Nil, fmt.Errorf("залогинить пользователя: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		slog.Warn("неудачная попытка входа: неверный пароль",
			slog.String("user_uuid", user.UUID.String()),
			slog.String("login", user.Login),
		)
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

	slog.Info("пользователь успешно залогинен",
		slog.String("user_uuid", user.UUID.String()),
		slog.String("login", user.Login),
		slog.String("session_uuid", session.UUID.String()),
	)

	return session.UUID, nil
}
