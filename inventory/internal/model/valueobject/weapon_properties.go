package valueobject

import (
	"fmt"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
)

const (
	WeaponTypeLaser   = "laser"
	WeaponTypeMissile = "missile"
)

type WeaponType string

type WeaponProperties struct {
	weaponType WeaponType
}

func NewWeaponProperties(weaponType string) (PartProperties, error) {
	switch weaponType {
	case WeaponTypeMissile, WeaponTypeLaser:
		return PartProperties{
			weapon: &WeaponProperties{
				weaponType: WeaponType(weaponType),
			},
		}, nil
	default:
		return PartProperties{}, fmt.Errorf("неизвестный тип оружия %q: %w", weaponType, errs.ErrInvalidProperties)
	}
}
