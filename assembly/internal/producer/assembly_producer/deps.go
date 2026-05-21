package assembly_producer

import (
	"context"

	"github.com/FOCCms/go-microservices-course/platform/pkg/kafka"
)

type KafkaProducer interface {
	Send(ctx context.Context, msg *kafka.Message) error
}
