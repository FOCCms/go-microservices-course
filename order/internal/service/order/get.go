package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (s *service) Get(ctx context.Context, id uuid.UUID) (model.Order, error) {
	order, err := s.orderRepository.Get(ctx, id.String())
	if err != nil {
		return model.Order{}, fmt.Errorf("получить заказ: %w", err)
	}

	return order, nil
}
