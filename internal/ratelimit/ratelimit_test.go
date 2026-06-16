package ratelimit

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	cachemocks "github.com/kunduk1/manage-task-service/internal/clients/cache/mocks"
)

func newTestLimiter(t *testing.T, limit int, window time.Duration) (*Limiter, *cachemocks.MockClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	client := cachemocks.NewMockClient(ctrl)
	return New(client, limit, window), client
}

func TestAllow_FirstRequestSetsTTL(t *testing.T) {
	l, client := newTestLimiter(t, 100, time.Minute)

	// Первое попадание в окно: Incr вернул 1 → выставляем TTL.
	client.EXPECT().Incr(gomock.Any(), gomock.Any()).Return(int64(1), nil)
	client.EXPECT().Expire(gomock.Any(), gomock.Any(), time.Minute).Return(nil)

	res, err := l.Allow(context.Background(), "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Allowed {
		t.Error("first request must be allowed")
	}
	if res.RetryAfter <= 0 || res.RetryAfter > time.Minute {
		t.Errorf("retry-after out of range: %v", res.RetryAfter)
	}
}

func TestAllow_UnderLimitNoTTL(t *testing.T) {
	l, client := newTestLimiter(t, 100, time.Minute)

	// Не первый запрос в окне (Incr вернул 50): TTL уже выставлен, Expire не зовём.
	client.EXPECT().Incr(gomock.Any(), gomock.Any()).Return(int64(50), nil)

	res, err := l.Allow(context.Background(), "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Allowed {
		t.Error("request within limit must be allowed")
	}
}

func TestAllow_AtLimitBoundary(t *testing.T) {
	l, client := newTestLimiter(t, 100, time.Minute)

	// Ровно лимит (100) — ещё разрешено.
	client.EXPECT().Incr(gomock.Any(), gomock.Any()).Return(int64(100), nil)

	res, err := l.Allow(context.Background(), "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Allowed {
		t.Error("request at exact limit must be allowed")
	}
}

func TestAllow_OverLimitBlocked(t *testing.T) {
	l, client := newTestLimiter(t, 100, time.Minute)

	// Превышение лимита (101) — блокируем, с положительным retry-after.
	client.EXPECT().Incr(gomock.Any(), gomock.Any()).Return(int64(101), nil)

	res, err := l.Allow(context.Background(), "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Allowed {
		t.Error("request over limit must be blocked")
	}
	if res.RetryAfter <= 0 {
		t.Errorf("blocked request must report positive retry-after, got %v", res.RetryAfter)
	}
}

func TestAllow_IncrError(t *testing.T) {
	l, client := newTestLimiter(t, 100, time.Minute)

	boom := stderrors.New("boom")
	client.EXPECT().Incr(gomock.Any(), gomock.Any()).Return(int64(0), boom)

	_, err := l.Allow(context.Background(), "42")
	if !stderrors.Is(err, boom) {
		t.Errorf("expected boom error, got %v", err)
	}
}
