package converter

import (
	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
	"github.com/FOCCms/go-microservices-course/order/internal/repository/record"
)

func OrderToModel(order record.Order, orderItems []record.OrderItem) model.Order {
	o := model.Order{
		UUID:            uuid.MustParse(order.UUID),
		TotalPrice:      order.TotalPrice,
		TransactionUUID: parseOptionalUUID(order.TransactionUUID),
		PaymentMethod:   toModelPaymentMethod(order.PaymentMethod),
		Status:          model.OrderStatus(order.Status),
		CreatedAt:       order.CreatedAt,
	}

	for _, item := range orderItems {
		switch item.PartType {
		case "HULL":
			o.HullUUID = uuid.MustParse(item.PartUUID)
		case "ENGINE":
			o.EngineUUID = uuid.MustParse(item.PartUUID)
		case "SHIELD":
			o.ShieldUUID = parseOptionalUUID(&item.PartUUID)
		case "WEAPON":
			o.WeaponUUID = parseOptionalUUID(&item.PartUUID)
		}
	}

	return o
}

func toModelPaymentMethod(s *string) *model.PaymentMethod {
	if s == nil {
		return nil
	}
	return new(model.PaymentMethod(*s))
}

func parseOptionalUUID(s *string) *uuid.UUID {
	if s == nil {
		return nil
	}
	return new(uuid.MustParse(*s))
}

func ModelOrderToRecordOrder(order model.Order) record.Order {
	return record.Order{
		UUID:            order.UUID.String(),
		TotalPrice:      order.TotalPrice,
		TransactionUUID: uuidToOptionalString(order.TransactionUUID),
		PaymentMethod:   toRecordPaymentMethod(order.PaymentMethod),
		Status:          string(order.Status),
		CreatedAt:       order.CreatedAt,
	}
}

func uuidToOptionalString(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	return new(u.String())
}

func toRecordPaymentMethod(p *model.PaymentMethod) *string {
	if p == nil {
		return nil
	}
	return new(string(*p))
}
