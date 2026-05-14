package order_consumer

import (
	"context"
	"log/slog"
)

type service struct {
	orderPaidConsumer Consumer
	assembleService   AssembleService
}

func NewService(orderPaidConsumer Consumer, assembleService AssembleService) *service {
	return &service{
		orderPaidConsumer: orderPaidConsumer,
		assembleService:   assembleService,
	}
}

func (s *service) RunConsumer(ctx context.Context) error {
	slog.InfoContext(ctx, "запуск потребителя OrderPaid")

	return s.orderPaidConsumer.Consume(ctx, s.OrderPaidHandler)
}
