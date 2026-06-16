package jwt

import (
	"fmt"
	"time"

	"github.com/kunduk1/manage-task-service/internal/config/envx"
)

type Config struct {
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New() (*Config, error) {
	secret := envx.String("JWT_SECRET", "")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return &Config{
		secret:     secret,
		accessTTL:  envx.Duration("JWT_ACCESS_TTL", 15*time.Minute),
		refreshTTL: envx.Duration("JWT_REFRESH_TTL", 720*time.Hour),
	}, nil
}

func (c *Config) Secret() string { return c.secret }

func (c *Config) AccessTTL() time.Duration { return c.accessTTL }

func (c *Config) RefreshTTL() time.Duration { return c.refreshTTL }
