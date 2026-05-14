package assembly_consumer

import (
	"context"
	"log/slog"
)

type service struct {
	shipAssembledConsumer Consumer
	orderService          OrderService
}

func NewService(shipAssembledConsumer Consumer, orderService OrderService) *service {
	return &service{
		shipAssembledConsumer: shipAssembledConsumer,
		orderService:          orderService,
	}
}

func (s *service) RunConsumer(ctx context.Context) error {
	slog.InfoContext(ctx, "запуск потребителя ShipAssembled")

	return s.shipAssembledConsumer.Consume(ctx, s.OrderPaidHandler)
}
