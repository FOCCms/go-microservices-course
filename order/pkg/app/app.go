package app

import (
	"log/slog"
	"os"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"

	orderV1API "github.com/FOCCms/go-microservices-course/order/internal/api/order/v1"
	invetntoryV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/inventory/v1"
	paymentV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/payment/v1"
	orderProducer "github.com/FOCCms/go-microservices-course/order/internal/producer/order_producer"
	orderRepository "github.com/FOCCms/go-microservices-course/order/internal/repository/order"
	orderService "github.com/FOCCms/go-microservices-course/order/internal/service/order"
	wrappedKafkaProducer "github.com/FOCCms/go-microservices-course/platform/pkg/kafka/producer"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

func NewHTTPHandler(pool *pgxpool.Pool, txManager orderService.TxManager, inventoryClient inventoryv1.InventoryServiceClient, paymentClient paymentv1.PaymentServiceClient) (*orderv1.Server, error) {
	orderRepo := orderRepository.NewRepository(pool, txManager)

	pc := paymentV1Client.New(paymentClient)
	ic := invetntoryV1Client.New(inventoryClient)

	orderProducerService := initProducer()

	service := orderService.NewService(orderRepo, pc, ic, txManager, orderProducerService)

	api := orderV1API.NewAPI(service)

	return orderV1API.SetupServer(api)
}

func initProducer() orderService.OrderProducerService {
	p, err := sarama.NewSyncProducer(
		[]string{"localhost:9092"},
		saramaConfig(),
	)
	if err != nil {
		slog.Error("initProducer", "error", err)
		os.Exit(1)
	}
	wrapped := wrappedKafkaProducer.NewProducer(p, "test-topic")

	srv := orderProducer.NewService(wrapped)
	return srv
}

func saramaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_0_0_0
	cfg.Producer.Return.Successes = true

	return cfg
}
