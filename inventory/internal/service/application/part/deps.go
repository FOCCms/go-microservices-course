package part

import (
	"context"

	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
)

type PartRepository interface {
	Get(ctx context.Context, id string) (model.Part, error)
	List(ctx context.Context, filter record.PartFilter) ([]model.Part, error)
	UpdateReservedBatch(ctx context.Context, parts []model.Part) error
}

type CompatibilityChecker interface {
	Check(slots model.ResolvedShipSlots) error
}
