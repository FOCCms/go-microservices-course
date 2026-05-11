package part

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	model "github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
	"github.com/FOCCms/go-microservices-course/inventory/internal/service/application/part/mocks"
	"github.com/FOCCms/go-microservices-course/inventory/internal/service/domain"
)

func TestList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	partID := uuid.New().String()
	secondPartID := uuid.New().String()
	anyTime := time.Now()

	parts := []model.Part{
		model.RestorePart(secondPartID, "Тестовый корпус", "", "", 0, 0, 0, valueobject.PartProperties{}, anyTime),
		model.RestorePart(partID, "Тестовый двигатель", "", "", 0, 0, 0, valueobject.PartProperties{}, anyTime),
	}
	sortedParts := []model.Part{
		model.RestorePart(partID, "Тестовый двигатель", "", "", 0, 0, 0, valueobject.PartProperties{}, anyTime),
		model.RestorePart(secondPartID, "Тестовый корпус", "", "", 0, 0, 0, valueobject.PartProperties{}, anyTime),
	}

	tests := []struct {
		name      string
		filter    valueobject.PartFilter
		setupMock func(repo *mocks.PartRepository)
		want      []model.Part
		wantErr   error
	}{
		{
			name:   "успешный список по UUID",
			filter: valueobject.PartFilter{UUIDs: []string{partID}},
			setupMock: func(repo *mocks.PartRepository) {
				repo.EXPECT().List(ctx, mock.MatchedBy(func(f record.PartFilter) bool {
					return len(f.UUIDs) == 1 && f.UUIDs[0] == partID
				})).Return(parts, nil)
			},
			want:    parts,
			wantErr: nil,
		},
		{
			name:      "ошибка: неверный UUID в фильтре",
			filter:    valueobject.PartFilter{UUIDs: []string{"невалидный uuid"}},
			setupMock: func(repo *mocks.PartRepository) {},
			want:      []model.Part{},
			wantErr:   errs.ErrInvalidUUID,
		},
		{
			name:   "успешная сортировка при пустых UUID",
			filter: valueobject.PartFilter{UUIDs: []string{}, PartType: valueobject.PartTypeUnspecified},
			setupMock: func(repo *mocks.PartRepository) {
				repo.EXPECT().List(ctx, mock.Anything).Return(sortedParts, nil)
			},
			want: []model.Part{
				model.RestorePart(partID, "Тестовый двигатель", "", "", 0, 0, 0, valueobject.PartProperties{}, anyTime),
				model.RestorePart(secondPartID, "Тестовый корпус", "", "", 0, 0, 0, valueobject.PartProperties{}, anyTime),
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewPartRepository(t)

			tc.setupMock(repo)

			svc := NewService(repo, domain.NewCompatibilityChecker())
			res, err := svc.List(ctx, tc.filter)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, res)
			}
		})
	}
}
