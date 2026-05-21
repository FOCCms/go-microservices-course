package assembly

import (
	"context"

	"github.com/FOCCms/go-microservices-course/assembly/internal/model"
)

type AssemblyProducerService interface {
	ProduceShipAssembled(ctx context.Context, event model.ShipAssembledEvent) error
}
