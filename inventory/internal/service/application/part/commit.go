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

func (s *service) Commit(ctx context.Context, uuids []string) error {
	ctx, span := otel.Tracer("inventory-service").Start(ctx, "inventory.Commit")
	defer span.End()

	span.SetAttributes(
		attribute.StringSlice("inventory.commit_uuids", uuids),
		attribute.Int("inventory.requested_count", len(uuids)),
	)

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		_, err := s.listForUpdate(ctx, valueobject.PartFilter{
			UUIDs: uuids,
		})
		if err != nil {
			return fmt.Errorf("списать детали: %w", err)
		}

		err = s.partRepository.Commit(ctx, uuids)
		if err != nil {
			return fmt.Errorf("списать детали: %w", err)
		}

		slog.Info("списание деталей со склада завершено",
			slog.Int("parts_count", len(uuids)), // Кол-во деталей
			slog.Any("part_uuids", uuids),       // Список деталей
		)
		span.SetStatus(codes.Ok, "детали успешно списаны")

		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}
