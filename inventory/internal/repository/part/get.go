package part

import (
	"context"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model"
	repoConverter "github.com/FOCCms/go-microservices-course/inventory/internal/repository/converter"
)

func (r *repository) Get(_ context.Context, id uuid.UUID) (model.Part, error) {
	r.mu.RLock()
	part, ok := r.parts[id]
	r.mu.RUnlock()

	if !ok {
		return model.Part{}, errs.ErrPartNotFound
	}

	return repoConverter.PartToModel(part), nil
}
