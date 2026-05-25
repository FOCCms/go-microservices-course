package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func ErrorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("перехвачена паника",
					"error", rec,
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)
			}
		}()

		sw := &statusWriter{ResponseWriter: w, code: http.StatusOK}

		next.ServeHTTP(sw, r)

		if sw.code > 400 {
			slog.Warn("ошибка HTTP",
				"status", sw.code,
				"method", r.Method,
				"path", r.URL.Path,
			)
		}
	})
}
