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

func TestGet(t *testing.T) {
	t.Parallel()

	type args struct {
		id string
	}

	type expected struct {
		err   error
		order model.Order
	}

	var (
		ctx       = context.Background()
		orderID   = gofakeit.UUID()
		mockOrder = model.Order{
			UUID: uuid.MustParse(orderID),
		}
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(repo *mocks.OrderRepository)
		expected  expected
	}{
		{
			name: "успешное получение заказа",
			args: args{id: orderID},
			setupMock: func(repo *mocks.OrderRepository) {
				repo.EXPECT().
					Get(mock.Anything, orderID).
					Return(mockOrder, nil)
			},
			expected: expected{err: nil, order: mockOrder},
		},
		{
			name: "заказ не найден",
			args: args{id: orderID},
			setupMock: func(repo *mocks.OrderRepository) {
				repo.EXPECT().
					Get(mock.Anything, orderID).
					Return(model.Order{}, errs.ErrOrderNotFound)
			},
			expected: expected{err: errs.ErrOrderNotFound, order: model.Order{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orderRepo := mocks.NewOrderRepository(t)
			inventoryClient := mocks.NewInventoryClient(t)
			paymentClient := mocks.NewPaymentClient(t)
			txManager := mocks.NewTxManager(t)
			producer := mocks.NewOrderProducerService(t)
			tc.setupMock(orderRepo)

			svc := NewService(orderRepo, paymentClient, inventoryClient, txManager, producer)
			order, err := svc.Get(ctx, uuid.MustParse(tc.args.id))

			if tc.expected.err != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expected.err)
				assert.Equal(t, model.Order{}, order)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.order, order)
			}
		})
	}
}
