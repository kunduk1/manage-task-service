package token

import (
	"context"
	"time"
)

func (r *repo) SaveRefresh(ctx context.Context, token string, userID int64, ttl time.Duration) error {
	return r.client.Set(ctx, key(token), userID, ttl)
}
