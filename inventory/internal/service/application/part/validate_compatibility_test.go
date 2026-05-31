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
	"github.com/FOCCms/go-microservices-course/inventory/internal/service/application/part/mocks"
)

func TestValidateCompatibility(t *testing.T) {
	t.Parallel()

	var (
		ctx      = context.Background()
		hullID   = uuid.New().String()
		engineID = uuid.New().String()
		anyTime  = time.Now()

		slots = model.ShipSlots{
			Hull:   hullID,
			Engine: engineID,
		}
	)

	tests := []struct {
		name      string
		slots     model.ShipSlots
		setupMock func(repo *mocks.PartRepository, checker *mocks.CompatibilityChecker)
		wantErr   error
	}{
		{
			name:  "успешная проверка совместимости",
			slots: slots,
			setupMock: func(repo *mocks.PartRepository, checker *mocks.CompatibilityChecker) {
				parts := []model.Part{
					model.RestorePart(hullID, "Корпус", "", valueobject.PartTypeHull, 10, 10, 0, valueobject.PartProperties{}, anyTime),
					model.RestorePart(engineID, "Двигатель", "", valueobject.PartTypeEngine, 10, 10, 0, valueobject.PartProperties{}, anyTime),
				}

				repo.EXPECT().List(mock.Anything, mock.Anything).Return(parts, nil)
				checker.EXPECT().Check(mock.Anything).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:      "ошибка: дублирующиеся UUID",
			slots:     model.ShipSlots{Hull: hullID, Engine: hullID}, // Дважды один ID
			setupMock: func(repo *mocks.PartRepository, checker *mocks.CompatibilityChecker) {},
			wantErr:   errs.ErrPartIsNonUnique,
		},
		{
			name:  "ошибка: несоответствие типа слота (двигатель вместо корпуса)",
			slots: slots,
			setupMock: func(repo *mocks.PartRepository, checker *mocks.CompatibilityChecker) {
				parts := []model.Part{
					model.RestorePart(hullID, "Не корпус", "", valueobject.PartTypeEngine, 10, 10, 0, valueobject.PartProperties{}, anyTime),
					model.RestorePart(engineID, "Двигатель", "", valueobject.PartTypeEngine, 10, 10, 0, valueobject.PartProperties{}, anyTime),
				}

				repo.EXPECT().List(mock.Anything, mock.Anything).Return(parts, nil)
			},
			wantErr: errs.ErrPartTypeMismatch,
		},
		{
			name:  "ошибка: доменный чекер нашел несовместимость",
			slots: slots,
			setupMock: func(repo *mocks.PartRepository, checker *mocks.CompatibilityChecker) {
				parts := []model.Part{
					model.RestorePart(hullID, "Корпус", "", valueobject.PartTypeHull, 10, 10, 0, valueobject.PartProperties{}, anyTime),
					model.RestorePart(engineID, "Двигатель", "", valueobject.PartTypeEngine, 10, 10, 0, valueobject.PartProperties{}, anyTime),
				}

				repo.EXPECT().List(mock.Anything, mock.Anything).Return(parts, nil)
				checker.EXPECT().Check(mock.Anything).Return(errs.ErrIncompatibleParts)
			},
			wantErr: errs.ErrIncompatibleParts,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewPartRepository(t)
			checker := mocks.NewCompatibilityChecker(t)
			txManager := mocks.NewTxManager(t)
			tc.setupMock(repo, checker)

			svc := NewService(repo, checker, txManager)
			err := svc.ValidateCompatibility(ctx, tc.slots)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
