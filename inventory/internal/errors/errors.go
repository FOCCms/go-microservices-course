package errs

import "errors"

var (
	ErrPartNotFound      = errors.New("деталь не найдена")
	ErrInvalidUUID       = errors.New("неверный формат UUID")
	ErrOutOfStock        = errors.New("деталь отсутствует на складе")
	ErrNothingToRelease  = errors.New("нечего освобождать")
	ErrIncompatibleParts = errors.New("детали несовместимы")
	ErrPartTypeMismatch  = errors.New("тип детали не соответствует слоту корабля")
	ErrInvalidProperties = errors.New("некорректные свойства детали")
	ErrPartIsNonUnique   = errors.New("повторяющиеся UUID в запросе")
	ErrCommitParts       = errors.New("не все детали получилось списать")
)
