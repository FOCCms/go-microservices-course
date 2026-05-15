package app

import (
	"context"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"

	orderV1API "github.com/FOCCms/go-microservices-course/order/internal/api/order/v1"
	invetntoryV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/inventory/v1"
	paymentV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/payment/v1"
	orderRepository "github.com/FOCCms/go-microservices-course/order/internal/repository/order"
	orderService "github.com/FOCCms/go-microservices-course/order/internal/service/order"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

func NewHTTPHandler(pool *pgxpool.Pool, txManager orderService.TxManager, inventoryClient inventoryv1.InventoryServiceClient, paymentClient paymentv1.PaymentServiceClient) (*orderv1.Server, error) {
	orderRepo := orderRepository.NewRepository(pool, txManager)

	pc := paymentV1Client.New(paymentClient)
	ic := invetntoryV1Client.New(inventoryClient)

	orderProducerService := NewNoopProducer()

	service := orderService.NewService(orderRepo, pc, ic, txManager, orderProducerService)

	api := orderV1API.NewAPI(service)

	return orderV1API.SetupServer(api)
}

func NewHTTPHandlerWithProducer(pool *pgxpool.Pool, txManager orderService.TxManager, inventoryClient inventoryv1.InventoryServiceClient, paymentClient paymentv1.PaymentServiceClient, producer orderService.OrderProducerService) (*orderv1.Server, error) {
	orderRepo := orderRepository.NewRepository(pool, txManager)

	pc := paymentV1Client.New(paymentClient)
	ic := invetntoryV1Client.New(inventoryClient)

	service := orderService.NewService(orderRepo, pc, ic, txManager, producer)

	api := orderV1API.NewAPI(service)

	return orderV1API.SetupServer(api)
}

type noopProducer struct{}

func (n *noopProducer) ProduceOrderPaid(ctx context.Context, event model.OrderPaidEvent) error {
	return nil
}

func NewNoopProducer() *noopProducer {
	return &noopProducer{}
}
