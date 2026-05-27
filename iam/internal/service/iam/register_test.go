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
	"github.com/FOCCms/go-microservices-course/iam/internal/service/input"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	var (
		ctx          = context.Background()
		testLogin    = "new_user"
		testPassword = "secure_password_123"
		dbErr        = errors.New("ошибка БД")
	)

	tests := []struct {
		name      string
		input     input.RegisterInput
		setupMock func(uRepo *mocks.UserRepository)
		wantErr   error
	}{
		{
			name: "ошибка: пустой логин",
			input: input.RegisterInput{
				Login:    "",
				Password: testPassword,
			},
			setupMock: func(uRepo *mocks.UserRepository) {},
			wantErr:   errs.ErrInvalidLogin,
		},
		{
			name: "ошибка: слишком короткий пароль (меньше 8 символов)",
			input: input.RegisterInput{
				Login:    testLogin,
				Password: "short", // 5 символов
			},
			setupMock: func(uRepo *mocks.UserRepository) {},
			wantErr:   errs.ErrWeakPassword,
		},
		{
			name:  "ошибка: сбой репозитория при сохранении пользователя",
			input: input.RegisterInput{Login: testLogin, Password: testPassword},
			setupMock: func(uRepo *mocks.UserRepository) {
				uRepo.EXPECT().Create(ctx, mock.MatchedBy(func(u model.User) bool {
					return u.Login == testLogin && u.PasswordHash != ""
				})).Return(dbErr)
			},
			wantErr: dbErr,
		},
		{
			name:  "успешная регистрация",
			input: input.RegisterInput{Login: testLogin, Password: testPassword},
			setupMock: func(uRepo *mocks.UserRepository) {
				uRepo.EXPECT().Create(ctx, mock.MatchedBy(func(u model.User) bool {
					return u.Login == testLogin && u.PasswordHash != ""
				})).Return(nil)
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

			tc.setupMock(userRepo)

			svc := NewService(userRepo, sessionRepo, ttl, bcryptCost)
			userUUID, err := svc.Register(ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, uuid.Nil, userUUID)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, userUUID)
			}
		})
	}
}
