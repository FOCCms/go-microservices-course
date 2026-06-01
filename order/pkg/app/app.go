package app

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	orderV1API "github.com/FOCCms/go-microservices-course/order/internal/api/order/v1"
	iamV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/iam/v1"
	invetntoryV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/inventory/v1"
	paymentV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/payment/v1"
	orderMiddleware "github.com/FOCCms/go-microservices-course/order/internal/middleware"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
	orderRepository "github.com/FOCCms/go-microservices-course/order/internal/repository/order"
	orderService "github.com/FOCCms/go-microservices-course/order/internal/service/order"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
	authv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/auth/v1"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

type noopProducer struct{}

func (noopProducer) ProduceOrderPaid(_ context.Context, _ model.OrderPaidEvent) error {
	return nil
}

func NewHTTPHandler(pool *pgxpool.Pool, txManager orderService.TxManager, inventoryClient inventoryv1.InventoryServiceClient, paymentClient paymentv1.PaymentServiceClient, auth authv1.AuthServiceClient) (http.Handler, error) {
	return NewHTTPHandlerWithProducer(pool, txManager, inventoryClient, paymentClient, noopProducer{}, auth)
}

func NewHTTPHandlerWithProducer(
	pool *pgxpool.Pool,
	txManager orderService.TxManager,
	grpcInventoryClient inventoryv1.InventoryServiceClient,
	grpcPaymentClient paymentv1.PaymentServiceClient,
	producer orderService.OrderProducerService,
	auth authv1.AuthServiceClient,
) (http.Handler, error) {
	repo := orderRepository.NewRepository(pool, txManager)

	invClient := invetntoryV1Client.New(grpcInventoryClient)
	payClient := paymentV1Client.New(grpcPaymentClient)

	service := orderService.NewService(repo, payClient, invClient, txManager, producer)

	apiHandler := orderV1API.NewAPI(service)

	authClient := iamV1Client.New(auth)

	server, err := orderv1.NewServer(apiHandler, orderv1.WithErrorHandler(orderV1API.ErrorHandler))
	if err != nil {
		return nil, err
	}

	return orderMiddleware.AuthMiddleware(authClient)(server), nil
}
