package app

import (
	"context"
	"fmt"

	"github.com/FOCCms/go-microservices-course/assembly/internal/config"
	assemblySrv "github.com/FOCCms/go-microservices-course/assembly/internal/service/assembly"

	orderConsumer "github.com/FOCCms/go-microservices-course/assembly/internal/consumer/order_consumer"
	assemblyProsucer "github.com/FOCCms/go-microservices-course/assembly/internal/producer/assembly_producer"
	"github.com/FOCCms/go-microservices-course/platform/pkg/closer"
	wrappedKafkaProducer "github.com/FOCCms/go-microservices-course/platform/pkg/kafka/producer"

	wrappedKafkaConsumer "github.com/FOCCms/go-microservices-course/platform/pkg/kafka/consumer"
	kafkaMiddleware "github.com/FOCCms/go-microservices-course/platform/pkg/middleware/kafka"
	"github.com/IBM/sarama"
)

type ConsumerService interface {
	RunConsumer(ctx context.Context) error
}

type diContainer struct {
	orderConsumerService ConsumerService

	syncProducer  sarama.SyncProducer
	consumerGroup sarama.ConsumerGroup

	orderPaidConsumer     *wrappedKafkaConsumer.Consumer
	shipAssembledProducer *wrappedKafkaProducer.Producer

	assemblyService         orderConsumer.AssembleService
	assemblyProducerService assemblySrv.AssemblyProducerService
}

func (d *diContainer) ConsumerGroup() (sarama.ConsumerGroup, error) {
	if d.consumerGroup == nil {
		consumerGroup, err := sarama.NewConsumerGroup(
			config.AppConfig().Kafka.Brokers,
			config.AppConfig().OrderPaidConsumer.GroupID(),
			config.AppConfig().OrderPaidConsumer.SaramaConfig(),
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

func (d *diContainer) OrderPaidConsumer() (*wrappedKafkaConsumer.Consumer, error) {
	if d.orderPaidConsumer == nil {
		group, err := d.ConsumerGroup()
		if err != nil {
			return nil, fmt.Errorf("инициализировать OrderPaidConsumer: %w", err)
		}
		d.orderPaidConsumer = wrappedKafkaConsumer.NewConsumer(
			group,
			[]string{
				config.AppConfig().OrderPaidConsumer.Topic(),
			},
			wrappedKafkaConsumer.WithMiddlewares(
				kafkaMiddleware.ConsumerLogging(),
			),
		)
	}

	return d.orderPaidConsumer, nil
}

func (d *diContainer) SyncProducer() (sarama.SyncProducer, error) {
	if d.syncProducer == nil {
		p, err := sarama.NewSyncProducer(
			config.AppConfig().Kafka.Brokers,
			config.AppConfig().ShipAssembledProducer.SaramaConfig(),
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

func (d *diContainer) ShipAssembledProducer() (*wrappedKafkaProducer.Producer, error) {
	if d.shipAssembledProducer == nil {
		p, err := d.SyncProducer()
		if err != nil {
			return nil, fmt.Errorf("инициализировать ShipAssembledProducer: %w", err)
		}
		d.shipAssembledProducer = wrappedKafkaProducer.NewProducer(
			p,
			config.AppConfig().ShipAssembledProducer.Topic(),
		)
	}

	return d.shipAssembledProducer, nil
}

func (d *diContainer) ShipProducerService() (assemblySrv.AssemblyProducerService, error) {
	if d.assemblyProducerService == nil {
		p, err := d.ShipAssembledProducer()
		if err != nil {
			return nil, fmt.Errorf("инициализировать ShipProducerService: %w", err)
		}
		d.assemblyProducerService = assemblyProsucer.NewService(p)
	}

	return d.assemblyProducerService, nil
}

func (d *diContainer) AssemblyService() (orderConsumer.AssembleService, error) {
	if d.assemblyService == nil {
		producer, err := d.ShipProducerService()
		if err != nil {
			return nil, fmt.Errorf("инициализировать AssemblyService: %w", err)
		}
		d.assemblyService = assemblySrv.NewService(producer)
	}
	return d.assemblyService, nil
}

func (d *diContainer) OrderConsumerService() (ConsumerService, error) {
	if d.orderConsumerService == nil {
		orderPaidConsumer, err := d.OrderPaidConsumer()
		if err != nil {
			return nil, fmt.Errorf("инициализировать OrderConsumerService: %w", err)
		}
		assemblyService, err := d.AssemblyService()
		if err != nil {
			return nil, fmt.Errorf("инициализировать OrderConsumerService: %w", err)
		}
		d.orderConsumerService = orderConsumer.NewService(orderPaidConsumer, assemblyService)
	}
	return d.orderConsumerService, nil
}
