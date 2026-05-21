package assembly_consumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/FOCCms/go-microservices-course/platform/pkg/kafka"
)

func (s *service) ShipAssembledHandler(ctx context.Context, msg kafka.Message) error {
	event, err := decodeShipAssembled(msg.Value)
	if err != nil {
		slog.ErrorContext(ctx, "десериализовать ShipAssembled, сообщение будет пропущено",
			slog.Any("error", err),
			slog.Int64("offset", msg.Offset))
		return nil
	}

	err = s.orderService.CommitParts(ctx, event)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("списать детали со склада: %w", err)
	}

	return nil
}
