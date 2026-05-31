package v1

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/ogen-go/ogen/ogenerrors"
)

// errorResponse представляет JSON-ответ с ошибкой
type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorHandler маппит доменные ошибки в HTTP-ответы
func ErrorHandler(ctx context.Context, w http.ResponseWriter, _ *http.Request, err error) {
	code, message := mapError(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if encErr := json.NewEncoder(w).Encode(errorResponse{
		Code:    code,
		Message: message,
	}); encErr != nil {
		slog.ErrorContext(ctx, "ошибка кодирования ответа", "error", encErr)
	}
}

func mapError(err error) (int, string) {
	// Ошибки декодирования/валидации запроса от ogen — всегда 400.
	var decodeParams *ogenerrors.DecodeParamsError
	var decodeRequest *ogenerrors.DecodeRequestError

	switch {
	case errors.As(err, &decodeParams), errors.As(err, &decodeRequest):
		return http.StatusBadRequest, err.Error()

	// 401 Unauthorized
	case errors.Is(err, errs.ErrUnauthorized):
		return http.StatusUnauthorized, err.Error()

	// 404 Not Found
	case errors.Is(err, errs.ErrOrderNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, errs.ErrPartNotFound):
		return http.StatusNotFound, err.Error()

	// 409 Conflict
	case errors.Is(err, errs.ErrOutOfStock):
		return http.StatusConflict, err.Error()
	case errors.Is(err, errs.ErrIncompatibleParts):
		return http.StatusConflict, err.Error()
	case errors.Is(err, errs.ErrOrderAlreadyPaid):
		return http.StatusConflict, err.Error()
	case errors.Is(err, errs.ErrOrderCancelled):
		return http.StatusConflict, err.Error()
	case errors.Is(err, errs.ErrOrderAssembled):
		return http.StatusConflict, err.Error()

	// 400 Bad Request
	case errors.Is(err, errs.ErrInvalidUUID):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, errs.ErrPartTypeMismatch):
		return http.StatusBadRequest, err.Error()

	// 500 Internal Server Error
	default:
		return http.StatusInternalServerError, "внутренняя ошибка"
	}
}
