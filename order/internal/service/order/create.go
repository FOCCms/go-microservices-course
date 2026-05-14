package order

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/order/internal/converter"
	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (s *service) Create(ctx context.Context, req model.CreateOrderRequest) (model.Order, error) {
	if req.HullUUID == uuid.Nil || req.EngineUUID == uuid.Nil {
		return model.Order{}, fmt.Errorf("создать заказ: %w", errs.ErrPartRequired)
	}

	var order model.Order

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		parts, err := listParts(ctx, req, s)
		if err != nil {
			return fmt.Errorf("создать заказ: %w", err)
		}

		err = s.inventoryClient.ValidateCompatibility(ctx, converter.ToShipSlots(req))
		if err != nil {
			return fmt.Errorf("создать заказ: %w", err)
		}

		totalPrice, err := countTotalPrice(parts)
		if err != nil {
			return fmt.Errorf("создать заказ: %w", err)
		}

		err = s.inventoryClient.ReserveParts(ctx, req.AssemblePartUUIDs())
		if err != nil {
			return fmt.Errorf("создать заказ: %w", err)
		}

		orderUUID := uuid.New()

		order = model.Order{
			UUID:       orderUUID,
			HullUUID:   req.HullUUID,
			EngineUUID: req.EngineUUID,
			ShieldUUID: req.ShieldUUID,
			WeaponUUID: req.WeaponUUID,
			TotalPrice: totalPrice,
			Status:     model.OrderStatusPendingPayment,
			CreatedAt:  time.Now(),
			UserUUID:   req.UserUUID,
		}

		items := make([]model.OrderItem, len(parts))
		for i, part := range parts {
			items[i] = model.OrderItem{
				UUID:      uuid.New(),
				OrderUUID: order.UUID,
				PartUUID:  uuid.MustParse(part.UUID),
				PartType:  part.PartType,
				Price:     part.Price,
				CreatedAt: time.Now(),
			}
		}

		err = s.orderRepository.Create(ctx, order, items)
		if err != nil {
			return fmt.Errorf("создать заказ: %w", err)
		}

		return nil
	})
	if err != nil {
		return model.Order{}, err
	}

	return order, nil
}

func listParts(ctx context.Context, req model.CreateOrderRequest, s *service) ([]model.Part, error) {
	parts, err := s.inventoryClient.ListParts(ctx, req.AssemblePartUUIDs())
	if err != nil {
		return nil, err
	}

	if len(parts) != len(req.AssemblePartUUIDs()) {
		return nil, errs.ErrPartNotFound
	}
	return parts, nil
}

func countTotalPrice(parts []model.Part) (int64, error) {
	var totalPrice int64
	for _, part := range parts {
		if part.StockQuantity <= 0 {
			return 0, errs.ErrOutOfStock
		}

		totalPrice += part.Price
	}
	return totalPrice, nil
}
