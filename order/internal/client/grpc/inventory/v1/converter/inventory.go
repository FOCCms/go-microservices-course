package converter

import (
	"github.com/FOCCms/go-microservices-course/order/internal/model"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func PartsToModel(parts []*inventoryv1.Part) []model.Part {
	res := make([]model.Part, 0, len(parts))
	for _, p := range parts {
		res = append(res, PartToModel(p))
	}

	return res
}

func PartToModel(p *inventoryv1.Part) model.Part {
	return model.Part{
		UUID:          p.GetUuid(),
		Name:          p.GetName(),
		Price:         p.GetPrice(),
		StockQuantity: p.GetStockQuantity(),
	}
}
