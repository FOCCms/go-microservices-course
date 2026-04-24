package v1

import (
	"context"
	"net/http"

	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
)

func (a *api) CancelOrder(ctx context.Context, params orderv1.CancelOrderParams) (orderv1.CancelOrderRes, error) {
	err := a.orderService.Cancel(ctx, params.OrderUUID)
	if err != nil {
		// TODO добавить остальные ошибки
		return &orderv1.CancelOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "ошибка отмены заказа",
		}, nil
	}

	return &orderv1.CancelOrderResponse{}, nil
}
