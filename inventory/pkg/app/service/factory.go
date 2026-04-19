package service

import inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"

// partToProto конвертирует внутреннюю структуру Part в proto-структуру.
func partToProto(part Part) *inventoryv1.Part {
	return &inventoryv1.Part{
		Uuid:          part.UUID,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		PartType:      part.PartType,
		StockQuantity: part.StockQuantity,
		CreatedAt:     part.CreatedAt,
	}
}
