package iam

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	"github.com/FOCCms/go-microservices-course/iam/internal/service/iam/mocks"
)

func TestWhoami(t *testing.T) {
	t.Parallel()

	var (
		ctx            = context.Background()
		validSessionID = uuid.New().String()
		testUserUUID   = uuid.New()
		dbErr          = errors.New("ошибка БД")

		testSession = model.Session{
			UUID:      uuid.MustParse(validSessionID),
			UserUUID:  testUserUUID,
			Login:     "test_user",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Hour),
		}

		testUser = model.User{
			UUID:         testUserUUID,
			Login:        "test_user",
			PasswordHash: "hash",
			CreatedAt:    time.Now(),
		}
	)

	tests := []struct {
		name         string
		sessionID    string
		setupMock    func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository)
		expectedSess model.Session
		expectedUser model.User
		wantErr      error
	}{
		{
			name:         "ошибка: пустой ID сессии",
			sessionID:    "",
			setupMock:    func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository) {},
			expectedSess: model.Session{},
			expectedUser: model.User{},
			wantErr:      errs.ErrEmptySessionID,
		},
		{
			name:         "ошибка: невалидный формат UUID сессии",
			sessionID:    "not-a-valid-uuid",
			setupMock:    func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository) {},
			expectedSess: model.Session{},
			expectedUser: model.User{},
			wantErr:      errs.ErrInvalidUUID,
		},
		{
			name:      "ошибка: сбой при получении сессии",
			sessionID: validSessionID,
			setupMock: func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository) {
				sRepo.EXPECT().Get(mock.Anything, validSessionID).Return(model.Session{}, dbErr)
			},
			expectedSess: model.Session{},
			expectedUser: model.User{},
			wantErr:      dbErr,
		},
		{
			name:      "ошибка: сессия найдена, но сбой при получении пользователя",
			sessionID: validSessionID,
			setupMock: func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository) {
				sRepo.EXPECT().Get(mock.Anything, validSessionID).Return(testSession, nil)
				uRepo.EXPECT().GetByUUID(mock.Anything, testUserUUID).Return(model.User{}, dbErr)
			},
			expectedSess: model.Session{},
			expectedUser: model.User{},
			wantErr:      dbErr,
		},
		{
			name:      "успешное получение сессии и пользователя",
			sessionID: validSessionID,
			setupMock: func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository) {
				sRepo.EXPECT().Get(mock.Anything, validSessionID).Return(testSession, nil)
				uRepo.EXPECT().GetByUUID(mock.Anything, testUserUUID).Return(testUser, nil)
			},
			expectedSess: testSession,
			expectedUser: testUser,
			wantErr:      nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userRepo := mocks.NewUserRepository(t)
			sessionRepo := mocks.NewSessionRepository(t)
			ttl := time.Hour
			bcryptCost := 4

			tc.setupMock(userRepo, sessionRepo)

			svc := NewService(userRepo, sessionRepo, ttl, bcryptCost)
			resSess, resUser, err := svc.Whoami(ctx, tc.sessionID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, resSess)
				assert.Empty(t, resUser)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedSess, resSess)
				assert.Equal(t, tc.expectedUser, resUser)
			}
		})
	}
}
