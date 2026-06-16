package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry        *prometheus.Registry
	requestsTotal   *prometheus.CounterVec
	errorsTotal     *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

// New создаёт реестр, регистрирует HTTP-коллекторы, а также
// стандартные коллекторы рантайма и процесса.
func New() *Metrics {
	labels := []string{"method", "path", "status"}

	m := &Metrics{
		registry: prometheus.NewRegistry(),
		requestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество обработанных HTTP-запросов.",
		}, labels),
		errorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Количество HTTP-запросов, завершившихся ошибкой (status >= 500).",
		}, labels),
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Время ответа на HTTP-запрос в секундах.",
			Buckets: prometheus.DefBuckets,
		}, labels),
	}

	m.registry.MustRegister(
		m.requestsTotal,
		m.errorsTotal,
		m.requestDuration,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return m
}

// ObserveRequest учитывает один обработанный запрос: инкрементирует счётчики
// и пишет длительность в гистограмму. status == 0 трактуется как 200
// (обёртка chi возвращает 0, если хендлер не выставил код явно).
func (m *Metrics) ObserveRequest(method, path string, status int, dur time.Duration) {
	if status == 0 {
		status = http.StatusOK
	}
	statusStr := strconv.Itoa(status)

	m.requestsTotal.WithLabelValues(method, path, statusStr).Inc()
	m.requestDuration.WithLabelValues(method, path, statusStr).Observe(dur.Seconds())

	if status >= http.StatusInternalServerError {
		m.errorsTotal.WithLabelValues(method, path, statusStr).Inc()
	}
}

// Handler отдаёт HTTP-обработчик для эндпоинта /metrics.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
