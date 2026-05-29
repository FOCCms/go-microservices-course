package payment

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/payment/internal/errors"
	"github.com/FOCCms/go-microservices-course/payment/internal/model"
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
		slog.String("order_uuid", req.OrderUUID),
		slog.String("transaction_uuid", transactionUuid.String()),
		slog.String("payment_method", string(req.PaymentMethod)),
	)

	return transactionUuid.String(), nil
}
