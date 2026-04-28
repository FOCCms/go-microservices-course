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
		setupMock   func(repo *mocks.OrderRepository)
		expectedErr error
	}{
		{
			name: "успешная отмена заказа",
			args: args{id: orderID},
			setupMock: func(repo *mocks.OrderRepository) {
				repo.EXPECT().
					Get(ctx, orderID).
					Return(model.Order{UUID: uuid.MustParse(orderID), Status: model.OrderStatusPendingPayment}, nil)

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
			setupMock: func(repo *mocks.OrderRepository) {
				repo.EXPECT().
					Get(ctx, orderID).
					Return(model.Order{UUID: uuid.MustParse(orderID), Status: model.OrderStatusPaid}, nil)
			},
			expectedErr: errs.ErrOrderAlreadyPaid,
		},
		{
			name: "ошибка: заказ уже отменен",
			args: args{id: orderID},
			setupMock: func(repo *mocks.OrderRepository) {
				repo.EXPECT().
					Get(ctx, orderID).
					Return(model.Order{UUID: uuid.MustParse(orderID), Status: model.OrderStatusCancelled}, nil)
			},
			expectedErr: errs.ErrOrderCancelled,
		},
		{
			name: "ошибка: заказ не найден",
			args: args{id: orderID},
			setupMock: func(repo *mocks.OrderRepository) {
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
			orderItemRepo := mocks.NewOrderItemRepository(t)
			inventoryClient := mocks.NewInventoryClient(t)
			paymentClient := mocks.NewPaymentClient(t)
			txManager := mocks.NewTxManager(t)

			tc.setupMock(orderRepo)

			svc := NewService(orderRepo, orderItemRepo, paymentClient, inventoryClient, txManager)
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
