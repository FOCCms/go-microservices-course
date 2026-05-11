package valueobject

import (
	"fmt"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
)

type HullProperties struct {
	strength int
}

func (h *HullProperties) CanSupport(e *EngineProperties) bool {
	return h.strength >= e.requiredStrength
}

func NewHullProperties(strength int) (PartProperties, error) {
	if strength < 30 || strength > 200 {
		return PartProperties{}, fmt.Errorf("прочность корпуса должна быть от 30 до 200, получено %d: %w", strength, errs.ErrInvalidProperties)
	}
	return PartProperties{
		hull: &HullProperties{strength: strength},
	}, nil
}
