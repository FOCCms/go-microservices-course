package app

import (
	"context"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	partV1API "github.com/FOCCms/go-microservices-course/inventory/internal/api/inventory/v1"
	iamV1Client "github.com/FOCCms/go-microservices-course/inventory/internal/client/grpc/iam/v1"
	"github.com/FOCCms/go-microservices-course/inventory/internal/config"
	"github.com/FOCCms/go-microservices-course/inventory/internal/interceptor"
	partRepository "github.com/FOCCms/go-microservices-course/inventory/internal/repository/part"
	partService "github.com/FOCCms/go-microservices-course/inventory/internal/service/application/part"
	"github.com/FOCCms/go-microservices-course/inventory/internal/service/domain"
	"github.com/FOCCms/go-microservices-course/platform/pkg/closer"
	authv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/auth/v1"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

type diContainer struct {
	pgPool    *pgxpool.Pool
	txManager partService.TxManager

	inventoryRepo partService.PartRepository

	compatibilityChecker partService.CompatibilityChecker

	inventoryService partV1API.PartService

	inventoryV1Handler inventoryv1.InventoryServiceServer

	iamClient interceptor.IAMClient
}

func (d *diContainer) PGPool(ctx context.Context) (*pgxpool.Pool, error) {
	if d.pgPool == nil {
		// Подключаем БД
		pool, err := pgxpool.New(ctx, config.AppConfig().PG.DSN())
		if err != nil {
			return nil, fmt.Errorf("создать пул соединений: %w", err)
		}

		// Проверяем соединение
		err = pool.Ping(ctx)
		if err != nil {
			return nil, fmt.Errorf("проверить соединение с БД: %w", err)
		}

		closer.Add("PostgreSQL pool", func(_ context.Context) error {
			pool.Close()
			return nil
		})

		d.pgPool = pool
	}
	return d.pgPool, nil
}

func (d *diContainer) TxManager(ctx context.Context) (partService.TxManager, error) {
	if d.txManager == nil {
		pool, err := d.PGPool(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать transaction manager: %w", err)
		}
		txManager, err := manager.New(trmpgx.NewDefaultFactory(pool))
		if err != nil {
			return nil, fmt.Errorf("создать transaction manager: %w", err)
		}
		d.txManager = txManager
	}

	return d.txManager, nil
}

func (d *diContainer) InventoryRepo(ctx context.Context) (partService.PartRepository, error) {
	if d.inventoryRepo == nil {
		pool, err := d.PGPool(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать репо: %w", err)
		}
		d.inventoryRepo = partRepository.NewRepository(pool)
	}

	return d.inventoryRepo, nil
}

func (d *diContainer) CompatibilityChecker() partService.CompatibilityChecker {
	if d.compatibilityChecker == nil {
		d.compatibilityChecker = domain.NewCompatibilityChecker()
	}

	return d.compatibilityChecker
}

func (d *diContainer) InventoryService(ctx context.Context) (partV1API.PartService, error) {
	if d.inventoryService == nil {
		repo, err := d.InventoryRepo(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать сервис: %w", err)
		}
		txManager, err := d.TxManager(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать сервис: %w", err)
		}

		d.inventoryService = partService.NewService(repo, d.CompatibilityChecker(), txManager)
	}

	return d.inventoryService, nil
}

func (d *diContainer) InventoryV1API(ctx context.Context) (inventoryv1.InventoryServiceServer, error) {
	if d.inventoryV1Handler == nil {
		service, err := d.InventoryService(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать хендлер: %w", err)
		}
		d.inventoryV1Handler = partV1API.NewAPI(service)
	}

	return d.inventoryV1Handler, nil
}

func (d *diContainer) IAMClient(_ context.Context) (interceptor.IAMClient, error) {
	if d.iamClient == nil {
		iamConn, err := grpc.NewClient(iamServiceAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                iamServiceKeepaliveTime,
				Timeout:             iamServiceKeepaliveTimeout,
				PermitWithoutStream: true,
			}))
		if err != nil {
			return nil, fmt.Errorf("инициализировать iam client: %w", err)
		}
		closer.Add("iam conn", func(_ context.Context) error {
			return iamConn.Close()
		})

		client := authv1.NewAuthServiceClient(iamConn)
		d.iamClient = iamV1Client.New(client)
	}
	return d.iamClient, nil
}
