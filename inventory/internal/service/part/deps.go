package part

import (
	"context"

	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/inventory/internal/model"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
)

type PartRepository interface {
	Get(ctx context.Context, id uuid.UUID) (model.Part, error)
	List(ctx context.Context, filter record.PartFilter) ([]model.Part, error)
}
