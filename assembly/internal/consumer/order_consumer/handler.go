package order_consumer

import (
	"context"
	"fmt"

	"github.com/FOCCms/go-microservices-course/platform/pkg/kafka"
)

func (s *service) OrderPaidHandler(ctx context.Context, msg kafka.Message) error {
	event, err := decodeOrderPaid(msg.Value)
	if err != nil {
		return fmt.Errorf("декодировать OrderPaid: %w", err)
	}

	err = s.assembleService.Assemble(ctx, event)
	if err != nil {
		return fmt.Errorf("собрать корабль: %w", err)
	}

	return nil
}
