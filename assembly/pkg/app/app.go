package app

import (
	"context"

	"github.com/IBM/sarama"

	orderConsumer "github.com/FOCCms/go-microservices-course/assembly/internal/consumer/order_consumer"
	assemblyProducer "github.com/FOCCms/go-microservices-course/assembly/internal/producer/assembly_producer"
	assemblySrv "github.com/FOCCms/go-microservices-course/assembly/internal/service/assembly"
	wrappedKafkaConsumer "github.com/FOCCms/go-microservices-course/platform/pkg/kafka/consumer"
	wrappedKafkaProducer "github.com/FOCCms/go-microservices-course/platform/pkg/kafka/producer"
	kafkaMiddleware "github.com/FOCCms/go-microservices-course/platform/pkg/middleware/kafka"
)

type Config struct {
	OrderPaidTopic     string
	ShipAssembledTopic string
	MinBuildTimeSec    int
	MaxBuildTimeSec    int
}

type App struct {
	orderConsumerService *orderConsumer.Service
}

func New(producer sarama.SyncProducer, cg sarama.ConsumerGroup, cfg Config) *App {
	orderPaidConsumer := wrappedKafkaConsumer.NewConsumer(
		cg,
		[]string{
			cfg.OrderPaidTopic,
		},
		wrappedKafkaConsumer.WithMiddlewares(
			kafkaMiddleware.ConsumerLogging(),
		),
	)

	shipAssembledProducer := wrappedKafkaProducer.NewProducer(
		producer,
		cfg.ShipAssembledTopic,
	)

	p := assemblyProducer.NewService(shipAssembledProducer)

	srv := assemblySrv.NewService(p)

	orderConsumerService := orderConsumer.NewService(orderPaidConsumer, srv)

	return &App{
		orderConsumerService: orderConsumerService,
	}
}

func (a *App) RunConsumer(ctx context.Context) error {
	return a.orderConsumerService.RunConsumer(ctx) // Реализуй вызов своего внутреннего сервиса
}
