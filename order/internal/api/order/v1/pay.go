package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/FOCCms/go-microservices-course/order/internal/converter"
	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
)

func (a *api) PayOrder(ctx context.Context, req *orderv1.PayOrderRequest, params orderv1.PayOrderParams) (orderv1.PayOrderRes, error) {
	transactionUUID, err := a.orderService.Pay(ctx, params.OrderUUID, converter.ToPaymentMethod(req.GetPaymentMethod()))
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			return &orderv1.PayOrderNotFound{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}, nil
		}
		if errors.Is(err, errs.ErrOrderAlreadyPaid) || errors.Is(err, errs.ErrOrderCancelled) {
			return &orderv1.PayOrderConflict{
				Code:    http.StatusConflict,
				Message: err.Error(),
			}, nil
		}
		return &orderv1.PayOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "ошибка оплаты заказа",
		}, nil
	}

	return &orderv1.PayOrderResponse{
		TransactionUUID: transactionUUID,
	}, nil
}
