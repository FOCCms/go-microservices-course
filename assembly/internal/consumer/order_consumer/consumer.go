package order_consumer

import (
	"context"
	"log/slog"
)

type Service struct {
	orderPaidConsumer Consumer
	assembleService   AssembleService
}

func NewService(orderPaidConsumer Consumer, assembleService AssembleService) *Service {
	return &Service{
		orderPaidConsumer: orderPaidConsumer,
		assembleService:   assembleService,
	}
}

func (s *Service) RunConsumer(ctx context.Context) error {
	slog.InfoContext(ctx, "запуск потребителя OrderPaid")

	return s.orderPaidConsumer.Consume(ctx, s.OrderPaidHandler)
}
