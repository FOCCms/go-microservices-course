package domain

import (
	"fmt"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	model "github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

type partsSet struct {
	hull   *valueobject.HullProperties
	engine *valueobject.EngineProperties
	shield *valueobject.ShieldProperties
	weapon *valueobject.WeaponProperties
}

type compatibilityChecker struct{}

func NewCompatibilityChecker() *compatibilityChecker {
	return &compatibilityChecker{}
}

func (c *compatibilityChecker) Check(slots model.ResolvedShipSlots) error {
	if slots.Count() <= 1 {
		// нет необходимости проверять только одну деталь на совместимость
		return nil
	}

	if err := checkPartTypesUnique(slots.List()); err != nil {
		return err
	}

	set := extractParts(slots)

	if err := checkHullAndEngineCompatibility(set.hull, set.engine); err != nil {
		return err
	}

	if err := checkShieldAndWeaponCompatibility(set.shield, set.weapon); err != nil {
		return err
	}

	return nil
}

func checkHullAndEngineCompatibility(hull *valueobject.HullProperties, engine *valueobject.EngineProperties) error {
	if hull == nil || engine == nil {
		return nil
	}
	if !hull.CanSupport(engine) {
		return fmt.Errorf("двигатель и корпус несовместимы: %w", errs.ErrIncompatibleParts)
	}
	return nil
}

func checkShieldAndWeaponCompatibility(shield *valueobject.ShieldProperties, weapon *valueobject.WeaponProperties) error {
	if shield == nil || weapon == nil {
		return nil
	}
	if shield.ConflictsWith(weapon) {
		return fmt.Errorf("щит и оружие несовместимы: %w", errs.ErrIncompatibleParts)
	}
	return nil
}

func checkPartTypesUnique(parts []*model.Part) error {
	m := make(map[valueobject.PartType]struct{})
	for _, p := range parts {
		if p == nil {
			continue
		}
		if _, exists := m[p.PartType()]; exists {
			return fmt.Errorf("детали одного типа: %w", errs.ErrIncompatibleParts)
		}
		m[p.PartType()] = struct{}{}
	}
	return nil
}

func extractParts(slots model.ResolvedShipSlots) partsSet {
	var set partsSet
	if slots.Hull != nil {
		p := slots.Hull.Properties()
		set.hull = p.Hull()
	}
	if slots.Engine != nil {
		p := slots.Engine.Properties()
		set.engine = p.Engine()
	}
	if slots.Shield != nil {
		p := slots.Shield.Properties()
		set.shield = p.Shield()
	}
	if slots.Weapon != nil {
		p := slots.Weapon.Properties()
		set.weapon = p.Weapon()
	}
	return set
}
