package part

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
)

func (s *service) List(ctx context.Context, filter model.PartFilter) ([]model.Part, error) {
	// Валидируем параметры фильтра.
	for _, id := range filter.UUIDs {
		if err := uuid.Validate(id); err != nil {
			return []model.Part{}, fmt.Errorf("получить детали: %w", errs.ErrInvalidUUID)
		}
	}

	// Получаем список деталей.
	parts, err := s.partRepository.List(ctx, record.PartFilter{
		UUIDs:    filter.UUIDs,
		PartType: string(filter.PartType),
	})
	if err != nil {
		return []model.Part{}, fmt.Errorf("получить детали: %w", err)
	}

	return parts, nil
}
