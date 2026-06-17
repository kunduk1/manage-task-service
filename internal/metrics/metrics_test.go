package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObserveRequest_Success(t *testing.T) {
	m := New()
	m.ObserveRequest(http.MethodGet, "/widgets/{id}", http.StatusOK, 5*time.Millisecond)

	assert.Equal(t, float64(1), testutil.ToFloat64(m.requestsTotal.WithLabelValues("GET", "/widgets/{id}", "200")))
	assert.Equal(t, 0, testutil.CollectAndCount(m.errorsTotal))
	assert.Equal(t, 1, testutil.CollectAndCount(m.requestDuration))
}

func TestObserveRequest_ServerError(t *testing.T) {
	m := New()
	m.ObserveRequest(http.MethodPost, "/boom", http.StatusInternalServerError, time.Millisecond)

	assert.Equal(t, float64(1), testutil.ToFloat64(m.requestsTotal.WithLabelValues("POST", "/boom", "500")))
	assert.Equal(t, float64(1), testutil.ToFloat64(m.errorsTotal.WithLabelValues("POST", "/boom", "500")))
}

func TestObserveRequest_StatusZeroDefaultsTo200(t *testing.T) {
	m := New()
	m.ObserveRequest(http.MethodGet, "/x", 0, time.Millisecond)

	assert.Equal(t, float64(1), testutil.ToFloat64(m.requestsTotal.WithLabelValues("GET", "/x", "200")))
}

func TestHandler_ExposesMetrics(t *testing.T) {
	m := New()
	m.ObserveRequest(http.MethodGet, "/x", http.StatusOK, time.Millisecond)

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	for _, want := range []string{"http_requests_total", "http_request_duration_seconds", "go_goroutines"} {
		assert.Contains(t, body, want)
	}
}
