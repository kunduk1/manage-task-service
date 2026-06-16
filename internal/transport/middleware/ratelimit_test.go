package middleware

import (
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/kunduk1/manage-task-service/internal/logger"
	"github.com/kunduk1/manage-task-service/internal/ratelimit"
)

// Middleware best-effort логирует fail-open через глобальный логгер — в тестах он не
// инициализирован, поэтому ставим no-op, чтобы logger.Warn не паниковал на nil.
func init() {
	logger.NewGlobalLogger(zapcore.NewNopCore())
}

// fakeLimiter — ручная подмена лимитера для middleware-тестов.
type fakeLimiter struct {
	res ratelimit.Result
	err error
}

func (f fakeLimiter) Allow(_ context.Context, _ string) (ratelimit.Result, error) {
	return f.res, f.err
}

// serveRateLimit прогоняет запрос через RateLimit-middleware c заданным userID в контексте.
func serveRateLimit(l limiter, userID string, withUser bool) (rec *httptest.ResponseRecorder, nextCalled bool) {
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if withUser {
		req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	RateLimit(l)(next).ServeHTTP(rec, req)
	return rec, nextCalled
}

func TestRateLimit_Allowed(t *testing.T) {
	l := fakeLimiter{res: ratelimit.Result{Allowed: true, RetryAfter: 30 * time.Second}}
	rec, nextCalled := serveRateLimit(l, "42", true)

	if !nextCalled {
		t.Fatal("expected next handler to be called when allowed")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRateLimit_Blocked(t *testing.T) {
	l := fakeLimiter{res: ratelimit.Result{Allowed: false, RetryAfter: 30 * time.Second}}
	rec, nextCalled := serveRateLimit(l, "42", true)

	if nextCalled {
		t.Error("next handler must not be called when blocked")
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}
	if got := rec.Header().Get("Retry-After"); got != "30" {
		t.Errorf("expected Retry-After=30, got %q", got)
	}
}

func TestRateLimit_FailOpenOnError(t *testing.T) {
	l := fakeLimiter{err: stderrors.New("redis down")}
	rec, nextCalled := serveRateLimit(l, "42", true)

	if !nextCalled {
		t.Fatal("expected fail-open: next handler must be called on limiter error")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 on fail-open, got %d", rec.Code)
	}
}

func TestRateLimit_NoUserSkips(t *testing.T) {
	// Без userID в контексте (защитная ветка) — лимитер не вызывается, запрос проходит.
	l := fakeLimiter{res: ratelimit.Result{Allowed: false}}
	rec, nextCalled := serveRateLimit(l, "", false)

	if !nextCalled {
		t.Fatal("expected next handler to be called when no userID in context")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
