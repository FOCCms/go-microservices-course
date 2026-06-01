package part

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
)

func (s *service) List(ctx context.Context, filter valueobject.PartFilter) ([]model.Part, error) {
	ctx, span := otel.Tracer("inventory-service").Start(ctx, "inventory.List")
	defer span.End()

	span.SetAttributes(
		attribute.StringSlice("inventory.filter_uuids", filter.UUIDs),
		attribute.String("inventory.filter_part_type", string(filter.PartType)),
	)

	// Валидируем параметры фильтра.
	for _, id := range filter.UUIDs {
		if err := uuid.Validate(id); err != nil {
			span.RecordError(errs.ErrInvalidUUID)
			span.SetStatus(codes.Error, errs.ErrInvalidUUID.Error())
			return []model.Part{}, fmt.Errorf("получить детали: %w", errs.ErrInvalidUUID)
		}
	}

	// Получаем список деталей.
	parts, err := s.partRepository.List(ctx, record.PartFilter{
		UUIDs:    filter.UUIDs,
		PartType: string(filter.PartType),
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return []model.Part{}, fmt.Errorf("получить детали: %w", err)
	}

	span.SetAttributes(attribute.Int("inventory.resolved_parts_count", len(parts)))
	span.SetStatus(codes.Ok, "детали успешно получены")

	return parts, nil
}

func (s *service) listForUpdate(ctx context.Context, filter valueobject.PartFilter) ([]model.Part, error) {
	// Валидируем параметры фильтра.
	for _, id := range filter.UUIDs {
		if err := uuid.Validate(id); err != nil {
			return []model.Part{}, fmt.Errorf("получить детали: %w", errs.ErrInvalidUUID)
		}
	}

	// Получаем список деталей.
	parts, err := s.partRepository.ListForUpdate(ctx, record.PartFilter{
		UUIDs:    filter.UUIDs,
		PartType: string(filter.PartType),
	})
	if err != nil {
		return []model.Part{}, fmt.Errorf("получить детали: %w", err)
	}

	return parts, nil
}
