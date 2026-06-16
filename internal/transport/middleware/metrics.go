package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/kunduk1/manage-task-service/internal/metrics"
)

// Metrics учитывает каждый запрос в Prometheus-метриках. Шаблон маршрута
// (например, /api/v1/tasks/{id}/history) читается после next.ServeHTTP, когда
// chi уже разрешил роут — это удерживает кардинальность лейблов под контролем.
func Metrics(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			path := chi.RouteContext(r.Context()).RoutePattern()
			if path == "" {
				path = "unknown"
			}

			m.ObserveRequest(r.Method, path, ww.Status(), time.Since(start))
		})
	}
}
