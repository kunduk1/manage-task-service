// Package token — кэш-репозиторий refresh-токенов.
package token

import (
	"github.com/kunduk1/manage-task-service/internal/clients/cache"
	"github.com/kunduk1/manage-task-service/internal/repository"
)

const keyPrefix = "refresh:"

type repo struct {
	client cache.Client
}

func NewRepository(client cache.Client) repository.TokenRepository {
	return &repo{client: client}
}

func key(token string) string {
	return keyPrefix + token
}
