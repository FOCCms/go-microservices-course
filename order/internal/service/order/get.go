package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (s *Service) Get(ctx context.Context, id uuid.UUID) (model.Order, error) {
	ctx, span := otel.Tracer("order-service").Start(ctx, "order.Get")
	defer span.End()

	order, err := s.orderRepository.Get(ctx, id.String())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.Order{}, fmt.Errorf("получить заказ: %w", err)
	}

	return order, nil
}
