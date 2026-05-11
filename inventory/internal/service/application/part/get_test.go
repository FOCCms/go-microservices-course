package part

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
	"github.com/FOCCms/go-microservices-course/inventory/internal/service/application/part/mocks"
	"github.com/FOCCms/go-microservices-course/inventory/internal/service/domain"
)

func TestGet(t *testing.T) {
	t.Parallel()

	var (
		ctx    = context.Background()
		partID = uuid.New().String()
		part   = model.RestorePart(partID, "Тестовый образец", "", "", 0, 0, 0, valueobject.PartProperties{}, time.Now())
	)

	tests := []struct {
		name      string
		idStr     string
		setupMock func(repo *mocks.PartRepository)
		expected  model.Part
		err       error
	}{
		{
			name:  "успешное получение",
			idStr: partID,
			setupMock: func(repo *mocks.PartRepository) {
				repo.EXPECT().Get(ctx, partID).Return(part, nil)
			},
			expected: part,
			err:      nil,
		},
		{
			name:  "ошибка: неверный UUID формат",
			idStr: "невалидный uuid",
			setupMock: func(repo *mocks.PartRepository) {
			},
			expected: model.Part{},
			err:      errs.ErrInvalidUUID,
		},
		{
			name:  "ошибка: деталь не найдена в БД",
			idStr: partID,
			setupMock: func(repo *mocks.PartRepository) {
				repo.EXPECT().Get(ctx, partID).Return(model.Part{}, errs.ErrPartNotFound)
			},
			expected: model.Part{},
			err:      errs.ErrPartNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewPartRepository(t)
			checker := domain.NewCompatibilityChecker()
			tc.setupMock(repo)

			svc := NewService(repo, checker)
			res, err := svc.Get(ctx, tc.idStr)

			if tc.err != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.err)
				assert.Empty(t, res)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, res)
			}
		})
	}
}
