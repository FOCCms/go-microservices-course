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
)

func (s *service) Get(ctx context.Context, id string) (model.Part, error) {
	ctx, span := otel.Tracer("inventory-service").Start(ctx, "inventory.Get")
	defer span.End()

	span.SetAttributes(attribute.String("inventory.part_uuid", id))

	if err := uuid.Validate(id); err != nil {
		span.RecordError(errs.ErrInvalidUUID)
		span.SetStatus(codes.Error, errs.ErrInvalidUUID.Error())
		return model.Part{}, fmt.Errorf("получить деталь: %w", errs.ErrInvalidUUID)
	}

	part, err := s.partRepository.Get(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.Part{}, fmt.Errorf("получить деталь: %w", err)
	}

	span.SetStatus(codes.Ok, "деталь успешно получена")

	return part, nil
}
