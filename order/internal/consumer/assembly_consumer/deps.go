package assembly_consumer

import (
	"context"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
	"github.com/FOCCms/go-microservices-course/platform/pkg/kafka"
)

type Consumer interface {
	Consume(ctx context.Context, handler kafka.MessageHandler) error
}

type OrderService interface {
	CommitParts(ctx context.Context, event model.ShipAssembledEvent) error
}
