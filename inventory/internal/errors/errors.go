package errs

import "errors"

var (
	ErrPartNotFound = errors.New("деталь не найдена")
	ErrInvalidUUID  = errors.New("неверный формат UUID")
)
