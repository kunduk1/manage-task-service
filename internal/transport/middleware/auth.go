// Package middleware содержит HTTP-middleware.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/kunduk1/manage-task-service/internal/token"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
)

type ctxKey string

const userIDKey ctxKey = "userID"

const bearerPrefix = "Bearer "

// Auth возвращает middleware, проверяющий access-JWT и кладущий id пользователя (subject) в контекст.
// Пока не навешан ни на один маршрут — задел под защищённые ручки (tasks/teams).
func Auth(manager *token.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, bearerPrefix) {
				response.Error(w, http.StatusUnauthorized, "missing or malformed authorization header")
				return
			}

			claims, err := manager.ParseAccess(strings.TrimPrefix(header, bearerPrefix))
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext извлекает id аутентифицированного пользователя (subject JWT) из контекста запроса.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}
