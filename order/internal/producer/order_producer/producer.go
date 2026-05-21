package order_producer

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
	"github.com/FOCCms/go-microservices-course/platform/pkg/kafka"
	eventsv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/events/v1"
)

type service struct {
	orderPaidProducer KafkaProducer
}

func NewService(orderPaidProducer KafkaProducer) *service {
	return &service{
		orderPaidProducer: orderPaidProducer,
	}
}

func (s *service) ProduceOrderPaid(ctx context.Context, e model.OrderPaidEvent) error {
	msg := &eventsv1.OrderPaid{
		EventUuid: e.EventUUID.String(),
		OrderUuid: e.OrderUUID.String(),
		UserUuid:  e.UserUUID.String(),
	}

	payload, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("сериализовать OrderPaid: %w", err)
	}

	return s.orderPaidProducer.Send(ctx, &kafka.Message{
		Key:   []byte(e.EventUUID.String()),
		Value: payload,
	})
}
