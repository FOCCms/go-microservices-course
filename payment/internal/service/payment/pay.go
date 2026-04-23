package payment

import (
	"context"
	"log/slog"

	errs "github.com/FOCCms/go-microservices-course/payment/internal/errors"
	"github.com/FOCCms/go-microservices-course/payment/internal/model"
	"github.com/google/uuid"
)

// PayOrder обрабатывает оплату заказа.
func (*service) Pay(ctx context.Context, req model.PayRequest) (string, error) {
	// Парсим параметры.
	if _, err := uuid.Parse(req.OrderUUID); err != nil {
		return "", errs.ErrInvalidOrderUUID
	}

	if !req.PaymentMethod.IsValid() {
		return "", errs.ErrInvalidPaymentMethod
	}

	// Выполняем оплату.
	transactionUuid := uuid.New()

	slog.Info("оплата прошла успешно",
		"order_uuid:", req.OrderUUID, ", transaction_uuid:", transactionUuid,
	)

	return transactionUuid.String(), nil
}
