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
	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
)

func TestCreateOrder(t *testing.T) {
	t.Parallel()

	var (
		ctx      = context.Background()
		orderID  = uuid.New()
		hullID   = uuid.New()
		engineID = uuid.New()
		total    = int64(800000)
	)

	tests := []struct {
		name           string
		req            *orderv1.CreateOrderRequest
		setupMock      func(svc *mocks.OrderService)
		expectedResult interface{}
	}{
		{
			name: "успешное создание (200)",
			req: &orderv1.CreateOrderRequest{
				HullUUID:   hullID,
				EngineUUID: engineID,
			},
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Create(ctx, mock.Anything).Return(model.Order{
					UUID:       orderID,
					TotalPrice: total,
				}, nil)
			},
			expectedResult: &orderv1.CreateOrderResponse{
				OrderUUID:  orderID,
				TotalPrice: total,
			},
		},
		{
			name: "деталь не найдена (404)",
			req:  &orderv1.CreateOrderRequest{HullUUID: hullID, EngineUUID: engineID},
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Create(ctx, mock.Anything).Return(model.Order{}, errs.ErrPartNotFound)
			},
			expectedResult: &orderv1.CreateOrderNotFound{Code: 404, Message: errs.ErrPartNotFound.Error()},
		},
		{
			name: "нет на складе (409)",
			req:  &orderv1.CreateOrderRequest{HullUUID: hullID, EngineUUID: engineID},
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Create(ctx, mock.Anything).Return(model.Order{}, errs.ErrOutOfStock)
			},
			expectedResult: &orderv1.CreateOrderConflict{Code: 409, Message: errs.ErrOutOfStock.Error()},
		},
		{
			name: "внутренняя ошибка (500)",
			req:  &orderv1.CreateOrderRequest{HullUUID: hullID, EngineUUID: engineID},
			setupMock: func(svc *mocks.OrderService) {
				svc.EXPECT().Create(ctx, mock.Anything).Return(model.Order{}, errors.New("неизвестная ошибка"))
			},
			expectedResult: &orderv1.CreateOrderInternalServerError{Code: 500, Message: "ошибка создания заказа"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := mocks.NewOrderService(t)
			tc.setupMock(mockSvc)

			a := v1.NewAPI(mockSvc)
			res, err := a.CreateOrder(ctx, tc.req)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, res)
		})
	}
}
