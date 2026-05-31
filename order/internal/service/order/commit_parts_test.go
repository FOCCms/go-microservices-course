package order

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
	"github.com/FOCCms/go-microservices-course/order/internal/service/order/mocks"
)

func TestCommitParts(t *testing.T) {
	t.Parallel()

	type args struct {
		event model.ShipAssembledEvent
	}

	var (
		ctx     = context.Background()
		orderID = gofakeit.UUID()
		event   = model.ShipAssembledEvent{
			OrderUUID: uuid.MustParse(orderID),
		}
	)

	tests := []struct {
		name        string
		args        args
		setupMock   func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService)
		expectedErr error
	}{
		{
			name: "успешное списание деталей (заказ переходит в ASSEMBLED)",
			args: args{event: event},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(nil)

				repo.EXPECT().
					GetForUpdate(mock.Anything, orderID).
					Return(model.Order{UUID: uuid.MustParse(orderID), Status: model.OrderStatusPaid}, nil)

				client.EXPECT().
					CommitParts(mock.Anything, mock.Anything).
					Return(nil)

				repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(o model.Order) bool {
						return o.Status == model.OrderStatusAssembled
					})).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "идемпотентность: заказ уже ASSEMBLED, ничего не делаем",
			args: args{event: event},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(nil)

				repo.EXPECT().
					GetForUpdate(mock.Anything, orderID).
					Return(model.Order{UUID: uuid.MustParse(orderID), Status: model.OrderStatusAssembled}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "ошибка: заказ отменен (конфликт статуса)",
			args: args{event: event},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errs.ErrOrderCancelled)

				repo.EXPECT().
					GetForUpdate(mock.Anything, orderID).
					Return(model.Order{UUID: uuid.MustParse(orderID), Status: model.OrderStatusCancelled}, nil)
			},
			expectedErr: errs.ErrOrderCancelled,
		},
		{
			name: "ошибка: заказ не найден в репозитории",
			args: args{event: event},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errs.ErrOrderNotFound)

				repo.EXPECT().
					GetForUpdate(mock.Anything, orderID).
					Return(model.Order{}, errs.ErrOrderNotFound)
			},
			expectedErr: errs.ErrOrderNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orderRepo := mocks.NewOrderRepository(t)
			inventoryClient := mocks.NewInventoryClient(t)
			txManager := mocks.NewTxManager(t)
			paymentClient := mocks.NewPaymentClient(t)
			producer := mocks.NewOrderProducerService(t)

			tc.setupMock(orderRepo, inventoryClient, txManager, producer)

			svc := NewService(orderRepo, paymentClient, inventoryClient, txManager, producer)
			err := svc.CommitParts(ctx, tc.args.event)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
