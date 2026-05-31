package part

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

func (s *service) ValidateCompatibility(ctx context.Context, slots model.ShipSlots) error {
	ctx, span := otel.Tracer("inventory-service").Start(ctx, "inventory.ValidateCompatibility")
	defer span.End()

	span.SetAttributes(
		attribute.String("ship.hull_uuid", slots.Hull),
		attribute.String("ship.engine_uuid", slots.Engine),
		attribute.String("ship.shield_uuid", slots.Shield),
		attribute.String("ship.weapon_uuid", slots.Weapon),
	)

	if err := checkUUIDSNonUnique(slots.List()); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "повторяющиеся uuid деталей")
		return fmt.Errorf("проверить совместимость деталей: %w", err)
	}

	parts, err := s.List(ctx, valueobject.PartFilter{UUIDs: slots.List()})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("проверить совместимость деталей: %w", err)
	}

	resolved := getResolvedShipSlots(parts, slots)

	if err = checkSlotType(resolved); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "детали несовместимы со слотами")
		return fmt.Errorf("проверить совместимость деталей: %w", err)
	}

	if err = s.compatibilityChecker.Check(resolved); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	span.SetStatus(codes.Ok, "конфигурация корабля валидна и совместима")

	return nil
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
