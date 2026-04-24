package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (s *service) Pay(ctx context.Context, id uuid.UUID, method model.PaymentMethod) (uuid.UUID, error) {
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("оплатить заказ: %w", err)
	}

	if order.Status != model.OrderStatusPendingPayment {
		return uuid.Nil, errs.ErrOrderStatusConflict
	}

	transactionUUID, err := s.paymentClient.PayOrder(ctx, id, method)
	if err != nil {
		return uuid.Nil, err
	}

	order, err = s.orderRepository.Get(ctx, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("оплатить заказ: %w", err)
	}

	if order.Status != model.OrderStatusPendingPayment {
		return uuid.Nil, errs.ErrOrderStatusConflict
	}

	order.PaymentMethod = &method
	order.Status = model.OrderStatusPaid
	order.TransactionUUID = &transactionUUID

	s.orderRepository.Update(ctx, order)

	return transactionUUID, nil
}
