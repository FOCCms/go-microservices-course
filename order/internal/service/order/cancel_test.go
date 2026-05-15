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

func TestCancel(t *testing.T) {
	t.Parallel()

	type args struct {
		id string
	}

	var (
		ctx     = context.Background()
		orderID = gofakeit.UUID()
	)

	tests := []struct {
		name        string
		args        args
		setupMock   func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService)
		expectedErr error
	}{
		{
			name: "успешная отмена заказа",
			args: args{id: orderID},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(nil)

				repo.EXPECT().
					Get(ctx, orderID).
					Return(model.Order{UUID: uuid.MustParse(orderID), Status: model.OrderStatusPendingPayment}, nil)

				client.EXPECT().
					ReleaseParts(ctx, mock.Anything).
					Return(nil)

				repo.EXPECT().
					Update(ctx, mock.MatchedBy(func(o model.Order) bool {
						return o.Status == model.OrderStatusCancelled
					})).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "ошибка: заказ уже оплачен",
			args: args{id: orderID},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errs.ErrOrderAlreadyPaid)

				repo.EXPECT().
					Get(ctx, orderID).
					Return(model.Order{UUID: uuid.MustParse(orderID), Status: model.OrderStatusPaid}, nil)
			},
			expectedErr: errs.ErrOrderAlreadyPaid,
		},
		{
			name: "ошибка: заказ уже отменен",
			args: args{id: orderID},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errs.ErrOrderCancelled)

				repo.EXPECT().
					Get(ctx, orderID).
					Return(model.Order{UUID: uuid.MustParse(orderID), Status: model.OrderStatusCancelled}, nil)
			},
			expectedErr: errs.ErrOrderCancelled,
		},
		{
			name: "ошибка: заказ не найден",
			args: args{id: orderID},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errs.ErrOrderNotFound)

				repo.EXPECT().
					Get(ctx, orderID).
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
			err := svc.Cancel(ctx, uuid.MustParse(tc.args.id))

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
