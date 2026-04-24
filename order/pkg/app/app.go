package app

import (
	orderV1API "github.com/FOCCms/go-microservices-course/order/internal/api/order/v1"
	invetntoryV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/inventory/v1"
	paymentV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/payment/v1"
	orderRepository "github.com/FOCCms/go-microservices-course/order/internal/repository/order"
	orderService "github.com/FOCCms/go-microservices-course/order/internal/service/order"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

func NewHTTPHandler(inventoryClient inventoryv1.InventoryServiceClient, paymentClient paymentv1.PaymentServiceClient) (*orderv1.Server, error) {
	repo := orderRepository.NewRepository()

	pc := paymentV1Client.New(paymentClient)
	ic := invetntoryV1Client.New(inventoryClient)

	service := orderService.NewService(repo, pc, ic)

	api := orderV1API.NewAPI(service)

	return orderV1API.SetupServer(api)
}
