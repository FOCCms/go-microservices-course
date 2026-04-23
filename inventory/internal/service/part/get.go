package part

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model"
)

func (s *service) Get(ctx context.Context, idStr string) (model.Part, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return model.Part{}, errs.ErrInvalidUUID
	}

	part, err := s.partRepository.Get(ctx, id)
	if err != nil {
		return model.Part{}, fmt.Errorf("получить деталь: %w", err)
	}

	return part, nil
}
