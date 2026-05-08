package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	partV1API "github.com/FOCCms/go-microservices-course/inventory/internal/api/inventory/v1"
	"github.com/FOCCms/go-microservices-course/inventory/internal/app"
	partRepository "github.com/FOCCms/go-microservices-course/inventory/internal/repository/part"
	partService "github.com/FOCCms/go-microservices-course/inventory/internal/service/application/part"
	"github.com/FOCCms/go-microservices-course/inventory/internal/service/domain"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func RegisterServices(grpcServer *grpc.Server, pool *pgxpool.Pool) {
	repo := partRepository.NewRepository(pool)
	checker := domain.NewCompatibilityChecker()
	service := partService.NewService(repo, checker)
	api := partV1API.NewAPI(service)
	inventoryv1.RegisterInventoryServiceServer(grpcServer, api)
}

func Interceptors() []grpc.ServerOption {
	return app.Interceptors()
}
