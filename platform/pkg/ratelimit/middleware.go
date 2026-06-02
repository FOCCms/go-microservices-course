package ratelimit

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-redis/redis_rate/v10"
)

func Middleware(limiter *redis_rate.Limiter, rps, burst int) func(next http.Handler) http.Handler {
	limit := redis_rate.Limit{
		Rate:   rps,
		Burst:  burst,
		Period: time.Second,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res, err := limiter.Allow(r.Context(), r.URL.Path, limit)
			if err != nil {
				slog.Error("ошибка проверки rate limit", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			if res.Allowed == 0 {
				slog.Warn(
					"запрос отклонён rate limiter",
					"retry_after", res.RetryAfter,
				)
				http.Error(w, "запрос отклонён rate limiter", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
