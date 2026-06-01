package part

import (
	"context"
	"errors"
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

func TestReserve(t *testing.T) {
	t.Parallel()

	var (
		ctx     = context.Background()
		partID  = uuid.New().String()
		anyTime = time.Now()
		uuids   = []string{partID}
	)

	tests := []struct {
		name      string
		uuids     []string
		setupMock func(repo *mocks.PartRepository, tx *mocks.TxManager)
		wantErr   error
	}{
		{
			name:  "успешное резервирование",
			uuids: uuids,
			setupMock: func(repo *mocks.PartRepository, tx *mocks.TxManager) {
				tx.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(nil)

				// 1. Возвращаем деталь, у которой есть остаток (10) и текущий резерв (0)
				parts := []model.Part{
					model.RestorePart(partID, "Корпус", "", "", 10, 10, 0, valueobject.PartProperties{}, anyTime),
				}

				repo.EXPECT().ListForUpdate(mock.Anything, mock.MatchedBy(func(f record.PartFilter) bool {
					return len(f.UUIDs) == 1 && f.UUIDs[0] == partID
				})).Return(parts, nil)

				// 2. Проверяем, что в репозиторий ушла деталь с инкрементированным резервом (0 -> 1)
				repo.EXPECT().UpdateReservationsBatch(mock.Anything, mock.MatchedBy(func(p []model.Part) bool {
					return len(p) == 1 && p[0].Reserved() == 1
				})).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:  "ошибка: List вернул ошибку",
			uuids: uuids,
			setupMock: func(repo *mocks.PartRepository, tx *mocks.TxManager) {
				tx.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errors.New("db connection lost"))

				repo.EXPECT().ListForUpdate(mock.Anything, mock.Anything).Return(nil, errors.New("db connection lost"))
			},
			wantErr: errors.New("db connection lost"),
		},
		{
			name:  "ошибка: недостаточно товара на складе (ErrOutOfStock)",
			uuids: uuids,
			setupMock: func(repo *mocks.PartRepository, tx *mocks.TxManager) {
				tx.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errs.ErrOutOfStock)

				// Возвращаем деталь, где резерв равен общему количеству (свободных нет)
				parts := []model.Part{
					model.RestorePart(partID, "Лазер", "", "", 10, 0, 0, valueobject.PartProperties{}, anyTime),
				}
				repo.EXPECT().ListForUpdate(mock.Anything, mock.Anything).Return(parts, nil)

				// UpdateReservedBatch не вызывается
			},
			wantErr: errs.ErrOutOfStock,
		},
		{
			name:  "ошибка: сбой при пакетном обновлении",
			uuids: uuids,
			setupMock: func(repo *mocks.PartRepository, tx *mocks.TxManager) {
				tx.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errors.New("db write error"))

				parts := []model.Part{
					model.RestorePart(partID, "Двигатель", "", "", 10, 10, 0, valueobject.PartProperties{}, anyTime),
				}
				repo.EXPECT().ListForUpdate(mock.Anything, mock.Anything).Return(parts, nil)
				repo.EXPECT().UpdateReservationsBatch(mock.Anything, mock.Anything).Return(errors.New("db write error"))
			},
			wantErr: errors.New("db write error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewPartRepository(t)
			txManager := mocks.NewTxManager(t)
			tc.setupMock(repo, txManager)

			svc := NewService(repo, domain.NewCompatibilityChecker(), txManager)
			err := svc.Reserve(ctx, tc.uuids)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
