package token

import (
	"context"
	stderrors "errors"
	"strconv"

	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func (r *repo) GetRefresh(ctx context.Context, token string) (int64, error) {
	val, err := r.client.Get(ctx, key(token))
	if err != nil {
		if stderrors.Is(err, errors.ErrCacheMiss) {
			return 0, errors.ErrRefreshTokenNotFound
		}
		return 0, err
	}

	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
