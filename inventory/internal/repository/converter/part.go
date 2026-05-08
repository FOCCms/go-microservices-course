package converter

import (
	"encoding/json"
	"fmt"

	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
)

func PartRecordToModel(rec record.Part) (model.Part, error) {
	var propsRec record.PartPropertiesRecord
	if err := json.Unmarshal(rec.Properties, &propsRec); err != nil {
		return model.Part{}, fmt.Errorf("десериализовать свойства: %w", err)
	}

	props, err := partPropertiesFromRecord(propsRec)
	if err != nil {
		return model.Part{}, fmt.Errorf("конвертировать свойства: %w", err)
	}

	partType, err := valueobject.NewPartType(rec.PartType)
	if err != nil {
		return model.Part{}, fmt.Errorf("конвертировать тип детали: %w", err)
	}

	return model.RestorePart(
		rec.UUID,
		rec.Name,
		rec.Description,
		partType,
		rec.Price,
		rec.StockQuantity,
		rec.Reserved,
		props,
		rec.CreatedAt,
	), nil
}

func partPropertiesFromRecord(rec record.PartPropertiesRecord) (valueobject.PartProperties, error) {
	switch {
	case rec.Hull != nil:
		return valueobject.NewHullProperties(rec.Hull.Strength)
	case rec.Engine != nil:
		return valueobject.NewEngineProperties(rec.Engine.Class, rec.Engine.RequiredStrength)
	case rec.Shield != nil:
		return valueobject.NewShieldProperties(rec.Shield.ShieldType)
	case rec.Weapon != nil:
		return valueobject.NewWeaponProperties(rec.Weapon.WeaponType)
	default:
		return valueobject.PartProperties{}, nil
	}
}
