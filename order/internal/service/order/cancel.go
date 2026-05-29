package order

import (
	"context"
	"fmt"
	"log/slog"

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
			return fmt.Errorf("отменить заказ: %w", err)
		}

		err = s.inventoryClient.ReleaseParts(ctx, order.AssemblePartUUIDs())
		if err != nil {
			return fmt.Errorf("отменить заказ: освободить детали: %w", err)
		}

		order.Status = model.OrderStatusCancelled
		err = s.orderRepository.Update(ctx, order)
		if err != nil {
			return fmt.Errorf("отменить заказ: %w", err)
		}

		slog.Info("заказ успешно отменен",
			slog.String("order_uuid", order.UUID.String()),
			slog.String("user_uuid", order.UserUUID.String()),
			slog.Any("released_part_uuids", order.AssemblePartUUIDs()),
		)

		return nil
	})
	return err
}

func checkCancelStatus(status model.OrderStatus) error {
	if status != model.OrderStatusPendingPayment {
		if status == model.OrderStatusPaid {
			return errs.ErrOrderAlreadyPaid
		}
		if status == model.OrderStatusCancelled {
			return errs.ErrOrderCancelled
		}
		return errs.ErrOrderStatusConflict
	}
	return nil
}
