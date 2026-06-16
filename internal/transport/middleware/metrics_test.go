package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/kunduk1/manage-task-service/internal/metrics"
)

// scrape прогоняет один запрос через chi-роутер с Metrics-middleware,
// затем снимает текст экспозиции с m.Handler().
func scrape(t *testing.T, setup func(r chi.Router), reqMethod, reqTarget string) string {
	t.Helper()

	m := metrics.New()

	r := chi.NewRouter()
	r.Use(Metrics(m))
	setup(r)
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(reqMethod, reqTarget, nil))

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics handler returned %d, want 200", rec.Code)
	}
	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("failed to read metrics body: %v", err)
	}
	return string(body)
}

func mustContain(t *testing.T, body, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Errorf("metrics output missing %q", want)
	}
}

func mustNotContain(t *testing.T, body, unwanted string) {
	t.Helper()
	if strings.Contains(body, unwanted) {
		t.Errorf("metrics output unexpectedly contains %q", unwanted)
	}
}

func TestMetrics_Success(t *testing.T) {
	body := scrape(t, func(r chi.Router) {
		r.Get("/widgets/{id}", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}, http.MethodGet, "/widgets/42")

	// Путь — это шаблон маршрута, а не конкретный URL.
	mustContain(t, body, `http_requests_total{method="GET",path="/widgets/{id}",status="200"} 1`)
	mustContain(t, body, `http_request_duration_seconds_count{method="GET",path="/widgets/{id}",status="200"} 1`)
	// Успешный запрос не должен попадать в счётчик ошибок.
	mustNotContain(t, body, "http_errors_total{")
	// Стандартные коллекторы рантайма зарегистрированы.
	mustContain(t, body, "go_goroutines")
}

func TestMetrics_ServerError(t *testing.T) {
	body := scrape(t, func(r chi.Router) {
		r.Get("/boom", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
	}, http.MethodGet, "/boom")

	mustContain(t, body, `http_requests_total{method="GET",path="/boom",status="500"} 1`)
	mustContain(t, body, `http_errors_total{method="GET",path="/boom",status="500"} 1`)
	mustContain(t, body, `http_request_duration_seconds_count{method="GET",path="/boom",status="500"} 1`)
}

func TestMetrics_UnmatchedRoute(t *testing.T) {
	// Запрос не совпал ни с одним маршрутом → chi отдаёт 404, шаблон пустой.
	// Хотя бы один маршрут нужен, чтобы chi построил цепочку middleware.
	body := scrape(t, func(r chi.Router) {
		r.Get("/known", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}, http.MethodGet, "/does-not-exist")

	mustContain(t, body, `http_requests_total{method="GET",path="unknown",status="404"} 1`)
	// 404 — это не серверная ошибка.
	mustNotContain(t, body, "http_errors_total{")
}
