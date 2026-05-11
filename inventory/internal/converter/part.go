package converter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func PartToProto(part model.Part) *inventoryv1.Part {
	return &inventoryv1.Part{
		Uuid:          part.UUID(),
		Name:          part.Name(),
		Description:   part.Description(),
		Price:         part.Price(),
		PartType:      PartTypeToProtoPartType(part.PartType()),
		StockQuantity: int64(part.StockQuantity()),
		CreatedAt:     timestamppb.New(part.CreatedAt()),
	}
}

func PartsToProto(parts []model.Part) []*inventoryv1.Part {
	res := make([]*inventoryv1.Part, 0, len(parts))
	for _, p := range parts {
		res = append(res, PartToProto(p))
	}
	return res
}

func PartTypeToProtoPartType(t valueobject.PartType) inventoryv1.PartType {
	switch t {
	case valueobject.PartTypeHull:
		return inventoryv1.PartType_PART_TYPE_HULL
	case valueobject.PartTypeEngine:
		return inventoryv1.PartType_PART_TYPE_ENGINE
	case valueobject.PartTypeShield:
		return inventoryv1.PartType_PART_TYPE_SHIELD
	case valueobject.PartTypeWeapon:
		return inventoryv1.PartType_PART_TYPE_WEAPON
	default:
		return inventoryv1.PartType_PART_TYPE_UNSPECIFIED
	}
}

func ProtoPartTypeToPartType(t inventoryv1.PartType) valueobject.PartType {
	switch t {
	case inventoryv1.PartType_PART_TYPE_HULL:
		return valueobject.PartTypeHull
	case inventoryv1.PartType_PART_TYPE_ENGINE:
		return valueobject.PartTypeEngine
	case inventoryv1.PartType_PART_TYPE_SHIELD:
		return valueobject.PartTypeShield
	case inventoryv1.PartType_PART_TYPE_WEAPON:
		return valueobject.PartTypeWeapon
	default:
		return valueobject.PartTypeUnspecified
	}
}
