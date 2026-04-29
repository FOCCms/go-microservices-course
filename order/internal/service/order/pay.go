package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (s *service) Pay(ctx context.Context, id uuid.UUID, method model.PaymentMethod) (uuid.UUID, error) {
	order, err := s.orderRepository.Get(ctx, id.String())
	if err != nil {
		return uuid.Nil, fmt.Errorf("оплатить заказ: %w", err)
	}

	if order.Status != model.OrderStatusPendingPayment {
		if order.Status == model.OrderStatusPaid {
			return uuid.Nil, fmt.Errorf("оплатить заказ: %w", errs.ErrOrderAlreadyPaid)
		}
		if order.Status == model.OrderStatusCancelled {
			return uuid.Nil, fmt.Errorf("оплатить заказ: %w", errs.ErrOrderCancelled)
		}
		return uuid.Nil, fmt.Errorf("оплатить заказ: %w", errs.ErrOrderStatusConflict)
	}

	transactionUUID, err := s.paymentClient.PayOrder(ctx, id, method)
	if err != nil {
		return uuid.Nil, fmt.Errorf("оплатить заказ: %w", err)
	}

	order.PaymentMethod = &method
	order.Status = model.OrderStatusPaid
	order.TransactionUUID = &transactionUUID

	err = s.orderRepository.Update(ctx, order)
	if err != nil {
		return uuid.Nil, fmt.Errorf("оплатить заказ: %w", err)
	}

	return transactionUUID, nil
}
