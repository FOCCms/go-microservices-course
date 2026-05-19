package assembly

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/FOCCms/go-microservices-course/assembly/internal/model"
	"github.com/FOCCms/go-microservices-course/assembly/internal/service/assembly/mocks"
)

func TestAssemble(t *testing.T) {
	t.Parallel()

	type args struct {
		event model.OrderPaidEvent
	}

	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
		userUUID  = uuid.New()

		paidEvent = model.OrderPaidEvent{
			OrderUUID: orderUUID,
			UserUUID:  userUUID,
		}

		anyErr = errors.New("ошибка брокера")
	)

	tests := []struct {
		name        string
		args        args
		setupMock   func(producer *mocks.AssemblyProducerService)
		expectedErr error
	}{
		{
			name: "успешная сборка и отправка события в Кафку",
			args: args{event: paidEvent},
			setupMock: func(producer *mocks.AssemblyProducerService) {
				// Ждем, что метод ProduceShipAssembled вызовется 1 раз с любым контекстом и событием
				producer.EXPECT().
					ProduceShipAssembled(ctx, mock.MatchedBy(func(e model.ShipAssembledEvent) bool {
						// Проверяем, что данные из входящего события маппятся корректно
						return e.OrderUUID == paidEvent.OrderUUID && e.UserUUID == paidEvent.UserUUID
					})).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "ошибка отправки события в Кафку",
			args: args{event: paidEvent},
			setupMock: func(producer *mocks.AssemblyProducerService) {
				producer.EXPECT().
					ProduceShipAssembled(ctx, mock.Anything).
					Return(anyErr)
			},
			expectedErr: anyErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assemblyProducer := mocks.NewAssemblyProducerService(t)

			tc.setupMock(assemblyProducer)

			svc := NewService(assemblyProducer)
			err := svc.Assemble(ctx, tc.args.event)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
