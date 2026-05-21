package order_consumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/FOCCms/go-microservices-course/platform/pkg/kafka"
)

func (s *Service) OrderPaidHandler(ctx context.Context, msg kafka.Message) error {
	event, err := decodeOrderPaid(msg.Value)
	if err != nil {
		slog.ErrorContext(ctx, "десериализовать OrderPaid, сообщение будет пропущено",
			slog.Any("error", err),
			slog.Int64("offset", msg.Offset))
		return nil
	}

	err = s.assembleService.Assemble(ctx, event)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		return fmt.Errorf("собрать корабль: %w", err)
	}

	return nil
}
