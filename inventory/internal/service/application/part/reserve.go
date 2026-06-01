package part

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

func (s *service) Reserve(ctx context.Context, uuids []string) error {
	ctx, span := otel.Tracer("inventory-service").Start(ctx, "inventory.Reserve")
	defer span.End()

	span.SetAttributes(
		attribute.Int("inventory.requested_count", len(uuids)),
	)

	if len(uuids) <= 10 {
		span.SetAttributes(attribute.StringSlice("inventory.reserve_uuids", uuids))
	}

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		parts, err := s.listForUpdate(ctx, valueobject.PartFilter{
			UUIDs: uuids,
		})
		if err != nil {
			return fmt.Errorf("зарезервировать детали: %w", err)
		}

		for i := range parts {
			if err = parts[i].Reserve(); err != nil {
				return fmt.Errorf("зарезервировать детали (деталь %s): %w", parts[i].UUID(), errs.ErrOutOfStock)
			}
		}

		err = s.partRepository.UpdateReservationsBatch(ctx, parts)
		if err != nil {
			return fmt.Errorf("зарезервировать детали: %w", err)
		}

		slog.Info("детали успешно зарезервированы",
			slog.Int("parts_count", len(parts)),
			slog.Any("part_uuids", uuids),
		)

		span.SetStatus(codes.Ok, "детали успешно зарезервированы")
		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}
