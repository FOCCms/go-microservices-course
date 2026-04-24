package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (s *service) Cancel(ctx context.Context, id uuid.UUID) error {
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("отменить заказ: %w", err)
	}

	if order.Status != model.OrderStatusPendingPayment {
		return errs.ErrOrderStatusConflict
	}

	order.Status = model.OrderStatusCancelled
	err = s.orderRepository.Update(ctx, order)
	if err != nil {
		return fmt.Errorf("отменить заказ: %w", err)
	}

	return nil
}
