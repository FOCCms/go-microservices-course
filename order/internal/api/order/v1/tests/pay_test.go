package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	v1 "github.com/FOCCms/go-microservices-course/order/internal/api/order/v1"
	"github.com/FOCCms/go-microservices-course/order/internal/api/order/v1/mocks"
	"github.com/FOCCms/go-microservices-course/order/internal/converter"
	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
)

func TestPayOrder(t *testing.T) {
	t.Parallel()

	var (
		ctx           = context.Background()
		orderID       = uuid.New()
		transactionID = uuid.New()
		paymentMethod = orderv1.PaymentMethodCARD // Пример метода
	)

	tests := []struct {
		name           string
		params         orderv1.PayOrderParams
		req            *orderv1.PayOrderRequest
		setupMock      func(svc *mocks.OrderService)
		expectedResult interface{}
	}{
		{
			name:   "успешная оплата (200)",
			params: orderv1.PayOrderParams{OrderUUID: orderID},
			req:    &orderv1.PayOrderRequest{PaymentMethod: paymentMethod},
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().
					Pay(ctx, orderID, converter.ToPaymentMethod(paymentMethod)).
					Return(transactionID, nil)
			},
			expectedResult: &orderv1.PayOrderResponse{TransactionUUID: transactionID},
		},
		{
			name:   "заказ не найден (404)",
			params: orderv1.PayOrderParams{OrderUUID: orderID},
			req:    &orderv1.PayOrderRequest{PaymentMethod: paymentMethod},
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Pay(ctx, orderID, mock.Anything).Return(uuid.Nil, errs.ErrOrderNotFound)
			},
			expectedResult: &orderv1.PayOrderNotFound{Code: 404, Message: errs.ErrOrderNotFound.Error()},
		},
		{
			name:   "конфликт: заказ уже оплачен (409)",
			params: orderv1.PayOrderParams{OrderUUID: orderID},
			req:    &orderv1.PayOrderRequest{PaymentMethod: paymentMethod},
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Pay(ctx, orderID, mock.Anything).Return(uuid.Nil, errs.ErrOrderAlreadyPaid)
			},
			expectedResult: &orderv1.PayOrderConflict{Code: 409, Message: errs.ErrOrderAlreadyPaid.Error()},
		},
		{
			name:   "внутренняя ошибка (500)",
			params: orderv1.PayOrderParams{OrderUUID: orderID},
			req:    &orderv1.PayOrderRequest{PaymentMethod: paymentMethod},
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Pay(ctx, orderID, mock.Anything).Return(uuid.Nil, errors.New("неизвестная ошибка"))
			},
			expectedResult: &orderv1.PayOrderInternalServerError{Code: 500, Message: "ошибка оплаты заказа"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := mocks.NewOrderService(t)
			tc.setupMock(mockSvc)

			a := v1.NewAPI(mockSvc)
			res, err := a.PayOrder(ctx, tc.req, tc.params)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, res)
		})
	}
}
