package errs

import "errors"

var (
	ErrOrderNotFound       = errors.New("заказ не найден")
	ErrOrderAlreadyPaid    = errors.New("заказ уже оплачен")
	ErrOrderCancelled      = errors.New("заказ отменён")
	ErrOrderAssembled      = errors.New("заказ собран")
	ErrOrderStatusConflict = errors.New("невалидный статус для операции")
	ErrPartNotFound        = errors.New("деталь не найдена")
	ErrOutOfStock          = errors.New("деталь отсутствует на складе")
	ErrInvalidUUID         = errors.New("неверный формат UUID")
	ErrPartRequired        = errors.New("не указаны обязательные детали")
	ErrIncompatibleParts   = errors.New("детали несовместимы")
	ErrPartTypeMismatch    = errors.New("тип детали не соответствует слоту корабля")
	ErrUnauthorized        = errors.New("пользователь не авторизован")
)
