package order

import (
	"context"
	"fmt"
	"log/slog"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (s *Service) CommitParts(ctx context.Context, event model.ShipAssembledEvent) error {
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		order, err := s.orderRepository.GetForUpdate(ctx, event.OrderUUID.String())
		if err != nil {
			return fmt.Errorf("списать детали: %w", err)
		}

		if order.Status == model.OrderStatusAssembled {
			slog.Debug("событие ShipAssembled пропущено: заказ уже собран (идемпотентность)",
				slog.String("order_uuid", event.OrderUUID.String()),
			)
			return nil
		}
		if err = checkCommitStatus(order.Status); err != nil {
			return fmt.Errorf("списать детали: %w", err)
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

		slog.Info("детали успешно списаны со склада для собранного корабля",
			slog.String("order_uuid", order.UUID.String()),
			slog.String("user_uuid", order.UserUUID.String()),
		)

		return nil
	})
	if err != nil {
		return fmt.Errorf("списать детали в транзакции: %w", err)
	}

	return nil
}

func checkCommitStatus(status model.OrderStatus) error {
	if status != model.OrderStatusPaid {
		if status == model.OrderStatusCancelled {
			return errs.ErrOrderCancelled
		}
		return errs.ErrOrderStatusConflict
	}
	return nil
}
