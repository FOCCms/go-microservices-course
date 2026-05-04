package app

import (
	"github.com/FOCCms/go-microservices-course/inventory/internal/app"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	partV1API "github.com/FOCCms/go-microservices-course/inventory/internal/api/inventory/v1"
	partRepository "github.com/FOCCms/go-microservices-course/inventory/internal/repository/part"
	partService "github.com/FOCCms/go-microservices-course/inventory/internal/service/part"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func RegisterServices(grpcServer *grpc.Server, pool *pgxpool.Pool) {
	repo := partRepository.NewRepository(pool)
	service := partService.NewService(repo)
	api := partV1API.NewAPI(service)
	inventoryv1.RegisterInventoryServiceServer(grpcServer, api)
}

func Interceptors() []grpc.ServerOption {
	return app.Interceptors()
}
