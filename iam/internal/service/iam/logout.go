package iam

import (
	"context"
	"fmt"

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

	return nil
}
