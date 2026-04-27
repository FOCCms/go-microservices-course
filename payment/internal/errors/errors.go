package errs

import "errors"

var (
	ErrInvalidOrderUUID     = errors.New("неверный формат UUID заказа")
	ErrInvalidPaymentMethod = errors.New("неверный метод оплаты")
)
