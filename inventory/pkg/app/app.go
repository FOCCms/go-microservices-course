package app

import (
	"google.golang.org/grpc"

	partV1API "github.com/FOCCms/go-microservices-course/inventory/internal/api/inventory/v1"
	"github.com/FOCCms/go-microservices-course/inventory/internal/interceptor"
	partRepository "github.com/FOCCms/go-microservices-course/inventory/internal/repository/part"
	partService "github.com/FOCCms/go-microservices-course/inventory/internal/service/part"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func RegisterServices(grpcServer *grpc.Server) {
	repo := partRepository.NewRepository()
	service := partService.NewService(repo)
	api := partV1API.NewAPI(service)
	inventoryv1.RegisterInventoryServiceServer(grpcServer, api)
}

func Interceptors() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.UnaryErrorInterceptor),
	}
}
