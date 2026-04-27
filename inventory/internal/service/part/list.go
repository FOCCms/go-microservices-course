package part

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
)

func (s *service) List(ctx context.Context, filter model.PartFilter) ([]model.Part, error) {
	// Парсим параметры фильтра.
	ids := make([]uuid.UUID, 0, len(filter.UUIDs))

	for _, idStr := range filter.UUIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return []model.Part{}, errs.ErrInvalidUUID
		}

		ids = append(ids, id)
	}

	// Получаем список деталей.
	parts, err := s.partRepository.List(ctx, record.PartFilter{
		UUIDs:    ids,
		PartType: string(filter.PartType),
	})
	if err != nil {
		return []model.Part{}, fmt.Errorf("получить детали: %w", err)
	}

	// Если фильтр по PartType -- сортируем
	if len(filter.UUIDs) == 0 {
		slices.SortFunc(parts, func(a, b model.Part) int {
			return cmp.Compare(a.Name, b.Name)
		})
	}

	return parts, nil
}
