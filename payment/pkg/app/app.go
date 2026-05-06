package app

import (
	"google.golang.org/grpc"

	paymentV1API "github.com/FOCCms/go-microservices-course/payment/internal/api/payment/v1"
	"github.com/FOCCms/go-microservices-course/payment/internal/app"
	paymentService "github.com/FOCCms/go-microservices-course/payment/internal/service/payment"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

func RegisterServices(grpcServer *grpc.Server) {
	service := paymentService.NewService()
	api := paymentV1API.NewAPI(service)
	paymentv1.RegisterPaymentServiceServer(grpcServer, api)
}

func Interceptors() []grpc.ServerOption {
	return app.Interceptors()
}
