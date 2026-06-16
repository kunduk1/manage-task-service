package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestObserveRequest_Success(t *testing.T) {
	m := New()
	m.ObserveRequest(http.MethodGet, "/widgets/{id}", http.StatusOK, 5*time.Millisecond)

	if got := testutil.ToFloat64(m.requestsTotal.WithLabelValues("GET", "/widgets/{id}", "200")); got != 1 {
		t.Errorf("requestsTotal = %v, want 1", got)
	}
	if got := testutil.CollectAndCount(m.errorsTotal); got != 0 {
		t.Errorf("errorsTotal series = %d, want 0", got)
	}
	if got := testutil.CollectAndCount(m.requestDuration); got != 1 {
		t.Errorf("requestDuration series = %d, want 1", got)
	}
}

func TestObserveRequest_ServerError(t *testing.T) {
	m := New()
	m.ObserveRequest(http.MethodPost, "/boom", http.StatusInternalServerError, time.Millisecond)

	if got := testutil.ToFloat64(m.requestsTotal.WithLabelValues("POST", "/boom", "500")); got != 1 {
		t.Errorf("requestsTotal = %v, want 1", got)
	}
	if got := testutil.ToFloat64(m.errorsTotal.WithLabelValues("POST", "/boom", "500")); got != 1 {
		t.Errorf("errorsTotal = %v, want 1", got)
	}
}

func TestObserveRequest_StatusZeroDefaultsTo200(t *testing.T) {
	m := New()
	m.ObserveRequest(http.MethodGet, "/x", 0, time.Millisecond)

	if got := testutil.ToFloat64(m.requestsTotal.WithLabelValues("GET", "/x", "200")); got != 1 {
		t.Errorf("requestsTotal{status=200} = %v, want 1", got)
	}
}

func TestHandler_ExposesMetrics(t *testing.T) {
	m := New()
	m.ObserveRequest(http.MethodGet, "/x", http.StatusOK, time.Millisecond)

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("handler status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"http_requests_total", "http_request_duration_seconds", "go_goroutines"} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics output missing %q", want)
		}
	}
}
