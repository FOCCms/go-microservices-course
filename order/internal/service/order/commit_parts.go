package order

import (
	"context"
	"fmt"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (s *Service) CommitParts(ctx context.Context, event model.ShipAssembledEvent) error {
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		order, err := s.orderRepository.Get(ctx, event.OrderUUID.String())
		if err != nil {
			return fmt.Errorf("списать детали: %w", err)
		}

		if err = checkCommitStatus(order.Status); err != nil {
			return err
		}

		err = s.inventoryClient.CommitParts(ctx, order.AssemblePartUUIDs())
		if err != nil {
			return fmt.Errorf("списать детали: %w", err)
		}

		order.Status = model.OrderStatusAssembled

		err = s.orderRepository.Update(ctx, order)
		if err != nil {
			return fmt.Errorf("списать детали: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("списать детали в транзакции: %w", err)
	}

	return nil
}

func checkCommitStatus(status model.OrderStatus) error {
	if status != model.OrderStatusPaid {
		if status == model.OrderStatusAssembled {
			return fmt.Errorf("списать детали: %w", errs.ErrOrderAssembled)
		}
		if status == model.OrderStatusCancelled {
			return fmt.Errorf("списать детали: %w", errs.ErrOrderCancelled)
		}
		return fmt.Errorf("списать детали: %w", errs.ErrOrderStatusConflict)
	}
	return nil
}
