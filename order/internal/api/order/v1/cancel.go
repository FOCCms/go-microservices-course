package v1

import (
	"context"
	"errors"
	"net/http"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
)

func (a *api) CancelOrder(ctx context.Context, params orderv1.CancelOrderParams) (orderv1.CancelOrderRes, error) {
	err := a.orderService.Cancel(ctx, params.OrderUUID)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			return &orderv1.CancelOrderNotFound{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}, nil
		}
		if errors.Is(err, errs.ErrOrderAlreadyPaid) || errors.Is(err, errs.ErrOrderCancelled) || errors.Is(err, errs.ErrOrderStatusConflict) {
			return &orderv1.CancelOrderConflict{
				Code:    http.StatusConflict,
				Message: err.Error(),
			}, nil
		}
		return &orderv1.CancelOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "ошибка отмены заказа",
		}, nil
	}

	return &orderv1.CancelOrderResponse{}, nil
}
