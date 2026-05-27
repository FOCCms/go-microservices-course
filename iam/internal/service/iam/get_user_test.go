package iam

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	"github.com/FOCCms/go-microservices-course/iam/internal/service/iam/mocks"
)

func TestGetUser(t *testing.T) {
	t.Parallel()

	var (
		ctx         = context.Background()
		testUUID    = uuid.New()
		testUUIDStr = testUUID.String()

		testUser = model.User{
			UUID:         testUUID,
			Login:        "test_user",
			PasswordHash: "some_hashed_password",
			CreatedAt:    time.Now(),
		}
	)

	tests := []struct {
		name      string
		uuidStr   string
		setupMock func(repo *mocks.UserRepository)
		expected  model.User
		err       error
	}{
		{
			name:    "успешное получение пользователя",
			uuidStr: testUUIDStr,
			setupMock: func(repo *mocks.UserRepository) {
				repo.EXPECT().GetByUUID(ctx, testUUID).Return(testUser, nil)
			},
			expected: testUser,
			err:      nil,
		},
		{
			name:    "ошибка: невалидный формат UUID",
			uuidStr: "invalid-uuid-string",
			setupMock: func(repo *mocks.UserRepository) {
				// Мок не вызывается, так как метод упадет на валидации строки
			},
			expected: model.User{},
			err:      errs.ErrInvalidUUID,
		},
		{
			name:    "ошибка: пользователь не найден в репозитории",
			uuidStr: testUUIDStr,
			setupMock: func(repo *mocks.UserRepository) {
				repo.EXPECT().GetByUUID(ctx, testUUID).Return(model.User{}, errs.ErrUserNotFound)
			},
			expected: model.User{},
			err:      errs.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Инициализируем мок репозитория
			userRepo := mocks.NewUserRepository(t)
			sessionRepo := mocks.NewSessionRepository(t)
			ttl := time.Hour
			bcryptCost := 4
			tc.setupMock(userRepo)

			// Создаем сервис, прокидывая туда мок
			// Если у тебя в конструкторе NewService есть другие зависимости (например, txManager),
			// передай их как nil или создай пустые моки, если сервис к ним не обращается.
			svc := NewService(userRepo, sessionRepo, ttl, bcryptCost)

			res, err := svc.GetUser(ctx, tc.uuidStr)

			if tc.err != nil {
				require.Error(t, err)
				// Так как метод оборачивает ошибку через fmt.Errorf("...: %w", err),
				// мы обязательно используем ErrorIs для проверки по цепочке ошибок.
				assert.ErrorIs(t, err, tc.err)
				assert.Empty(t, res)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, res)
			}
		})
	}
}
