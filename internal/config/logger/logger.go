package logger

import (
	"github.com/kunduk1/manage-task-service/internal/config/envx"
)

type Config struct {
	level      string
	logPath    string
	maxSize    int
	maxBackups int
	maxAge     int
}

func New() *Config {
	return &Config{
		level:      envx.String("LOG_LEVEL", "info"),
		logPath:    envx.String("LOG_PATH", "logs/app.log"),
		maxSize:    envx.Int("LOG_MAX_SIZE", 100),
		maxBackups: envx.Int("LOG_MAX_BACKUPS", 3),
		maxAge:     envx.Int("LOG_MAX_AGE", 28),
	}
}

func (c *Config) Level() string { return c.level }

func (c *Config) LogPath() string { return c.logPath }

func (c *Config) MaxSize() int { return c.maxSize }

func (c *Config) MaxBackups() int { return c.maxBackups }

func (c *Config) MaxAge() int { return c.maxAge }
