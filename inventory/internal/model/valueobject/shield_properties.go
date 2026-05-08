package valueobject

import (
	"fmt"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
)

const (
	ShieldTypeEnergy = "energy"
	ShieldTypePlasma = "plasma"
)

type ShieldType string

type ShieldProperties struct {
	shieldType ShieldType
}

func (s *ShieldProperties) ConflictsWith(weapon *WeaponProperties) bool {
	return s.shieldType == ShieldTypePlasma && weapon.weaponType == WeaponTypeLaser
}

func NewShieldProperties(shieldType string) (PartProperties, error) {
	switch shieldType {
	case ShieldTypeEnergy, ShieldTypePlasma:
		return PartProperties{
			shield: &ShieldProperties{
				shieldType: ShieldType(shieldType),
			},
		}, nil
	default:
		return PartProperties{}, fmt.Errorf("неизвестный тип щита %q: %w", shieldType, errs.ErrInvalidProperties)
	}
}
