package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

// PaymentServer реализует gRPC сервис оплаты.
type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
}

// PayOrder обрабатывает оплату заказа.
func (s *PaymentServer) PayOrder(
	ctx context.Context,
	req *paymentv1.PayOrderRequest,
) (*paymentv1.PayOrderResponse, error) {
	if req.GetOrderUuid() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "uuid не указан")
	}

	if _, err := uuid.Parse(req.GetOrderUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "неверный формат uuid: %s", req.GetOrderUuid())
	}

	if req.GetPaymentMethod() == paymentv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED {
		return nil, status.Errorf(codes.InvalidArgument, "неверный payment_method: %s", req.GetPaymentMethod())
	}

	transactionUuid := uuid.New()

	slog.Info("оплата прошла успешно",
		"order_uuid:", req.GetOrderUuid(), ", transaction_uuid:", transactionUuid,
	)

	return &paymentv1.PayOrderResponse{TransactionUuid: transactionUuid.String()}, nil
}
