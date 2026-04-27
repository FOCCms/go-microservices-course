package order

import (
	"context"

	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

type OrderRepository interface {
	Create(ctx context.Context, order model.Order) error
	Get(ctx context.Context, uuid uuid.UUID) (model.Order, error)
	Update(ctx context.Context, order model.Order) error
}

type PaymentClient interface {
	PayOrder(ctx context.Context, orderUUID uuid.UUID, method model.PaymentMethod) (uuid.UUID, error)
}

type InventoryClient interface {
	ListParts(ctx context.Context, uuids []uuid.UUID) ([]model.Part, error)
}
