package order

import (
	"context"

	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

type OrderRepository interface {
	Create(ctx context.Context, order model.Order, orderItems []model.OrderItem) error
	Get(ctx context.Context, uuid string) (model.Order, error)
	Update(ctx context.Context, order model.Order) error
}

type PaymentClient interface {
	PayOrder(ctx context.Context, orderUUID uuid.UUID, method model.PaymentMethod) (uuid.UUID, error)
}

type InventoryClient interface {
	ListParts(ctx context.Context, uuids []uuid.UUID) ([]model.Part, error)
	ValidateCompatibility(ctx context.Context, slots model.ShipSlots) error
	ReserveParts(ctx context.Context, uuids []uuid.UUID) error
	ReleaseParts(ctx context.Context, uuids []uuid.UUID) error
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type OrderProducerService interface {
	ProduceOrderPaid(ctx context.Context, event model.OrderPaidEvent) error
}
