package part

import (
	"context"
	"fmt"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

func (s *service) ValidateCompatibility(ctx context.Context, slots model.ShipSlots) error {
	if err := checkUUIDSNonUnique(slots.List()); err != nil {
		return fmt.Errorf("проверить совместимость деталей: %w", err)
	}

	parts, err := s.List(ctx, valueobject.PartFilter{UUIDs: slots.List()})
	if err != nil {
		return fmt.Errorf("проверить совместимость деталей: %w", err)
	}

	resolved := getResolvedShipSlots(parts, slots)

	if err = checkSlotType(resolved); err != nil {
		return fmt.Errorf("проверить совместимость деталей: %w", err)
	}

	return s.compatibilityChecker.Check(resolved)
}

func getResolvedShipSlots(parts []model.Part, slots model.ShipSlots) model.ResolvedShipSlots {
	partsMap := make(map[string]*model.Part, len(parts))
	for _, p := range parts {
		partsMap[p.UUID()] = &p
	}

	resolved := model.ResolvedShipSlots{
		Hull:   partsMap[slots.Hull],
		Engine: partsMap[slots.Engine],
		Shield: partsMap[slots.Shield],
		Weapon: partsMap[slots.Weapon],
	}
	return resolved
}

func checkUUIDSNonUnique(uuids []string) error {
	m := make(map[string]struct{}, len(uuids))
	for _, s := range uuids {
		if s == "" {
			continue
		}
		if _, exists := m[s]; exists {
			return fmt.Errorf("повторяющиеся uuid деталей: %w", errs.ErrPartIsNonUnique)
		}
		m[s] = struct{}{}
	}
	return nil
}

func checkSlotType(r model.ResolvedShipSlots) error {
	if r.Hull != nil && r.Hull.PartType() != valueobject.PartTypeHull {
		return errs.ErrPartTypeMismatch
	}
	if r.Engine != nil && r.Engine.PartType() != valueobject.PartTypeEngine {
		return errs.ErrPartTypeMismatch
	}
	if r.Shield != nil && r.Shield.PartType() != valueobject.PartTypeShield {
		return errs.ErrPartTypeMismatch
	}
	if r.Weapon != nil && r.Weapon.PartType() != valueobject.PartTypeWeapon {
		return errs.ErrPartTypeMismatch
	}
	return nil
}
