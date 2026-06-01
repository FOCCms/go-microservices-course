package part

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

func (s *service) Release(ctx context.Context, uuids []string) error {
	ctx, span := otel.Tracer("inventory-service").Start(ctx, "inventory.Release")
	defer span.End()

	span.SetAttributes(
		attribute.Int("inventory.requested_count", len(uuids)),
	)

	if len(uuids) <= 10 {
		span.SetAttributes(attribute.StringSlice("inventory.release_uuids", uuids))
	}

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		parts, err := s.listForUpdate(ctx, valueobject.PartFilter{
			UUIDs: uuids,
		})
		if err != nil {
			return fmt.Errorf("освободить детали: %w", err)
		}

		for i := range parts {
			if err = parts[i].Release(); err != nil {
				span.SetAttributes(attribute.String("inventory.failed_part_uuid", parts[i].UUID()))
				return fmt.Errorf("освободить детали: %w", err)
			}
		}

		err = s.partRepository.UpdateReservationsBatch(ctx, parts)
		if err != nil {
			return fmt.Errorf("освободить детали: %w", err)
		}

		slog.InfoContext(ctx, "детали успешно освобождены",
			slog.Int("parts_count", len(parts)), // сколько объектов обновили
			slog.Any("part_uuids", uuids),       // какие UUID освободились
		)

		span.SetStatus(codes.Ok, "детали успешно освобождены")

		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}
