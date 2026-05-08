package valueobject

import (
	"fmt"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
)

type PartType string

const (
	PartTypeUnspecified PartType = "UNSPECIFIED"
	PartTypeHull        PartType = "HULL"
	PartTypeEngine      PartType = "ENGINE"
	PartTypeShield      PartType = "SHIELD"
	PartTypeWeapon      PartType = "WEAPON"
)

func NewPartType(s string) (PartType, error) {
	pt := PartType(s)
	switch pt {
	case PartTypeHull, PartTypeEngine, PartTypeShield, PartTypeWeapon:
		return pt, nil
	default:
		return "", fmt.Errorf("неизвестный тип детали %q: %w", s, errs.ErrInvalidProperties)
	}
}
