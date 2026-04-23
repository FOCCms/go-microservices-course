package converter

import (
	"github.com/FOCCms/go-microservices-course/payment/internal/model"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

func ProtoPartTypeToPartType(t paymentv1.PaymentMethod) model.PaymentMethod {
	switch t {
	case paymentv1.PaymentMethod_PAYMENT_METHOD_CARD:
		return model.PaymentMethodCard
	case paymentv1.PaymentMethod_PAYMENT_METHOD_SBP:
		return model.PaymentMethodSBP
	case paymentv1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD:
		return model.PaymentMethodCreditCard
	case paymentv1.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY:
		return model.PaymentMethodInvestorMoney
	default:
		return model.PaymentMethodUnspecified
	}
}
