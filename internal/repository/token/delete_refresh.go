package token

import "context"

func (r *repo) DeleteRefresh(ctx context.Context, token string) error {
	return r.client.Del(ctx, key(token))
}
