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

func TestPay(t *testing.T) {
	t.Parallel()

	var (
		ctx             = context.Background()
		orderID         = uuid.MustParse(gofakeit.UUID())
		transactionUUID = uuid.MustParse(gofakeit.UUID())
		paymentMethod   = model.PaymentMethodCard
	)

	tests := []struct {
		name string
		args struct {
			id     uuid.UUID
			method model.PaymentMethod
		}
		setupMock   func(repo *mocks.OrderRepository, client *mocks.PaymentClient, tx *mocks.TxManager, producer *mocks.OrderProducerService)
		expected    uuid.UUID
		expectedErr error
	}{
		{
			name: "успешная оплата",
			args: struct {
				id     uuid.UUID
				method model.PaymentMethod
			}{id: orderID, method: paymentMethod},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.PaymentClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(nil)

				repo.EXPECT().
					GetForUpdate(ctx, orderID.String()).
					Return(model.Order{UUID: orderID, Status: model.OrderStatusPendingPayment}, nil)

				client.EXPECT().
					PayOrder(ctx, orderID, paymentMethod).
					Return(transactionUUID, nil)

				repo.EXPECT().
					Update(ctx, mock.MatchedBy(func(o model.Order) bool {
						return o.Status == model.OrderStatusPaid && *o.TransactionUUID == transactionUUID
					})).
					Return(nil)

				producer.EXPECT().
					ProduceOrderPaid(ctx, mock.Anything).
					Return(nil)
			},
			expected:    transactionUUID,
			expectedErr: nil,
		},
		{
			name: "ошибка: сервис оплаты вернул ошибку",
			args: struct {
				id     uuid.UUID
				method model.PaymentMethod
			}{id: orderID, method: paymentMethod},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.PaymentClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errs.ErrOrderNotFound)

				repo.EXPECT().
					GetForUpdate(ctx, orderID.String()).
					Return(model.Order{UUID: orderID, Status: model.OrderStatusPendingPayment}, nil)

				client.EXPECT().
					PayOrder(ctx, orderID, paymentMethod).
					Return(uuid.Nil, errs.ErrOrderNotFound)
			},
			expected:    uuid.Nil,
			expectedErr: errs.ErrOrderNotFound,
		},
		{
			name: "ошибка: статус заказа не PENDING_PAYMENT",
			args: struct {
				id     uuid.UUID
				method model.PaymentMethod
			}{id: orderID, method: paymentMethod},
			setupMock: func(repo *mocks.OrderRepository, client *mocks.PaymentClient, tx *mocks.TxManager, producer *mocks.OrderProducerService) {
				tx.EXPECT().Do(ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(ctx context.Context, f func(context.Context) error) {
						_ = f(ctx)
					}).
					Return(errs.ErrOrderCancelled)

				repo.EXPECT().
					GetForUpdate(ctx, orderID.String()).
					Return(model.Order{UUID: orderID, Status: model.OrderStatusCancelled}, nil)
			},
			expected:    uuid.Nil,
			expectedErr: errs.ErrOrderCancelled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orderRepo := mocks.NewOrderRepository(t)
			paymentClient := mocks.NewPaymentClient(t)
			inventoryClient := mocks.NewInventoryClient(t)
			txManager := mocks.NewTxManager(t)
			producer := mocks.NewOrderProducerService(t)
			tc.setupMock(orderRepo, paymentClient, txManager, producer)

			svc := NewService(orderRepo, paymentClient, inventoryClient, txManager, producer)
			txID, err := svc.Pay(ctx, tc.args.id, tc.args.method)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr.Error())
				assert.Equal(t, uuid.Nil, txID)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, txID)
			}
		})
	}
}
