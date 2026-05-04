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

func TestCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		req model.CreateOrderRequest
	}

	type expected struct {
		err           error
		wantOrderUUID bool
	}

	var (
		ctx = context.Background()

		hullUUID   = gofakeit.UUID()
		engineUUID = gofakeit.UUID()

		partsInStock = []model.Part{
			{UUID: hullUUID, Name: "Hull", Price: 500000, StockQuantity: 10},
			{UUID: engineUUID, Name: "Engine", Price: 300000, StockQuantity: 5},
		}

		partsOutOfStock = []model.Part{
			{UUID: hullUUID, Name: "Hull", Price: 500000, StockQuantity: 10},
			{UUID: engineUUID, Name: "Engine", Price: 300000, StockQuantity: 0},
		}
	)

	type setupMockFunc func(
		repo *mocks.OrderRepository,
		client *mocks.InventoryClient,
	)

	tests := []struct {
		name      string
		args      args
		setupMock setupMockFunc
		expected  expected
	}{
		{
			name: "успешное создание заказа",
			args: args{
				req: model.CreateOrderRequest{
					HullUUID:   uuid.MustParse(hullUUID),
					EngineUUID: uuid.MustParse(engineUUID),
				},
			},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient) {
				client.EXPECT().
					ListParts(ctx, mock.Anything).
					Return(partsInStock, nil)

				repo.EXPECT().
					Create(ctx, mock.MatchedBy(func(o model.Order) bool {
						return o.HullUUID == uuid.MustParse(hullUUID) && o.TotalPrice == 800000
					}), mock.Anything).
					Return(nil)
			},
			expected: expected{err: nil, wantOrderUUID: true},
		},
		{
			name: "деталь не найдена",
			args: args{
				req: model.CreateOrderRequest{
					HullUUID:   uuid.MustParse(hullUUID),
					EngineUUID: uuid.MustParse(engineUUID),
				},
			},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient) {
				client.EXPECT().
					ListParts(ctx, []uuid.UUID{uuid.MustParse(hullUUID), uuid.MustParse(engineUUID)}).
					Return(nil, errs.ErrPartNotFound)
			},
			expected: expected{err: errs.ErrPartNotFound, wantOrderUUID: false},
		},
		{
			name: "деталь закончилась на складе",
			args: args{
				req: model.CreateOrderRequest{
					HullUUID:   uuid.MustParse(hullUUID),
					EngineUUID: uuid.MustParse(engineUUID),
				},
			},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient) {
				client.EXPECT().
					ListParts(ctx, []uuid.UUID{uuid.MustParse(hullUUID), uuid.MustParse(engineUUID)}).
					Return(partsOutOfStock, nil)
			},
			expected: expected{err: errs.ErrOutOfStock, wantOrderUUID: false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orderRepo := mocks.NewOrderRepository(t)
			inventoryClient := mocks.NewInventoryClient(t)
			paymentClient := mocks.NewPaymentClient(t)

			tc.setupMock(orderRepo, inventoryClient)

			svc := NewService(orderRepo, paymentClient, inventoryClient)
			order, err := svc.Create(ctx, tc.args.req)

			if tc.expected.err != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expected.err)
				assert.Empty(t, order.UUID)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, order.UUID)
			}
		})
	}
}
