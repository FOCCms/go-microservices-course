package iam

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/service/iam/mocks"
)

func TestLogout(t *testing.T) {
	t.Parallel()

	var (
		ctx            = context.Background()
		validSessionID = uuid.New().String()
		dbErr          = errors.New("ошибка БД")
	)

	tests := []struct {
		name      string
		sessionID string
		setupMock func(sRepo *mocks.SessionRepository)
		wantErr   error
	}{
		{
			name:      "ошибка: пустой ID сессии",
			sessionID: "",
			setupMock: func(sRepo *mocks.SessionRepository) {},
			wantErr:   errs.ErrEmptySessionID,
		},
		{
			name:      "ошибка: невалидный формат UUID сессии",
			sessionID: "not-a-valid-uuid",
			setupMock: func(sRepo *mocks.SessionRepository) {},
			wantErr:   errs.ErrInvalidUUID,
		},
		{
			name:      "ошибка: сбой базы данных/redis при удалении сессии",
			sessionID: validSessionID,
			setupMock: func(sRepo *mocks.SessionRepository) {
				sRepo.EXPECT().Delete(ctx, validSessionID).Return(dbErr)
			},
			wantErr: dbErr,
		},
		{
			name:      "успешный логаут",
			sessionID: validSessionID,
			setupMock: func(sRepo *mocks.SessionRepository) {
				sRepo.EXPECT().Delete(ctx, validSessionID).Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userRepo := mocks.NewUserRepository(t)
			sessionRepo := mocks.NewSessionRepository(t)
			ttl := time.Hour
			bcryptCost := 4

			tc.setupMock(sessionRepo)

			svc := NewService(userRepo, sessionRepo, ttl, bcryptCost)
			err := svc.Logout(ctx, tc.sessionID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
