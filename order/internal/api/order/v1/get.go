package v1

import (
	"context"
	"net/http"

	"github.com/FOCCms/go-microservices-course/order/internal/converter"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
)

// GetOrder реализует операцию getOrder (пример реализации).
// GET /api/v1/orders/{order_uuid}.
func (a *api) GetOrder(ctx context.Context, params orderv1.GetOrderParams) (orderv1.GetOrderRes, error) {
	order, err := a.orderService.Get(ctx, params.OrderUUID)
	if err != nil {
		// TODO добавить остальные ошибки
		return &orderv1.GetOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "ошибка получения заказа",
		}, nil
	}

	return converter.OrderToOrderDto(order), nil
}
