package order

import (
	"context"
	"time"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (s *service) Create(ctx context.Context, req model.CreateOrderRequest) (model.Order, error) {
	if req.HullUUID == uuid.Nil || req.EngineUUID == uuid.Nil {
		return model.Order{}, errs.ErrPartRequired
	}

	parts, err := s.inventoryClient.ListParts(ctx, req.PartUUIDs())
	if err != nil {
		return model.Order{}, err
	}

	if len(parts) != len(req.PartUUIDs()) {
		return model.Order{}, errs.ErrPartNotFound
	}

	var totalPrice int64 = 0
	for _, part := range parts {
		if part.StockQuantity <= 0 {
			return model.Order{}, errs.ErrOutOfStock
		}

		totalPrice += part.Price
	}

	orderUUID := uuid.New()

	order := model.Order{
		UUID:       orderUUID,
		HullUUID:   req.HullUUID,
		EngineUUID: req.EngineUUID,
		ShieldUUID: req.ShieldUUID,
		WeaponUUID: req.WeaponUUID,
		TotalPrice: totalPrice,
		Status:     model.OrderStatusPendingPayment,
		CreatedAt:  time.Now(),
	}

	s.orderRepository.Create(ctx, order)

	return order, nil
}
