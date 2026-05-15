package assembly_consumer

import (
	"context"
	"fmt"

	"github.com/FOCCms/go-microservices-course/platform/pkg/kafka"
)

func (s *service) ShipAssembledHandler(ctx context.Context, msg kafka.Message) error {
	event, err := decodeShipAssembled(msg.Value)
	if err != nil {
		return fmt.Errorf("декодировать ShipAssembled: %w", err)
	}

	err = s.orderService.CommitParts(ctx, event)
	if err != nil {
		return fmt.Errorf("списать детали со склада: %w", err)
	}

	return nil
}
