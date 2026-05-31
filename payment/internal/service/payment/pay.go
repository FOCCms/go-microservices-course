package payment

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	errs "github.com/FOCCms/go-microservices-course/payment/internal/errors"
	"github.com/FOCCms/go-microservices-course/payment/internal/model"
)

// PayOrder обрабатывает оплату заказа.
func (*service) Pay(ctx context.Context, req model.PayRequest) (string, error) {
	ctx, span := otel.Tracer("payment-service").Start(ctx, "payment.Pay")
	defer span.End()

	span.SetAttributes(
		attribute.String("payment.order_uuid", req.OrderUUID),
		attribute.String("payment.method", string(req.PaymentMethod)),
	)
	// Парсим параметры.
	if _, err := uuid.Parse(req.OrderUUID); err != nil {
		span.RecordError(errs.ErrInvalidOrderUUID)
		span.SetStatus(codes.Error, errs.ErrInvalidOrderUUID.Error())
		return "", errs.ErrInvalidOrderUUID
	}

	if !req.PaymentMethod.IsValid() {
		span.RecordError(errs.ErrInvalidPaymentMethod)
		span.SetStatus(codes.Error, errs.ErrInvalidPaymentMethod.Error())
		return "", errs.ErrInvalidPaymentMethod
	}

	// Выполняем оплату.
	transactionUuid := uuid.New()

	slog.Info("оплата прошла успешно",
		slog.String("order_uuid", req.OrderUUID),
		slog.String("transaction_uuid", transactionUuid.String()),
		slog.String("payment_method", string(req.PaymentMethod)),
	)

	span.SetAttributes(attribute.String("payment.transaction_uuid", transactionUuid.String()))
	span.SetStatus(codes.Ok, "оплата прошла успешно")

	return transactionUuid.String(), nil
}
