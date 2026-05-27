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
	"github.com/FOCCms/go-microservices-course/platform/pkg/auth"
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
		userUUID = uuid.New()
		ctx      = auth.WithUserUUID(context.Background(), userUUID)

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
		tx *mocks.TxManager,
		producer *mocks.OrderProducerService,
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
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(nil)

				client.EXPECT().
					ListParts(ctx, mock.Anything).
					Return(partsInStock, nil)
				client.EXPECT().
					ValidateCompatibility(ctx, mock.Anything).
					Return(nil)
				client.EXPECT().
					ReserveParts(ctx, mock.Anything).
					Return(nil)

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
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errs.ErrPartNotFound)

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
			setupMock: func(repo *mocks.OrderRepository, client *mocks.InventoryClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errs.ErrPartNotFound)

				client.EXPECT().
					ListParts(ctx, []uuid.UUID{uuid.MustParse(hullUUID), uuid.MustParse(engineUUID)}).
					Return(partsOutOfStock, nil)

				client.EXPECT().
					ValidateCompatibility(ctx, mock.Anything).
					Return(errs.ErrPartNotFound)
			},
			expected: expected{err: errs.ErrPartNotFound, wantOrderUUID: false},
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

			tc.setupMock(orderRepo, inventoryClient, txManager, producer)

			svc := NewService(orderRepo, paymentClient, inventoryClient, txManager, producer)
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
