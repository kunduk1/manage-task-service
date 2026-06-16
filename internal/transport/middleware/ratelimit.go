package middleware

import (
	"context"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/kunduk1/manage-task-service/internal/logger"
	"github.com/kunduk1/manage-task-service/internal/ratelimit"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
)

// limiter — минимальный контракт лимитера, нужный middleware (для подмены в тестах).
type limiter interface {
	Allow(ctx context.Context, id string) (ratelimit.Result, error)
}

// RateLimit возвращает middleware, ограничивающий частоту запросов на пользователя.
// Навешивается после Auth: id берётся из контекста (subject JWT).
// При ошибке кэша работает по принципу fail-open — запрос пропускается.
func RateLimit(l limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				// Защитная ветка: без аутентификации лимитировать некого.
				next.ServeHTTP(w, r)
				return
			}

			res, err := l.Allow(r.Context(), userID)
			if err != nil {
				// Кэш недоступен — не роняем API, пропускаем запрос.
				logger.Warn("rate limiter unavailable, allowing request",
					zap.String("user_id", userID),
					zap.Error(err),
				)
				next.ServeHTTP(w, r)
				return
			}

			if !res.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(res.RetryAfter.Seconds())))
				response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
