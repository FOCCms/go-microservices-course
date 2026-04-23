package v1

import (
	"context"

	"github.com/FOCCms/go-microservices-course/inventory/internal/model"
)

type PartService interface {
	Get(ctx context.Context, uuid string) (model.Part, error)
	List(ctx context.Context, filter model.PartFilter) ([]model.Part, error)
}
