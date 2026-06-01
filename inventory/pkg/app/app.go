package app

import (
	"log/slog"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	partV1API "github.com/FOCCms/go-microservices-course/inventory/internal/api/inventory/v1"
	iamV1Client "github.com/FOCCms/go-microservices-course/inventory/internal/client/grpc/iam/v1"
	"github.com/FOCCms/go-microservices-course/inventory/internal/interceptor"
	partRepository "github.com/FOCCms/go-microservices-course/inventory/internal/repository/part"
	partService "github.com/FOCCms/go-microservices-course/inventory/internal/service/application/part"
	"github.com/FOCCms/go-microservices-course/inventory/internal/service/domain"
	authv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/auth/v1"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func RegisterServices(grpcServer *grpc.Server, pool *pgxpool.Pool) {
	repo := partRepository.NewRepository(pool)
	checker := domain.NewCompatibilityChecker()
	txManager, err := manager.New(trmpgx.NewDefaultFactory(pool))
	if err != nil {
		slog.Error("Ошибка инициализации RegisterServices для тестов", "error", err)
	}
	service := partService.NewService(repo, checker, txManager)
	api := partV1API.NewAPI(service)
	inventoryv1.RegisterInventoryServiceServer(grpcServer, api)
}

func Interceptors(authClient authv1.AuthServiceClient) []grpc.ServerOption {
	authInterceptor := interceptor.AuthIncomingInterceptor(iamV1Client.New(authClient))
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryErrorInterceptor,
			authInterceptor,
		),
	}
}
