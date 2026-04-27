package errs

import "errors"

var (
	ErrOrderNotFound       = errors.New("заказ не найден")
	ErrOrderAlreadyPaid    = errors.New("заказ уже оплачен")
	ErrOrderCancelled      = errors.New("заказ отменён")
	ErrOrderStatusConflict = errors.New("невалидный статус для операции")
	ErrPartNotFound        = errors.New("деталь не найдена")
	ErrOutOfStock          = errors.New("деталь отсутствует на складе")
	ErrInvalidUUID         = errors.New("неверный формат UUID")
	ErrPartRequired        = errors.New("не указаны обязательные детали")
)
