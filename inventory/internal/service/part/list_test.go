package part

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
	"github.com/FOCCms/go-microservices-course/inventory/internal/service/part/mocks"
)

func TestList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	partID := uuid.New()
	parts := []model.Part{
		{UUID: uuid.New().String(), Name: "Тестовый корпус"},
		{UUID: partID.String(), Name: "Тестовый двигатель"},
	}

	tests := []struct {
		name      string
		filter    model.PartFilter
		setupMock func(repo *mocks.PartRepository)
		want      []model.Part
		wantErr   error
	}{
		{
			name:   "успешный список по UUID",
			filter: model.PartFilter{UUIDs: []string{partID.String()}},
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
			filter:    model.PartFilter{UUIDs: []string{"невалидный uuid"}},
			setupMock: func(repo *mocks.PartRepository) {},
			want:      []model.Part{},
			wantErr:   errs.ErrInvalidUUID,
		},
		{
			name:   "успешная сортировка при пустых UUID",
			filter: model.PartFilter{UUIDs: []string{}, PartType: model.PartTypeUnspecified},
			setupMock: func(repo *mocks.PartRepository) {
				repo.EXPECT().List(ctx, mock.Anything).Return(parts, nil)
			},
			want: []model.Part{
				{UUID: parts[1].UUID, Name: "Тестовый двигатель"},
				{UUID: parts[0].UUID, Name: "Тестовый корпус"},
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewPartRepository(t)
			tc.setupMock(repo)

			svc := NewService(repo)
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
