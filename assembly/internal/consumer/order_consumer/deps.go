package order_consumer

import (
	"context"

	"github.com/FOCCms/go-microservices-course/assembly/internal/model"
	"github.com/FOCCms/go-microservices-course/platform/pkg/kafka"
)

type Consumer interface {
	Consume(ctx context.Context, handler kafka.MessageHandler) error
}

type AssembleService interface {
	Assemble(ctx context.Context, event model.OrderPaidEvent) error
}
