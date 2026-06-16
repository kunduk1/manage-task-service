// Package auth реализует бизнес-логику аутентификации.
package auth

import (
	"strings"
	"time"

	"github.com/kunduk1/manage-task-service/internal/repository"
	"github.com/kunduk1/manage-task-service/internal/service"
	"github.com/kunduk1/manage-task-service/internal/token"
)

type serv struct {
	userRepo   repository.UserRepository
	tokenRepo  repository.TokenRepository
	jwtManager *token.Manager
	refreshTTL time.Duration
}

func NewService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwtManager *token.Manager,
	refreshTTL time.Duration,
) service.AuthService {
	return &serv{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
		refreshTTL: refreshTTL,
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
