package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (s *Service) Cancel(ctx context.Context, id uuid.UUID) error {
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		order, err := s.orderRepository.GetForUpdate(ctx, id.String())
		if err != nil {
			return fmt.Errorf("отменить заказ: %w", err)
		}

		if err = checkCancelStatus(order.Status); err != nil {
			return err
		}

		err = s.inventoryClient.ReleaseParts(ctx, order.AssemblePartUUIDs())
		if err != nil {
			return fmt.Errorf("отменить заказ: %w", err)
		}

		order.Status = model.OrderStatusCancelled
		err = s.orderRepository.Update(ctx, order)
		if err != nil {
			return fmt.Errorf("отменить заказ: не удалось освободить детали в инвентаре: %w", err)
		}

		return nil
	})
	return err
}

func checkCancelStatus(status model.OrderStatus) error {
	if status != model.OrderStatusPendingPayment {
		if status == model.OrderStatusPaid {
			return fmt.Errorf("отменить заказ: %w", errs.ErrOrderAlreadyPaid)
		}
		if status == model.OrderStatusCancelled {
			return fmt.Errorf("отменить заказ: %w", errs.ErrOrderCancelled)
		}
		return fmt.Errorf("отменить заказ: %w", errs.ErrOrderStatusConflict)
	}
	return nil
}
