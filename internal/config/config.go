package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/kunduk1/manage-task-service/internal/config/http"
	"github.com/kunduk1/manage-task-service/internal/config/jwt"
	"github.com/kunduk1/manage-task-service/internal/config/logger"
	"github.com/kunduk1/manage-task-service/internal/config/mysql"
	"github.com/kunduk1/manage-task-service/internal/config/redis"
)

type HTTPConfig interface {
	Address() string
}

type MySQLConfig interface {
	DSN() string
	MaxOpenConns() int
	MaxIdleConns() int
	ConnMaxLifetime() time.Duration
}

type RedisConfig interface {
	Addr() string
	Password() string
	DB() int
	PoolSize() int
	MinIdleConns() int
	ConnMaxLifetime() time.Duration
}

type JWTConfig interface {
	Secret() string
	AccessTTL() time.Duration
	RefreshTTL() time.Duration
}

type LoggerConfig interface {
	Level() string
	LogPath() string
	MaxSize() int
	MaxBackups() int
	MaxAge() int
}

type Config struct {
	HTTP   HTTPConfig
	MySQL  MySQLConfig
	Redis  RedisConfig
	JWT    JWTConfig
	Logger LoggerConfig
}

// Load читает переменные окружения
// и возвращает заполненную конфигурацию. JWT_SECRET обязателен.
func Load(path string) (*Config, error) {
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				return nil, fmt.Errorf("failed to load env file %q: %w", path, err)
			}
		}
	}

	jwtCfg, err := jwt.New()
	if err != nil {
		return nil, err
	}

	return &Config{
		HTTP:   http.New(),
		MySQL:  mysql.New(),
		Redis:  redis.New(),
		JWT:    jwtCfg,
		Logger: logger.New(),
	}, nil
}
