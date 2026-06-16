package cache

//go:generate go tool mockgen -source=client.go -destination=mocks/client_mock.go -package=mocks

import (
	"context"
	"time"
)

// Client — абстракция кэша
type Client interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Incr(ctx context.Context, key string) (int64, error)
	HSet(ctx context.Context, key string, values interface{}) error
	HGetAll(ctx context.Context, key string) ([]interface{}, error)
	Del(ctx context.Context, key string) error
	Ping(ctx context.Context) error
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Close() error
}
