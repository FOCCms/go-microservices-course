package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

type OrderService interface {
	Create(ctx context.Context, req model.CreateOrderRequest) (model.Order, error)
	Get(ctx context.Context, id uuid.UUID) (model.Order, error)
	Pay(ctx context.Context, id uuid.UUID, method model.PaymentMethod) (uuid.UUID, error)
	Cancel(ctx context.Context, uuid uuid.UUID) error
}
