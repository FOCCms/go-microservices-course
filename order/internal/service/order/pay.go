package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (s *Service) Pay(ctx context.Context, id uuid.UUID, method model.PaymentMethod) (uuid.UUID, error) {
	var transactionUUID uuid.UUID
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		order, err := s.orderRepository.Get(ctx, id.String())
		if err != nil {
			return fmt.Errorf("оплатить заказ: %w", err)
		}

		if err = checkPayStatus(order.Status); err != nil {
			return err
		}

		transactionUUID, err = s.paymentClient.PayOrder(ctx, id, method)
		if err != nil {
			return fmt.Errorf("оплатить заказ: %w", err)
		}

		order.PaymentMethod = &method
		order.Status = model.OrderStatusPaid
		order.TransactionUUID = &transactionUUID

		err = s.orderRepository.Update(ctx, order)
		if err != nil {
			return fmt.Errorf("оплатить заказ: %w", err)
		}

		err = s.orderProducerService.ProduceOrderPaid(ctx, model.OrderPaidEvent{
			EventUUID: uuid.New(),
			OrderUUID: order.UUID,
			UserUUID:  order.UserUUID,
		})
		if err != nil {
			return fmt.Errorf("оплатить заказ: %w", err)
		}

		return nil
	})

	return transactionUUID, err
}

func checkPayStatus(status model.OrderStatus) error {
	if status != model.OrderStatusPendingPayment {
		if status == model.OrderStatusPaid {
			return fmt.Errorf("оплатить заказ: %w", errs.ErrOrderAlreadyPaid)
		}
		if status == model.OrderStatusCancelled {
			return fmt.Errorf("оплатить заказ: %w", errs.ErrOrderCancelled)
		}
		return fmt.Errorf("оплатить заказ: %w", errs.ErrOrderStatusConflict)
	}
	return nil
}
