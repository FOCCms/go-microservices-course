package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/FOCCms/go-microservices-course/order/internal/converter"
	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
)

func (a *api) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (orderv1.CreateOrderRes, error) {
	r := converter.ToCreateOrderRequest(*req)
	order, err := a.orderService.Create(ctx, r)
	if err != nil {
		if errors.Is(err, errs.ErrPartNotFound) {
			return &orderv1.CreateOrderNotFound{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}, nil
		}
		if errors.Is(err, errs.ErrOutOfStock) {
			return &orderv1.CreateOrderConflict{
				Code:    http.StatusConflict,
				Message: err.Error(),
			}, nil
		}
		if errors.Is(err, errs.ErrPartTypeMismatch) {
			return &orderv1.CreateOrderBadRequest{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}, nil
		}
		if errors.Is(err, errs.ErrIncompatibleParts) {
			return &orderv1.CreateOrderConflict{
				Code:    http.StatusConflict,
				Message: err.Error(),
			}, nil
		}
		if errors.Is(err, errs.ErrIncompatibleParts) {
			return &orderv1.CreateOrderConflict{
				Code:    http.StatusConflict,
				Message: err.Error(),
			}, nil
		}
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
