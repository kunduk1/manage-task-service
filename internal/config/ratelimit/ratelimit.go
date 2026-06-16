package ratelimit

import (
	"time"

	"github.com/kunduk1/manage-task-service/internal/config/envx"
)

type Config struct {
	enabled  bool
	requests int
	window   time.Duration
}

func New() *Config {
	return &Config{
		enabled:  envx.Bool("RATE_LIMIT_ENABLED", true),
		requests: envx.Int("RATE_LIMIT_REQUESTS", 100),
		window:   envx.Duration("RATE_LIMIT_WINDOW", time.Minute),
	}
}

func (c *Config) Enabled() bool { return c.enabled }

func (c *Config) Requests() int { return c.requests }

func (c *Config) Window() time.Duration { return c.window }
