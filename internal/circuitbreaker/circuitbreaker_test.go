package circuitbreaker

import (
	stderrors "errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var errBoom = stderrors.New("boom")

// newTestBreaker собирает брейкер с подменяемыми часами. Возвращает брейкер
// и указатель на «текущее время», которое тест двигает присваиванием *clock.
func newTestBreaker(t *testing.T, s Settings) (*Breaker, *time.Time) {
	t.Helper()
	b := New(s)
	clock := time.Unix(0, 0)
	b.SetNow(func() time.Time { return clock })
	return b, &clock
}

// fail/ok — удобные функции-вызовы для Execute.
func fail() error { return errBoom }
func ok() error   { return nil }

func TestClosed_AllowsSuccess(t *testing.T) {
	b, _ := newTestBreaker(t, Settings{MaxFailures: 3})

	require.NoError(t, b.Execute(t.Context(), ok))
	require.Equal(t, StateClosed, b.State())
}

func TestClosed_ResetsOnSuccess(t *testing.T) {
	b, _ := newTestBreaker(t, Settings{MaxFailures: 3})

	_ = b.Execute(t.Context(), fail)
	_ = b.Execute(t.Context(), fail)
	// Успех обнуляет счётчик подряд идущих ошибок.
	_ = b.Execute(t.Context(), ok)
	// Ещё две ошибки — суммарно их было бы >=3, но из-за сброса брейкер закрыт.
	_ = b.Execute(t.Context(), fail)
	_ = b.Execute(t.Context(), fail)

	require.Equal(t, StateClosed, b.State())
}

func TestTripsAfterNConsecutiveFailures(t *testing.T) {
	b, _ := newTestBreaker(t, Settings{MaxFailures: 3})

	for i := 0; i < 3; i++ {
		err := b.Execute(t.Context(), fail)
		require.ErrorIs(t, err, errBoom)
	}
	require.Equal(t, StateOpen, b.State())

	// В open вызов короткозамкнут: fn не выполняется, возвращается ErrOpen.
	called := false
	err := b.Execute(t.Context(), func() error { called = true; return nil })
	require.ErrorIs(t, err, ErrOpen)
	require.False(t, called)
}

func TestOpen_FastFailsBeforeTimeout(t *testing.T) {
	b, clock := newTestBreaker(t, Settings{MaxFailures: 1, OpenTimeout: 30 * time.Second})

	_ = b.Execute(t.Context(), fail) // тут же размыкается (MaxFailures=1)
	require.Equal(t, StateOpen, b.State())

	*clock = clock.Add(30*time.Second - time.Nanosecond) // ещё не истёк таймаут

	called := false
	err := b.Execute(t.Context(), func() error { called = true; return nil })
	require.ErrorIs(t, err, ErrOpen)
	require.False(t, called)
}

func TestOpen_TransitionsToHalfOpenAfterTimeout(t *testing.T) {
	b, clock := newTestBreaker(t, Settings{MaxFailures: 1, OpenTimeout: 30 * time.Second})

	_ = b.Execute(t.Context(), fail)
	*clock = clock.Add(30 * time.Second) // таймаут истёк ровно

	var observed State
	// В half-open пробный вызов выполняется; фиксируем состояние внутри fn.
	err := b.Execute(t.Context(), func() error {
		observed = b.State()
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, StateHalfOpen, observed)
}

func TestHalfOpen_SuccessCloses(t *testing.T) {
	b, clock := newTestBreaker(t, Settings{MaxFailures: 1, OpenTimeout: 30 * time.Second})

	_ = b.Execute(t.Context(), fail)
	*clock = clock.Add(30 * time.Second)

	require.NoError(t, b.Execute(t.Context(), ok)) // успешная проба закрывает
	require.Equal(t, StateClosed, b.State())
}

func TestHalfOpen_FailureReopens(t *testing.T) {
	b, clock := newTestBreaker(t, Settings{MaxFailures: 1, OpenTimeout: 30 * time.Second})

	_ = b.Execute(t.Context(), fail)
	*clock = clock.Add(30 * time.Second)

	// Неудачная проба немедленно возвращает в open и перезапускает таймер.
	_ = b.Execute(t.Context(), fail)
	require.Equal(t, StateOpen, b.State())

	// Сразу после повторного размыкания таймаут ещё не истёк — fast-fail.
	called := false
	err := b.Execute(t.Context(), func() error { called = true; return nil })
	require.ErrorIs(t, err, ErrOpen)
	require.False(t, called)
}

func TestHalfOpen_SecondProbeFastFails(t *testing.T) {
	b, clock := newTestBreaker(t, Settings{MaxFailures: 1, OpenTimeout: 30 * time.Second, HalfOpenMax: 1})

	_ = b.Execute(t.Context(), fail)
	*clock = clock.Add(30 * time.Second)

	entered := make(chan struct{})
	release := make(chan struct{})

	var probeErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		probeErr = b.Execute(t.Context(), func() error {
			close(entered)
			<-release // держим единственную пробу «в полёте»
			return nil
		})
	}()

	<-entered // дождались, что проба реально выполняется

	// Пока проба в полёте, второй вызов должен быть короткозамкнут.
	err := b.Execute(t.Context(), ok)
	require.ErrorIs(t, err, ErrOpen)

	close(release)
	wg.Wait()
	require.NoError(t, probeErr)
}

func TestOnStateChange_FiresFullCycle(t *testing.T) {
	type transition struct{ from, to State }
	var (
		mu          sync.Mutex
		transitions []transition
	)

	b, clock := newTestBreaker(t, Settings{
		MaxFailures: 2,
		OpenTimeout: 30 * time.Second,
		OnStateChange: func(_ string, from, to State) {
			mu.Lock()
			transitions = append(transitions, transition{from, to})
			mu.Unlock()
		},
	})

	_ = b.Execute(t.Context(), fail)
	_ = b.Execute(t.Context(), fail) // closed -> open
	*clock = clock.Add(30 * time.Second)
	_ = b.Execute(t.Context(), ok) // open -> half-open -> closed

	want := []transition{
		{StateClosed, StateOpen},
		{StateOpen, StateHalfOpen},
		{StateHalfOpen, StateClosed},
	}
	require.Equal(t, want, transitions)
}

func TestConcurrent_RaceSafe(t *testing.T) {
	b := New(Settings{MaxFailures: 5, OpenTimeout: time.Hour})

	var wg sync.WaitGroup
	var executed int64
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = b.Execute(t.Context(), func() error {
					atomic.AddInt64(&executed, 1)
					return errBoom
				})
			}
		}()
	}
	wg.Wait()

	// При сплошных ошибках брейкер обязан оказаться разомкнутым.
	require.Equal(t, StateOpen, b.State())
	require.NotZero(t, atomic.LoadInt64(&executed))
}
