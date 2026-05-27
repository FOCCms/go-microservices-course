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
	"golang.org/x/crypto/bcrypt"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	"github.com/FOCCms/go-microservices-course/iam/internal/service/iam/mocks"
	"github.com/FOCCms/go-microservices-course/iam/internal/service/input"
)

func TestLogin(t *testing.T) {
	t.Parallel()

	var (
		ctx          = context.Background()
		testLogin    = "test_user"
		testPassword = "secret_password"
		testUserUUID = uuid.New()

		correctHash, _ = bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)

		testUser = model.User{
			UUID:         testUserUUID,
			Login:        testLogin,
			PasswordHash: string(correctHash),
			CreatedAt:    time.Now(),
		}

		dbErr = errors.New("ошибка БД")
	)

	tests := []struct {
		name      string
		input     input.LoginInput
		setupMock func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository)
		wantErr   error
	}{
		{
			name: "ошибка: пустой логин",
			input: input.LoginInput{
				Login:    "",
				Password: testPassword,
			},
			setupMock: func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository) {},
			wantErr:   errs.ErrEmptyCredential,
		},
		{
			name: "ошибка: пустой пароль",
			input: input.LoginInput{
				Login:    testLogin,
				Password: "",
			},
			setupMock: func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository) {},
			wantErr:   errs.ErrEmptyCredential,
		},
		{
			name:  "ошибка: пользователь не найден (маппинг в ErrInvalidCredentials)",
			input: input.LoginInput{Login: testLogin, Password: testPassword},
			setupMock: func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository) {
				uRepo.EXPECT().GetByLogin(ctx, testLogin).Return(model.User{}, errs.ErrUserNotFound)
			},
			wantErr: errs.ErrInvalidCredentials,
		},
		{
			name:  "ошибка: сбой базы данных при поиске пользователя",
			input: input.LoginInput{Login: testLogin, Password: testPassword},
			setupMock: func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository) {
				uRepo.EXPECT().GetByLogin(ctx, testLogin).Return(model.User{}, dbErr)
			},
			wantErr: dbErr,
		},
		{
			name:  "ошибка: неверный пароль",
			input: input.LoginInput{Login: testLogin, Password: "wrong_password"},
			setupMock: func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository) {
				// Передаем юзера с хэшем от правильного пароля, но в input у нас wrong_password
				uRepo.EXPECT().GetByLogin(ctx, testLogin).Return(testUser, nil)
			},
			wantErr: errs.ErrInvalidCredentials,
		},
		{
			name:  "ошибка: не удалось сохранить сессию",
			input: input.LoginInput{Login: testLogin, Password: testPassword},
			setupMock: func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository) {
				uRepo.EXPECT().GetByLogin(ctx, testLogin).Return(testUser, nil)

				sRepo.EXPECT().Set(ctx, mock.Anything, time.Hour).Return(dbErr)
			},
			wantErr: dbErr,
		},
		{
			name:  "успешный логин",
			input: input.LoginInput{Login: testLogin, Password: testPassword},
			setupMock: func(uRepo *mocks.UserRepository, sRepo *mocks.SessionRepository) {
				uRepo.EXPECT().GetByLogin(ctx, testLogin).Return(testUser, nil)
				sRepo.EXPECT().Set(ctx, mock.Anything, time.Hour).Return(nil)
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

			tc.setupMock(userRepo, sessionRepo)

			svc := NewService(userRepo, sessionRepo, ttl, bcryptCost)
			sessionUUID, err := svc.Login(ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, uuid.Nil, sessionUUID)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, sessionUUID)
			}
		})
	}
}
