package v1

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/FOCCms/go-microservices-course/order/internal/client/grpc/payment/v1/converter"
	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

type client struct {
	paymentClient paymentv1.PaymentServiceClient
}

func New(c paymentv1.PaymentServiceClient) *client {
	return &client{paymentClient: c}
}

func (c *client) PayOrder(ctx context.Context, orderUUID uuid.UUID, method model.PaymentMethod) (uuid.UUID, error) {
	resp, err := c.paymentClient.PayOrder(ctx, &paymentv1.PayOrderRequest{
		OrderUuid:     orderUUID.String(),
		PaymentMethod: converter.ToPaymentMethod(method),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return uuid.Nil, errs.ErrOrderNotFound
		}
		return uuid.Nil, fmt.Errorf("оплатить заказ: %w", err)
	}

	id, err := uuid.Parse(resp.GetTransactionUuid())
	if err != nil {
		return uuid.Nil, errs.ErrInvalidUUID
	}

	return id, nil
}
