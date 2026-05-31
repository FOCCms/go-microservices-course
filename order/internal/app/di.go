package app

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	orderV1API "github.com/FOCCms/go-microservices-course/order/internal/api/order/v1"
	iamV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/iam/v1"
	invetntoryV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/inventory/v1"
	paymentV1Client "github.com/FOCCms/go-microservices-course/order/internal/client/grpc/payment/v1"
	"github.com/FOCCms/go-microservices-course/order/internal/config"
	assemblyConsumer "github.com/FOCCms/go-microservices-course/order/internal/consumer/assembly_consumer"
	"github.com/FOCCms/go-microservices-course/order/internal/interceptor"
	"github.com/FOCCms/go-microservices-course/order/internal/middleware"
	orderProducer "github.com/FOCCms/go-microservices-course/order/internal/producer/order_producer"
	orderRepository "github.com/FOCCms/go-microservices-course/order/internal/repository/order"
	orderSrv "github.com/FOCCms/go-microservices-course/order/internal/service/order"
	"github.com/FOCCms/go-microservices-course/platform/pkg/closer"
	wrappedKafkaConsumer "github.com/FOCCms/go-microservices-course/platform/pkg/kafka/consumer"
	wrappedKafkaProducer "github.com/FOCCms/go-microservices-course/platform/pkg/kafka/producer"
	kafkaMiddleware "github.com/FOCCms/go-microservices-course/platform/pkg/middleware/kafka"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
	authv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/auth/v1"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

type diContainer struct {
	pgPool    *pgxpool.Pool
	txManager orderSrv.TxManager

	orderRepo orderSrv.OrderRepository

	inventoryClient orderSrv.InventoryClient
	paymentClient   orderSrv.PaymentClient
	iamClient       middleware.IAMClient

	orderService *orderSrv.Service

	orderV1Handler *orderv1.Server

	syncProducer  sarama.SyncProducer
	consumerGroup sarama.ConsumerGroup

	orderPaidProducer     *wrappedKafkaProducer.Producer
	shipAssembledConsumer *wrappedKafkaConsumer.Consumer

	orderProducerService    orderSrv.OrderProducerService
	assemblyConsumerService ConsumerService
}

type ConsumerService interface {
	RunConsumer(ctx context.Context) error
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
		cfg := config.AppConfig().PaymentClient
		conn, err := d.newGRPCConnection(
			"payment",
			cfg.Address,
			cfg.KeepaliveTime,
			cfg.KeepaliveTimeout,
			cfg.PermitWithoutStream,
		)
		if err != nil {
			return nil, fmt.Errorf("инициализировать payment client: %w", err)
		}

		client := paymentv1.NewPaymentServiceClient(conn)
		d.paymentClient = paymentV1Client.New(client)
	}
	return d.paymentClient, nil
}

func (d *diContainer) InventoryClient(_ context.Context) (orderSrv.InventoryClient, error) {
	if d.inventoryClient == nil {
		inventoryConn, err := grpc.NewClient(config.AppConfig().InventoryClient.Address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                config.AppConfig().InventoryClient.KeepaliveTime,
				Timeout:             config.AppConfig().InventoryClient.KeepaliveTimeout,
				PermitWithoutStream: config.AppConfig().InventoryClient.PermitWithoutStream,
			}),
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
			grpc.WithChainUnaryInterceptor(
				interceptor.AuthOutgoingInterceptor()))
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

func (d *diContainer) IAMClient(_ context.Context) (middleware.IAMClient, error) {
	if d.iamClient == nil {
		cfg := config.AppConfig().IAMClient

		conn, err := d.newGRPCConnection(
			"iam",
			cfg.Address,
			cfg.KeepaliveTime,
			cfg.KeepaliveTimeout,
			cfg.PermitWithoutStream,
		)
		if err != nil {
			return nil, fmt.Errorf("инициализировать iam client: %w", err)
		}

		client := authv1.NewAuthServiceClient(conn)
		d.iamClient = iamV1Client.New(client)
	}
	return d.iamClient, nil
}

func (d *diContainer) newGRPCConnection(
	name string,
	address string,
	keepaliveTime time.Duration,
	keepaliveTimeout time.Duration,
	permitWithoutStream bool,
) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                keepaliveTime,
			Timeout:             keepaliveTimeout,
			PermitWithoutStream: permitWithoutStream,
		}),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, fmt.Errorf("инициализировать %s conn: %w", name, err)
	}

	closer.Add(name+" conn", func(_ context.Context) error {
		return conn.Close()
	})

	return conn, nil
}

