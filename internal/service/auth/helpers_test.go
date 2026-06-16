package auth

import (
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	repomocks "github.com/kunduk1/manage-task-service/internal/repository/mocks"
	"github.com/kunduk1/manage-task-service/internal/service"
	"github.com/kunduk1/manage-task-service/internal/token"
)

// newTestService собирает сервис с gomock-моками репозиториев.
func newTestService(t *testing.T) (service.AuthService, *repomocks.MockUserRepository, *repomocks.MockTokenRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	userRepo := repomocks.NewMockUserRepository(ctrl)
	tokenRepo := repomocks.NewMockTokenRepository(ctrl)
	jwt := token.NewManager("test-secret", 15*time.Minute)
	svc := NewService(userRepo, tokenRepo, jwt, time.Hour)
	return svc, userRepo, tokenRepo
}
