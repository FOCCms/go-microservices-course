package order

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/FOCCms/go-microservices-course/order/internal/converter"
	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
	"github.com/FOCCms/go-microservices-course/platform/pkg/auth"
)

func (s *Service) Create(ctx context.Context, req model.CreateOrderRequest) (model.Order, error) {
	ctx, span := otel.Tracer("order-service").Start(ctx, "order.Create")
	defer span.End()

	var shieldStr, weaponStr string
	if req.ShieldUUID != nil {
		shieldStr = req.ShieldUUID.String()
	}
	if req.WeaponUUID != nil {
		weaponStr = req.WeaponUUID.String()
	}

	span.SetAttributes(
		attribute.String("order.hull_uuid", req.HullUUID.String()),
		attribute.String("order.engine_uuid", req.EngineUUID.String()),
		attribute.String("order.shield_uuid", shieldStr),
		attribute.String("order.weapon_uuid", weaponStr),
	)

	if req.HullUUID == uuid.Nil || req.EngineUUID == uuid.Nil {
		span.RecordError(errs.ErrPartRequired)
		span.SetStatus(codes.Error, errs.ErrPartRequired.Error())
		return model.Order{}, fmt.Errorf("создать заказ: %w", errs.ErrPartRequired)
	}
	userUuid, ok := auth.UserUUIDFromContext(ctx)
	if !ok {
		span.RecordError(errs.ErrUnauthorized)
		span.SetStatus(codes.Error, errs.ErrUnauthorized.Error())
		return model.Order{}, fmt.Errorf("создать заказ: %w", errs.ErrUnauthorized)
	}

	order, err := s.createOrderTx(ctx, req, userUuid)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.Order{}, err
	}

	span.SetAttributes(
		attribute.String("order.uuid", order.UUID.String()),
		attribute.Int64("order.total_price", order.TotalPrice),
		attribute.String("order.status", string(order.Status)),
	)
	span.SetStatus(codes.Ok, "заказ успешно создан")

	slog.Info("заказ успешно создан",
		slog.String("order_uuid", order.UUID.String()),
		slog.String("user_uuid", order.UserUUID.String()),
		slog.Int64("total_price", order.TotalPrice),
		slog.String("status", string(order.Status)),
		slog.Any("part_uuids", req.AssemblePartUUIDs()),
	)
	ordersCreatedTotal.Add(ctx, 1)

	return order, nil
}

func (s *Service) createOrderTx(ctx context.Context, req model.CreateOrderRequest, userUuid uuid.UUID) (model.Order, error) {
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
			UserUUID:   userUuid,
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

func listParts(ctx context.Context, req model.CreateOrderRequest, s *Service) ([]model.Part, error) {
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
