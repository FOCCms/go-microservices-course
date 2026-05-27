package app

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ogen-go/ogen/middleware"

	orderV1API "github.com/FOCCms/go-microservices-course/order/internal/api/order/v1"
	invetntoryV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/inventory/v1"
	paymentV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/payment/v1"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
	orderRepository "github.com/FOCCms/go-microservices-course/order/internal/repository/order"
	orderService "github.com/FOCCms/go-microservices-course/order/internal/service/order"
	"github.com/FOCCms/go-microservices-course/platform/pkg/auth"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

type noopProducer struct{}

func (noopProducer) ProduceOrderPaid(_ context.Context, _ model.OrderPaidEvent) error {
	return nil
}

func NewHTTPHandler(pool *pgxpool.Pool, txManager orderService.TxManager, inventoryClient inventoryv1.InventoryServiceClient, paymentClient paymentv1.PaymentServiceClient) (http.Handler, error) {
	return NewHTTPHandlerWithProducer(pool, txManager, inventoryClient, paymentClient, noopProducer{})
}

func NewHTTPHandlerWithProducer(
	pool *pgxpool.Pool,
	txManager orderService.TxManager,
	grpcInventoryClient inventoryv1.InventoryServiceClient,
	grpcPaymentClient paymentv1.PaymentServiceClient,
	producer orderService.OrderProducerService,
) (http.Handler, error) {
	repo := orderRepository.NewRepository(pool, txManager)

	invClient := invetntoryV1Client.New(grpcInventoryClient)
	payClient := paymentV1Client.New(grpcPaymentClient)

	service := orderService.NewService(repo, payClient, invClient, txManager, producer)

	apiHandler := orderV1API.NewAPI(service)

	server, err := orderv1.NewServer(apiHandler, orderv1.WithMiddleware(TestAuthMiddleware))
	if err != nil {
		return nil, err
	}

	return server, nil
}

func TestAuthMiddleware(req middleware.Request, next middleware.Next) (middleware.Response, error) {
	userUUID := "00000000-0000-0000-0000-000000000000"
	ctx := req.Context
	newCtx := auth.WithUserUUID(ctx, uuid.MustParse(userUUID))

	req.SetContext(newCtx)

	return next(req)
}
