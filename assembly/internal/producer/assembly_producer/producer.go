package assembly_producer

import (
	"context"
	"fmt"

	"github.com/FOCCms/go-microservices-course/assembly/internal/model"
	"github.com/FOCCms/go-microservices-course/platform/pkg/kafka"
)

type service struct {
	shipAssembledProducer KafkaProducer
}

func NewService(shipAssembledProducer KafkaProducer) *service {
	return &service{
		shipAssembledProducer: shipAssembledProducer,
	}
}

func (s *service) ProduceShipAssembled(ctx context.Context, event model.ShipAssembledEvent) error {
	payload, err := encodeShipAssembled(event)
	if err != nil {
		return fmt.Errorf("сериализовать ShipAssembled: %w", err)
	}

	return s.shipAssembledProducer.Send(ctx, &kafka.Message{
		Key:   []byte(event.EventUUID.String()),
		Value: payload,
	})
}
