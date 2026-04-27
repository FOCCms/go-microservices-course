package v1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/FOCCms/go-microservices-course/payment/internal/converter"
	errs "github.com/FOCCms/go-microservices-course/payment/internal/errors"
	"github.com/FOCCms/go-microservices-course/payment/internal/model"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

func (a *api) PayOrder(ctx context.Context, req *paymentv1.PayOrderRequest) (*paymentv1.PayOrderResponse, error) {
	if req.OrderUuid == "" {
		return nil, status.Errorf(codes.InvalidArgument, "uuid не указан")
	}

	transactionUuid, err := a.paymentService.Pay(ctx, model.PayRequest{
		OrderUUID:     req.OrderUuid,
		PaymentMethod: converter.ProtoPartTypeToPartType(req.PaymentMethod),
	})
	if err != nil {
		if errors.Is(err, errs.ErrInvalidOrderUUID) {
			return nil, status.Errorf(codes.InvalidArgument, "неверный формат uuid: %s", req.GetOrderUuid())
		}
		if errors.Is(err, errs.ErrInvalidPaymentMethod) {
			return nil, status.Errorf(codes.InvalidArgument, "неверный payment_method: %s", req.GetPaymentMethod())
		}
		return nil, status.Errorf(codes.Internal, "ошибка оплаты: %v", err)
	}

	return &paymentv1.PayOrderResponse{
		TransactionUuid: transactionUuid,
	}, nil
}
