package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/FOCCms/go-microservices-course/order/internal/converter"
	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
)

func (a *api) GetOrder(ctx context.Context, params orderv1.GetOrderParams) (orderv1.GetOrderRes, error) {
	order, err := a.orderService.Get(ctx, params.OrderUUID)
	if err != nil {
		slog.Error("не удалось получить заказ",
			slog.String("error", err.Error()),
		)
		if errors.Is(err, errs.ErrOrderNotFound) {
			return &orderv1.GetOrderNotFound{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}, nil
		}
		return &orderv1.GetOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "ошибка получения заказа",
		}, nil
	}

	return converter.OrderToOrderDto(order), nil
}
