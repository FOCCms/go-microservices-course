package v1

import (
	"context"

	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

type PartService interface {
	Get(ctx context.Context, uuid string) (model.Part, error)
	List(ctx context.Context, filter valueobject.PartFilter) ([]model.Part, error)
	ValidateCompatibility(ctx context.Context, slots model.ShipSlots) error
	Reserve(ctx context.Context, uuids []string) error
	Release(ctx context.Context, uuids []string) error
}
