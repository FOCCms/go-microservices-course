package part

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
)

func (s *service) Get(ctx context.Context, id string) (model.Part, error) {
	if err := uuid.Validate(id); err != nil {
		return model.Part{}, fmt.Errorf("получить деталь: %w", errs.ErrInvalidUUID)
	}

	part, err := s.partRepository.Get(ctx, id)
	if err != nil {
		return model.Part{}, fmt.Errorf("получить деталь: %w", err)
	}

	return part, nil
}
