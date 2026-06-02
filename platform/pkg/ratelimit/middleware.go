package ratelimit

import (
	"net/http"

	"github.com/go-redis/redis_rate/v10"
)

func Middleware(limiter *redis_rate.Limiter, rate int, burst int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//TODO implement
			//TODO redis check limit
			next.ServeHTTP(w, r)
		})
	}
}