func (d *diContainer) SyncProducer() (sarama.SyncProducer, error) {
	if d.syncProducer == nil {
		p, err := sarama.NewSyncProducer(
			config.AppConfig().Kafka.Brokers,
			config.AppConfig().OrderPaidProducer.SaramaConfig(),
		)
		if err != nil {
			return nil, fmt.Errorf("инициализировать SyncProducer: %w", err)
		}

		closer.Add("Kafka sync producer", func(_ context.Context) error {
			return p.Close()
		})
		d.syncProducer = p
	}

	return d.syncProducer, nil
}

func (d *diContainer) OrderPaidProducer() (*wrappedKafkaProducer.Producer, error) {
	if d.orderPaidProducer == nil {
		p, err := d.SyncProducer()
		if err != nil {
			return nil, fmt.Errorf("инициализировать OrderPaidProducer: %w", err)
		}
		d.orderPaidProducer = wrappedKafkaProducer.NewProducer(p, config.AppConfig().OrderPaidProducer.TopicName)
	}
	return d.orderPaidProducer, nil
}

func (d *diContainer) OrderProducerService() (orderSrv.OrderProducerService, error) {
	if d.orderProducerService == nil {
		p, err := d.OrderPaidProducer()
		if err != nil {
			return nil, fmt.Errorf("инициализировать OrderProducerService: %w", err)
		}
		d.orderProducerService = orderProducer.NewService(p)
	}
	return d.orderProducerService, nil
}

func (d *diContainer) OrderService(ctx context.Context) (*orderSrv.Service, error) {
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
		txManager, err := d.TxManager(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать сервис: %w", err)
		}
		orderProducerService, err := d.OrderProducerService()
		if err != nil {
			return nil, fmt.Errorf("инициализировать сервис: %w", err)
		}

		d.orderService = orderSrv.NewService(repo, paymentClient, inventoryClient, txManager, orderProducerService)
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
		d.orderV1Handler, err = orderv1.NewServer(api, orderv1.WithErrorHandler(orderV1API.ErrorHandler))
		if err != nil {
			return nil, fmt.Errorf("инициализировать хендлер: %w", err)
		}
	}
	return d.orderV1Handler, nil
}

func (d *diContainer) ConsumerGroup() (sarama.ConsumerGroup, error) {
	if d.consumerGroup == nil {
		consumerGroup, err := sarama.NewConsumerGroup(
			config.AppConfig().Kafka.Brokers,
			config.AppConfig().ShipAssembledConsumer.GroupID(),
			config.AppConfig().ShipAssembledConsumer.SaramaConfig(),
		)
		if err != nil {
			return nil, fmt.Errorf("создать consumer group: %w", err)
		}

		closer.Add("Kafka consumer group", func(_ context.Context) error {
			return consumerGroup.Close()
		})

		d.consumerGroup = consumerGroup
	}

	return d.consumerGroup, nil
}

func (d *diContainer) ShipAssembledConsumer() (assemblyConsumer.Consumer, error) {
	if d.shipAssembledConsumer == nil {
		group, err := d.ConsumerGroup()
		if err != nil {
			return nil, fmt.Errorf("инициализировать ShipAssembledConsumer: %w", err)
		}
		d.shipAssembledConsumer = wrappedKafkaConsumer.NewConsumer(
			group,
			[]string{
				config.AppConfig().ShipAssembledConsumer.Topic(),
			},
			wrappedKafkaConsumer.WithMiddlewares(
				kafkaMiddleware.ConsumerLogging(),
				kafkaMiddleware.ConsumerSession(),
			),
		)
	}

	return d.shipAssembledConsumer, nil
}

func (d *diContainer) AssemblyConsumerService(ctx context.Context) (ConsumerService, error) {
	if d.assemblyConsumerService == nil {
		shipAssembledConsumer, err := d.ShipAssembledConsumer()
		if err != nil {
			return nil, fmt.Errorf("инициализировать AssemblyConsumerService: %w", err)
		}
		orderService, err := d.OrderService(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать AssemblyConsumerService: %w", err)
		}
		d.assemblyConsumerService = assemblyConsumer.NewService(shipAssembledConsumer, orderService)
	}

	return d.assemblyConsumerService, nil
}
