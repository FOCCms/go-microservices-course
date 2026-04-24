package converter

import (
	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
)

func OrderToOrderDto(order model.Order) *orderv1.OrderDto {
	return &orderv1.OrderDto{
		OrderUUID:       order.UUID,
		HullUUID:        order.HullUUID,
		EngineUUID:      order.EngineUUID,
		ShieldUUID:      toOptNilUUID(order.ShieldUUID),
		WeaponUUID:      toOptNilUUID(order.WeaponUUID),
		TotalPrice:      order.TotalPrice,
		TransactionUUID: toOptNilUUID(order.TransactionUUID),
		PaymentMethod:   toOptNilPaymentMethod(order.PaymentMethod),
		Status:          orderv1.OrderStatus(order.Status),
		CreatedAt:       order.CreatedAt,
	}
}

func toOptNilUUID(val *uuid.UUID) orderv1.OptNilUUID {
	var opt orderv1.OptNilUUID
	if val == nil {
		opt.SetToNull()
	} else {
		opt.SetTo(*val)
	}
	return opt
}

func toOptNilPaymentMethod(method *model.PaymentMethod) orderv1.OptNilPaymentMethod {
	var opt orderv1.OptNilPaymentMethod
	if method == nil || *method == model.PaymentMethodUnspecified {
		opt.SetToNull()
	} else {
		opt.SetTo(orderv1.PaymentMethod(*method))
	}
	return opt
}

func ToPaymentMethod(method orderv1.PaymentMethod) model.PaymentMethod {
	switch method {
	case orderv1.PaymentMethodCARD:
		return model.PaymentMethodCard
	case orderv1.PaymentMethodSBP:
		return model.PaymentMethodSBP
	case orderv1.PaymentMethodCREDITCARD:
		return model.PaymentMethodCreditCard
	case orderv1.PaymentMethodINVESTORMONEY:
		return model.PaymentMethodInvestorMoney
	default:
		return model.PaymentMethodUnspecified
	}
}

func ToCreateOrderRequest(r orderv1.CreateOrderRequest) model.CreateOrderRequest {
	req := model.CreateOrderRequest{
		HullUUID:   r.HullUUID,
		EngineUUID: r.EngineUUID,
	}

	shieldUUID, ok := r.ShieldUUID.Get()
	if !ok {
		req.ShieldUUID = nil
	} else {
		req.ShieldUUID = &shieldUUID
	}

	weaponUUID, ok := r.WeaponUUID.Get()
	if !ok {
		req.WeaponUUID = nil
	} else {
		req.WeaponUUID = &weaponUUID
	}

	return req
}
