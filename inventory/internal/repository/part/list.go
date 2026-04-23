package part

import (
	"context"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/converter"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
)

func (r *repository) List(ctx context.Context, filter record.PartFilter) ([]model.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(filter.UUIDs) > 0 {
		parts := make([]model.Part, 0, len(filter.UUIDs))

		for _, id := range filter.UUIDs {
			part, ok := r.parts[id]

			if !ok {
				return []model.Part{}, errs.ErrPartNotFound
			}

			parts = append(parts, converter.PartToModel(part))
		}

		return parts, nil
	} else {
		parts := make([]model.Part, 0, len(r.parts))

		for _, part := range r.parts {
			if filter.PartType == "UNSPECIFIED" || part.PartType == filter.PartType {
				parts = append(parts, converter.PartToModel(part))
			}
		}
		return parts, nil
	}
}
