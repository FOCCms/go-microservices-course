package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	v1 "github.com/FOCCms/go-microservices-course/order/internal/api/order/v1"
	"github.com/FOCCms/go-microservices-course/order/internal/api/order/v1/mocks"
	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
)

func TestCancelOrder(t *testing.T) {
	t.Parallel()

	var (
		ctx     = context.Background()
		orderID = uuid.MustParse(gofakeit.UUID())
	)

	tests := []struct {
		name           string
		setupMock      func(svc *mocks.OrderService)
		expectedResult interface{}
	}{
		{
			name: "успешная отмена (200)",
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Cancel(ctx, orderID).Return(nil)
			},
			expectedResult: &orderv1.CancelOrderResponse{},
		},
		{
			name: "заказ не найден (404)",
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Cancel(ctx, orderID).Return(errs.ErrOrderNotFound)
			},
			expectedResult: &orderv1.CancelOrderNotFound{Code: 404, Message: errs.ErrOrderNotFound.Error()},
		},
		{
			name: "конфликт: заказ уже оплачен (409)",
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Cancel(ctx, orderID).Return(errs.ErrOrderAlreadyPaid)
			},
			expectedResult: &orderv1.CancelOrderConflict{Code: 409, Message: errs.ErrOrderAlreadyPaid.Error()},
		},
		{
			name: "внутренняя ошибка (500)",
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Cancel(ctx, orderID).Return(errors.New("неизвестная ошибка"))
			},
			expectedResult: &orderv1.CancelOrderInternalServerError{Code: 500, Message: "ошибка отмены заказа"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := mocks.NewOrderService(t)
			tc.setupMock(mockSvc)

			a := v1.NewAPI(mockSvc)
			res, err := a.CancelOrder(ctx, orderv1.CancelOrderParams{OrderUUID: orderID})

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, res)
		})
	}
}
