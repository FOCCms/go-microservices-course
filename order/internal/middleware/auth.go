package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/platform/pkg/auth"
)

func AuthMiddleware(iamClient IAMClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "отсутствует заголовок Authorization", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "неверный формат Authorization", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				http.Error(w, "неверный формат Authorization", http.StatusUnauthorized)
				return
			}
			session, user, err := iamClient.Whoami(r.Context(), token)
			if err != nil {
				http.Error(w, "недействительная сессия", http.StatusUnauthorized)
				return
			}

			ctx := auth.WithSessionUUID(r.Context(), session.Uuid)
			ctx = auth.WithUserUUID(ctx, uuid.MustParse(user.Uuid))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
