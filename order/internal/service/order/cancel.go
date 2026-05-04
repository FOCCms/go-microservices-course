package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (s *service) Cancel(ctx context.Context, id uuid.UUID) error {
	order, err := s.orderRepository.Get(ctx, id.String())
	if err != nil {
		return fmt.Errorf("отменить заказ: %w", err)
	}

	if order.Status != model.OrderStatusPendingPayment {
		if order.Status == model.OrderStatusPaid {
			return fmt.Errorf("отменить заказ: %w", errs.ErrOrderAlreadyPaid)
		}
		if order.Status == model.OrderStatusCancelled {
			return fmt.Errorf("отменить заказ: %w", errs.ErrOrderCancelled)
		}
		return fmt.Errorf("отменить заказ: %w", errs.ErrOrderStatusConflict)
	}

	order.Status = model.OrderStatusCancelled
	err = s.orderRepository.Update(ctx, order)
	if err != nil {
		return fmt.Errorf("отменить заказ: %w", err)
	}

	return nil
}
