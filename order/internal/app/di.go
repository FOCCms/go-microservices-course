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

	orderV1API "github.com/FOCCms/go-microservices-course/order/internal/api/order/v1"
	invetntoryV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/inventory/v1"
	paymentV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/payment/v1"
	"github.com/FOCCms/go-microservices-course/order/internal/config"
	orderRepository "github.com/FOCCms/go-microservices-course/order/internal/repository/order"
	orderSrv "github.com/FOCCms/go-microservices-course/order/internal/service/order"
	"github.com/FOCCms/go-microservices-course/platform/pkg/closer"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

type diContainer struct {
	pgPool          *pgxpool.Pool
	txManager       orderSrv.TxManager
	orderRepo       orderSrv.OrderRepository
	inventoryClient orderSrv.InventoryClient
	paymentClient   orderSrv.PaymentClient
	orderService    orderV1API.OrderService
	orderV1Handler  *orderv1.Server
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

func (d *diContainer) TxManager(ctx context.Context) (orderSrv.TxManager, error) {
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

func (d *diContainer) OrderRepo(ctx context.Context) (orderSrv.OrderRepository, error) {
	if d.orderRepo == nil {
		pool, err := d.PGPool(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать репо: %w", err)
		}
		txManager, err := d.TxManager(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать репо: %w", err)
		}
		d.orderRepo = orderRepository.NewRepository(pool, txManager)
	}

	return d.orderRepo, nil
}

func (d *diContainer) PaymentClient(_ context.Context) (orderSrv.PaymentClient, error) {
	if d.paymentClient == nil {
		paymentConn, err := grpc.NewClient(paymentServiceAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                paymentServiceKeepaliveTime,
				Timeout:             paymentServiceKeepaliveTimeout,
				PermitWithoutStream: true,
			}))
		if err != nil {
			return nil, fmt.Errorf("инициализировать payment client: %w", err)
		}
		closer.Add("payment conn", func(_ context.Context) error {
			return paymentConn.Close()
		})

		client := paymentv1.NewPaymentServiceClient(paymentConn)
		d.paymentClient = paymentV1Client.New(client)
	}
	return d.paymentClient, nil
}

func (d *diContainer) InventoryClient(_ context.Context) (orderSrv.InventoryClient, error) {
	if d.inventoryClient == nil {
		inventoryConn, err := grpc.NewClient(inventoryServiceAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                inventoryServiceKeepaliveTime,
				Timeout:             inventoryServiceKeepaliveTimeout,
				PermitWithoutStream: true,
			}))
		if err != nil {
			return nil, fmt.Errorf("инициализировать inventory client: %w", err)
		}
		closer.Add("inventory conn", func(_ context.Context) error {
			return inventoryConn.Close()
		})

		client := inventoryv1.NewInventoryServiceClient(inventoryConn)
		d.inventoryClient = invetntoryV1Client.New(client)
	}
	return d.inventoryClient, nil
}

func (d *diContainer) OrderService(ctx context.Context) (orderV1API.OrderService, error) {
	if d.orderService == nil {
		repo, err := d.OrderRepo(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать сервис: %w", err)
		}
		paymentClient, err := d.PaymentClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать сервис: %w", err)
		}
		inventoryClient, err := d.InventoryClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать сервис: %w", err)
		}

		orderSrv.NewService(repo, paymentClient, inventoryClient)
	}
	return d.orderService, nil
}

func (d *diContainer) OrderV1API(ctx context.Context) (*orderv1.Server, error) {
	if d.orderV1Handler == nil {
		service, err := d.OrderService(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать хендлер: %w", err)
		}
		api := orderV1API.NewAPI(service)
		d.orderV1Handler, err = orderV1API.SetupServer(api)
		if err != nil {
			return nil, fmt.Errorf("инициализировать хендлер: %w", err)
		}
	}
	return d.orderV1Handler, nil
}
