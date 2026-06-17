// Package circuitbreaker реализует переиспользуемый брейкер (circuit breaker)
// для защиты вызовов внешних сервисов: при череде ошибок брейкер «размыкается»
// и мгновенно возвращает ошибку (fail-fast), не дёргая упавший бэкенд, а затем
// пробует восстановиться через одиночные пробные вызовы (half-open).
//
// Пакет намеренно не зависит ни от логгера, ни от метрик: наблюдаемость
// подключается снаружи через колбэк Settings.OnStateChange.
package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrOpen возвращается из Execute, когда брейкер разомкнут и вызов короткозамкнут.
var ErrOpen = errors.New("circuit breaker is open")

// State — состояние брейкера.
type State int

const (
	// StateClosed — нормальная работа: вызовы проходят, ошибки считаются.
	StateClosed State = iota
	// StateOpen — брейкер разомкнут: вызовы мгновенно отбиваются ErrOpen.
	StateOpen
	// StateHalfOpen — пробный режим: пропускается ограниченное число вызовов-проб.
	StateHalfOpen
)

// String возвращает человекочитаемое имя состояния (для логов и метрик).
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Дефолтные значения настроек.
const (
	defaultMaxFailures = 5
	defaultOpenTimeout = 30 * time.Second
	defaultHalfOpenMax = 1
)

// Settings — конфигурация брейкера.
type Settings struct {
	// Name — имя брейкера (прокидывается в OnStateChange для логов/метрик).
	Name string
	// MaxFailures — число подряд идущих ошибок в closed для перехода в open.
	MaxFailures int
	// OpenTimeout — сколько держим open, прежде чем пустить пробный вызов (half-open).
	OpenTimeout time.Duration
	// HalfOpenMax — максимум одновременных пробных вызовов в half-open (обычно 1).
	HalfOpenMax int
	// OnStateChange — хук смены состояния; вызывается ВНЕ мьютекса. Может быть nil.
	OnStateChange func(name string, from, to State)
}

// Breaker — потокобезопасный брейкер.
type Breaker struct {
	name          string
	maxFailures   int
	openTimeout   time.Duration
	halfOpenMax   int
	onStateChange func(name string, from, to State)

	mu               sync.Mutex
	state            State
	consecFailures   int       // подряд идущие ошибки в closed
	openedAt         time.Time // момент перехода в open
	halfOpenInFlight int       // активных пробных вызовов в half-open
	halfOpenSuccess  int       // успешных проб в half-open

	// now — источник времени; подменяется в тестах через SetNow.
	now func() time.Time
}

// New создаёт брейкер по настройкам, подставляя дефолты для нулевых значений.
func New(s Settings) *Breaker {
	if s.MaxFailures <= 0 {
		s.MaxFailures = defaultMaxFailures
	}
	if s.OpenTimeout <= 0 {
		s.OpenTimeout = defaultOpenTimeout
	}
	if s.HalfOpenMax <= 0 {
		s.HalfOpenMax = defaultHalfOpenMax
	}

	return &Breaker{
		name:          s.Name,
		maxFailures:   s.MaxFailures,
		openTimeout:   s.OpenTimeout,
		halfOpenMax:   s.HalfOpenMax,
		onStateChange: s.OnStateChange,
		state:         StateClosed,
		now:           time.Now,
	}
}

// Execute прогоняет fn через брейкер. Если брейкер разомкнут (или исчерпан лимит
// проб в half-open), fn не вызывается и возвращается ErrOpen. Иначе результат fn
// учитывается как успех/ошибка и при необходимости меняет состояние брейкера.
func (b *Breaker) Execute(ctx context.Context, fn func() error) error {
	if err := b.before(); err != nil {
		return err
	}

	err := fn()

	b.after(err == nil)
	return err
}

// before решает, пропустить ли вызов, и фиксирует переходы по таймауту/пробам.
func (b *Breaker) before() error {
	b.mu.Lock()
	var from, to State
	var changed bool

	defer func() {
		b.mu.Unlock()
		if changed {
			b.notify(from, to)
		}
	}()

	switch b.state {
	case StateClosed:
		return nil

	case StateOpen:
		// Время в open истекло — переходим в half-open и пускаем эту пробу.
		if b.now().Sub(b.openedAt) >= b.openTimeout {
			from, to, changed = b.setState(StateHalfOpen)
			b.halfOpenInFlight = 1
			return nil
		}
		return ErrOpen

	case StateHalfOpen:
		// В half-open пропускаем лишь ограниченное число проб одновременно.
		if b.halfOpenInFlight >= b.halfOpenMax {
			return ErrOpen
		}
		b.halfOpenInFlight++
		return nil

	default:
		return nil
	}
}

// after учитывает итог вызова и при необходимости меняет состояние.
func (b *Breaker) after(success bool) {
	b.mu.Lock()
	var from, to State
	var changed bool

	defer func() {
		b.mu.Unlock()
		if changed {
			b.notify(from, to)
		}
	}()

	switch b.state {
	case StateClosed:
		if success {
			b.consecFailures = 0
			return
		}
		b.consecFailures++
		if b.consecFailures >= b.maxFailures {
			from, to, changed = b.setState(StateOpen)
			b.openedAt = b.now()
		}

	case StateHalfOpen:
		if b.halfOpenInFlight > 0 {
			b.halfOpenInFlight--
		}
		if success {
			b.halfOpenSuccess++
			if b.halfOpenSuccess >= b.halfOpenMax {
				from, to, changed = b.setState(StateClosed)
			}
			return
		}
		// Любая неудачная проба немедленно возвращает брейкер в open.
		from, to, changed = b.setState(StateOpen)
		b.openedAt = b.now()

	case StateOpen:
		// До истечения таймаута вызовы сюда не доходят; на всякий случай — no-op.
	}
}

// setState переводит брейкер в новое состояние и сбрасывает счётчики.
// Возвращает (from, to, changed); сам колбэк НЕ вызывает (его дёргает notify
// уже после освобождения мьютекса). Вызывается только под b.mu.
func (b *Breaker) setState(to State) (State, State, bool) {
	from := b.state
	if from == to {
		return from, to, false
	}
	b.state = to
	b.consecFailures = 0
	b.halfOpenInFlight = 0
	b.halfOpenSuccess = 0
	return from, to, true
}

// notify вызывает OnStateChange вне мьютекса, чтобы колбэк мог безопасно
// (без дедлока) обращаться к брейкеру или к внешним системам.
func (b *Breaker) notify(from, to State) {
	if b.onStateChange != nil {
		b.onStateChange(b.name, from, to)
	}
}

// State возвращает текущее состояние брейкера (потокобезопасно).
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// SetNow подменяет источник времени. Предназначен для тестов
// (детерминированная проверка переходов по таймауту без реальных sleep).
func (b *Breaker) SetNow(f func() time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.now = f
}
