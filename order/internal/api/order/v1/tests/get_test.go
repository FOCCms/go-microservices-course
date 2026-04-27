package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	v1 "github.com/FOCCms/go-microservices-course/order/internal/api/order/v1"
	"github.com/FOCCms/go-microservices-course/order/internal/api/order/v1/mocks"
	"github.com/FOCCms/go-microservices-course/order/internal/converter"
	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
)

func TestGetOrder(t *testing.T) {
	t.Parallel()

	var (
		ctx        = context.Background()
		orderID    = uuid.New()
		orderModel = model.Order{
			UUID:   orderID,
			Status: model.OrderStatusPendingPayment,
		}
		orderDto = converter.OrderToOrderDto(orderModel)
	)

	tests := []struct {
		name           string
		params         orderv1.GetOrderParams
		setupMock      func(svc *mocks.OrderService)
		expectedResult interface{}
	}{
		{
			name:   "успешное получение (200)",
			params: orderv1.GetOrderParams{OrderUUID: orderID},
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Get(ctx, orderID).Return(orderModel, nil)
			},
			expectedResult: orderDto,
		},
		{
			name:   "заказ не найден (404)",
			params: orderv1.GetOrderParams{OrderUUID: orderID},
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Get(ctx, orderID).Return(model.Order{}, errs.ErrOrderNotFound)
			},
			expectedResult: &orderv1.GetOrderNotFound{Code: 404, Message: errs.ErrOrderNotFound.Error()},
		},
		{
			name:   "внутренняя ошибка (500)",
			params: orderv1.GetOrderParams{OrderUUID: orderID},
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Get(ctx, orderID).Return(model.Order{}, errors.New("неизвестная ошибка"))
			},
			expectedResult: &orderv1.GetOrderInternalServerError{Code: 500, Message: "ошибка получения заказа"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := mocks.NewOrderService(t)
			tc.setupMock(mockSvc)

			a := v1.NewAPI(mockSvc)
			res, err := a.GetOrder(ctx, tc.params)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, res)
		})
	}
}
