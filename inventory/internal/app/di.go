package app

import (
	"context"
	"fmt"

	partV1API "github.com/FOCCms/go-microservices-course/inventory/internal/api/inventory/v1"
	"github.com/FOCCms/go-microservices-course/inventory/internal/config"
	partRepository "github.com/FOCCms/go-microservices-course/inventory/internal/repository/part"
	partService "github.com/FOCCms/go-microservices-course/inventory/internal/service/part"
	"github.com/FOCCms/go-microservices-course/platform/pkg/closer"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

type diContainer struct {
	pgPool             *pgxpool.Pool
	inventoryRepo      partService.PartRepository
	inventoryService   partV1API.PartService
	inventoryV1Handler inventoryv1.InventoryServiceServer
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

func (d *diContainer) InventoryService(ctx context.Context) (partV1API.PartService, error) {
	if d.inventoryService == nil {
		srv, err := d.InventoryRepo(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать сервис: %w", err)
		}
		d.inventoryService = partService.NewService(srv)
	}

	return d.inventoryService, nil
}

func (d *diContainer) InventoryV1API(ctx context.Context) (inventoryv1.InventoryServiceServer, error) {
	if d.inventoryV1Handler == nil {
		api, err := d.InventoryService(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать хендлер: %w", err)
		}
		d.inventoryV1Handler = partV1API.NewAPI(api)
	}

	return d.inventoryV1Handler, nil
}
