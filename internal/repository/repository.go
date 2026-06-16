package repository

//go:generate go tool mockgen -source=repository.go -destination=mocks/repository_mock.go -package=mocks

import (
	"context"
	"time"

	"github.com/kunduk1/manage-task-service/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (int64, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type TokenRepository interface {
	SaveRefresh(ctx context.Context, token string, userID int64, ttl time.Duration) error
	GetRefresh(ctx context.Context, token string) (int64, error)
	DeleteRefresh(ctx context.Context, token string) error
}
