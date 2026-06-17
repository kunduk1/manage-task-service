// Package email — конфигурация мок-сервиса отправки писем (имитация внешнего бэкенда).
package email

import (
	"time"

	"github.com/kunduk1/manage-task-service/internal/config/envx"
)

type Config struct {
	enabled     bool
	failPercent int // доля искусственных сбоев, 0..100
	latency     time.Duration
}

func New() *Config {
	return &Config{
		enabled:     envx.Bool("EMAIL_ENABLED", true),
		failPercent: envx.Int("EMAIL_FAIL_PERCENT", 0),
		latency:     envx.Duration("EMAIL_LATENCY", 50*time.Millisecond),
	}
}

// Enabled — включён ли почтовый клиент (иначе отправка пропускается).
func (c *Config) Enabled() bool { return c.enabled }

// FailRate — вероятность искусственного сбоя отправки (0.0..1.0).
// envx не умеет float, поэтому знаем процент и делим на 100.
func (c *Config) FailRate() float64 { return float64(c.failPercent) / 100.0 }

// Latency — искусственная задержка «сетевого» вызова.
func (c *Config) Latency() time.Duration { return c.latency }
