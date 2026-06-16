package redis

import (
	"time"

	"github.com/kunduk1/manage-task-service/internal/config/envx"
)

type Config struct {
	addr            string
	password        string
	db              int
	poolSize        int
	minIdleConns    int
	connMaxLifetime time.Duration
}

func New() *Config {
	return &Config{
		addr:            envx.String("REDIS_ADDR", "localhost:6379"),
		password:        envx.String("REDIS_PASSWORD", ""),
		db:              envx.Int("REDIS_DB", 0),
		poolSize:        envx.Int("REDIS_POOL_SIZE", 0),
		minIdleConns:    envx.Int("REDIS_MIN_IDLE_CONNS", 0),
		connMaxLifetime: envx.Duration("REDIS_CONN_MAX_LIFETIME", 0),
	}
}

func (c *Config) Addr() string { return c.addr }

func (c *Config) Password() string { return c.password }

func (c *Config) DB() int { return c.db }

func (c *Config) PoolSize() int { return c.poolSize }

func (c *Config) MinIdleConns() int { return c.minIdleConns }

func (c *Config) ConnMaxLifetime() time.Duration { return c.connMaxLifetime }
