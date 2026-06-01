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

func TestRelease(t *testing.T) {
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
			name:  "успешное освобождение",
			uuids: uuids,
			setupMock: func(repo *mocks.PartRepository, tx *mocks.TxManager) {
				tx.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(nil)

				// 1. Возвращаем деталь, у которой есть резерв (1 единица)
				parts := []model.Part{
					model.RestorePart(partID, "Тестовая деталь", "", "", 10, 1, 1, valueobject.PartProperties{}, anyTime),
				}

				repo.EXPECT().ListForUpdate(mock.Anything, mock.MatchedBy(func(f record.PartFilter) bool {
					return len(f.UUIDs) == 1 && f.UUIDs[0] == partID
				})).Return(parts, nil)

				// 2. Проверяем, что в репозиторий ушла деталь с обновленным резервом (1 -> 0)
				repo.EXPECT().UpdateReservationsBatch(mock.Anything, mock.MatchedBy(func(p []model.Part) bool {
					return len(p) == 1 && p[0].Reserved() == 0
				})).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:  "ошибка: List вернул ошибку",
			uuids: uuids,
			setupMock: func(repo *mocks.PartRepository, tx *mocks.TxManager) {
				// var err error
				tx.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errors.New("db error"))

				repo.EXPECT().ListForUpdate(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
		{
			name:  "ошибка: нечего освобождать (Release() вернул ошибку)",
			uuids: uuids,
			setupMock: func(repo *mocks.PartRepository, tx *mocks.TxManager) {
				tx.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errs.ErrNothingToRelease)

				// Возвращаем деталь, у которой резерв уже 0
				parts := []model.Part{
					model.RestorePart(partID, "Двигатель", "", "", 10, 1, 0, valueobject.PartProperties{}, anyTime),
				}
				repo.EXPECT().ListForUpdate(mock.Anything, mock.Anything).Return(parts, nil)

				// UpdateReservedBatch не должен вызваться, так как цикл прервется на ошибке
			},
			wantErr: errs.ErrNothingToRelease,
		},
		{
			name:  "ошибка: сбой при сохранении в базу",
			uuids: uuids,
			setupMock: func(repo *mocks.PartRepository, tx *mocks.TxManager) {
				tx.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errors.New("ошибка обновления"))

				parts := []model.Part{
					model.RestorePart(partID, "Двигатель", "", "", 10, 1, 1, valueobject.PartProperties{}, anyTime),
				}
				repo.EXPECT().ListForUpdate(mock.Anything, mock.Anything).Return(parts, nil)
				repo.EXPECT().UpdateReservationsBatch(mock.Anything, mock.Anything).Return(errors.New("ошибка обновления"))
			},
			wantErr: errors.New("ошибка обновления"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewPartRepository(t)
			txManager := mocks.NewTxManager(t)
			tc.setupMock(repo, txManager)

			// CompatibilityChecker здесь не используется, но нужен для конструктора
			svc := NewService(repo, domain.NewCompatibilityChecker(), txManager)
			err := svc.Release(ctx, tc.uuids)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
