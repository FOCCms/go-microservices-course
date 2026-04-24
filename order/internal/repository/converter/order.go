package converter

import (
	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
	"github.com/FOCCms/go-microservices-course/order/internal/repository/record"
)

func OrderToModel(order record.Order) model.Order {
	return model.Order{
		UUID:            uuid.MustParse(order.UUID),
		HullUUID:        uuid.MustParse(order.HullUUID),
		EngineUUID:      uuid.MustParse(order.EngineUUID),
		ShieldUUID:      parseOptionalUUID(order.ShieldUUID),
		WeaponUUID:      parseOptionalUUID(order.WeaponUUID),
		TotalPrice:      order.TotalPrice,
		TransactionUUID: parseOptionalUUID(order.TransactionUUID),
		PaymentMethod:   toModelPaymentMethod(order.PaymentMethod),
		Status:          model.OrderStatus(order.Status),
		CreatedAt:       order.CreatedAt,
	}
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
		UUID:            order.HullUUID.String(),
		HullUUID:        order.HullUUID.String(),
		EngineUUID:      order.EngineUUID.String(),
		ShieldUUID:      uuidToOptionalString(order.ShieldUUID),
		WeaponUUID:      uuidToOptionalString(order.WeaponUUID),
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
