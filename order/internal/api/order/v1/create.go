package v1

import (
	"context"
	"net/http"

	"github.com/FOCCms/go-microservices-course/order/internal/converter"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
)

func (a *api) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (orderv1.CreateOrderRes, error) {
	r := converter.ToCreateOrderRequest(*req)
	order, err := a.orderService.Create(ctx, r)
	if err != nil {
		// TODO добавить остальные ошибки
		return &orderv1.CreateOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "ошибка создания заказа",
		}, nil
	}

	return &orderv1.CreateOrderResponse{
		OrderUUID:  order.UUID,
		TotalPrice: order.TotalPrice,
	}, nil
}
