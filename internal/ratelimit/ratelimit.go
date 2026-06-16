package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/kunduk1/manage-task-service/internal/clients/cache"
)

// Limiter — счётчик запросов с фиксированным окном поверх cache.Client.
type Limiter struct {
	cache  cache.Client
	limit  int
	window time.Duration
}

func New(c cache.Client, limit int, window time.Duration) *Limiter {
	return &Limiter{cache: c, limit: limit, window: window}
}

// Result — итог проверки лимита.
type Result struct {
	Allowed    bool          // разрешён ли запрос
	RetryAfter time.Duration // сколько ждать до сброса окна (для заголовка Retry-After)
}

// Allow увеличивает счётчик запросов для id в текущем окне
// и сообщает, не превышен ли лимит. Ключ привязан к номеру окна,
// поэтому следующее окно всегда использует свежий ключ.
func (l *Limiter) Allow(ctx context.Context, id string) (Result, error) {
	windowSec := int64(l.window / time.Second)
	if windowSec <= 0 {
		windowSec = 1
	}

	now := time.Now().Unix()
	bucket := now / windowSec
	key := fmt.Sprintf("ratelimit:user:%s:%d", id, bucket)

	n, err := l.cache.Incr(ctx, key)
	if err != nil {
		return Result{}, err
	}

	// Первое попадание в окно — выставляем TTL, чтобы ключ самоочистился.
	if n == 1 {
		if err := l.cache.Expire(ctx, key, l.window); err != nil {
			return Result{}, err
		}
	}

	// Время до конца текущего окна (от 1с до window).
	retryAfter := time.Duration((bucket+1)*windowSec-now) * time.Second

	return Result{
		Allowed:    n <= int64(l.limit),
		RetryAfter: retryAfter,
	}, nil
}
