package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/kunduk1/manage-task-service/internal/clients/cache"
	"github.com/kunduk1/manage-task-service/internal/config"
)

// NewClient создаёт и проверяет (ping) подключение к Redis и возвращает его
// за абстракцией cache.Client.
func NewClient(ctx context.Context, cfg config.RedisConfig) (cache.Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password(),
		DB:       cfg.DB(),
		// Параметры пула соединений. Нулевые значения => go-redis применяет
		// свои дефолты (PoolSize=10*GOMAXPROCS, без min idle, без max lifetime).
		PoolSize:        cfg.PoolSize(),
		MinIdleConns:    cfg.MinIdleConns(),
		ConnMaxLifetime: cfg.ConnMaxLifetime(),
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &client{rdb: rdb}, nil
}
