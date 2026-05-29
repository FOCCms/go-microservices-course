package iam

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
)

func (s *Service) Logout(ctx context.Context, sessionUuid string) error {
	if sessionUuid == "" {
		return fmt.Errorf("разлогинить пользователя: %w", errs.ErrEmptySessionID)
	}
	if err := uuid.Validate(sessionUuid); err != nil {
		return fmt.Errorf("разлогинить пользователя: %w", errs.ErrInvalidUUID)
	}

	err := s.sessionRepository.Delete(ctx, sessionUuid)
	if err != nil {
		return fmt.Errorf("разлогинить пользователя: %w", err)
	}

	slog.Info("сессия пользователя успешно завершена (logout)",
		slog.String("session_uuid", sessionUuid),
	)

	return nil
}
