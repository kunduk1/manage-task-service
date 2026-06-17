package circuitbreaker

import (
	"time"

	"github.com/kunduk1/manage-task-service/internal/config/envx"
)

type Config struct {
	maxFailures int
	openTimeout time.Duration
	halfOpenMax int
}

func New() *Config {
	return &Config{
		maxFailures: envx.Int("CB_MAX_FAILURES", 5),
		openTimeout: envx.Duration("CB_OPEN_TIMEOUT", 30*time.Second),
		halfOpenMax: envx.Int("CB_HALF_OPEN_MAX", 1),
	}
}

// MaxFailures — число подряд идущих ошибок для размыкания брейкера.
func (c *Config) MaxFailures() int { return c.maxFailures }

// OpenTimeout — сколько брейкер остаётся разомкнутым до пробного вызова.
func (c *Config) OpenTimeout() time.Duration { return c.openTimeout }

// HalfOpenMax — максимум одновременных пробных вызовов в half-open.
func (c *Config) HalfOpenMax() int { return c.halfOpenMax }
