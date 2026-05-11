package valueobject

import (
	"fmt"
	"slices"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
)

type EngineProperties struct {
	class            string
	requiredStrength int
}

func (e *EngineProperties) RequiredStrength() int {
	return e.requiredStrength
}

func NewEngineProperties(class string, requiredStrength int) (PartProperties, error) {
	if !slices.Contains([]string{"A", "B", "C"}, class) {
		return PartProperties{}, fmt.Errorf("невалидный класс двигателя, получено %s: %w", class, errs.ErrInvalidProperties)
	}
	return PartProperties{
		engine: &EngineProperties{
			class:            class,
			requiredStrength: requiredStrength,
		},
	}, nil
}
