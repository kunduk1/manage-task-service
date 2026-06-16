package redis

import (
	"context"
	stderrors "errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// client — обёртка над *goredis.Client, реализующая cache.Client.
type client struct {
	rdb *goredis.Client
}

func (c *client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if stderrors.Is(err, goredis.Nil) {
		return "", errors.ErrCacheMiss
	}
	return val, err
}

func (c *client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (c *client) HSet(ctx context.Context, key string, values interface{}) error {
	return c.rdb.HSet(ctx, key, values).Err()
}

func (c *client) HGetAll(ctx context.Context, key string) ([]interface{}, error) {
	res, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	// go-redis отдаёт map[string]string; разворачиваем в [field, value, ...].
	out := make([]interface{}, 0, len(res)*2)
	for field, value := range res {
		out = append(out, field, value)
	}
	return out, nil
}

func (c *client) Del(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

func (c *client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.rdb.Expire(ctx, key, expiration).Err()
}

func (c *client) Close() error {
	return c.rdb.Close()
}
