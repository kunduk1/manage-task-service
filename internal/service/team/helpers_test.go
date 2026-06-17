package team

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zapcore"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	dbmocks "github.com/kunduk1/manage-task-service/internal/clients/db/mocks"
	emailmocks "github.com/kunduk1/manage-task-service/internal/clients/email/mocks"
	"github.com/kunduk1/manage-task-service/internal/logger"
	repomocks "github.com/kunduk1/manage-task-service/internal/repository/mocks"
	"github.com/kunduk1/manage-task-service/internal/service"
	"github.com/kunduk1/manage-task-service/internal/service/authz"
)

// Сервис best-effort логирует сбои отправки письма через глобальный логгер — в
// тестах он не инициализирован, поэтому ставим no-op, чтобы logger.Warn не паниковал.
func init() {
	logger.NewGlobalLogger(zapcore.NewNopCore())
}

// newTestService собирает сервис с gomock-моками репозиториев, менеджера
// транзакций и почтового клиента.
func newTestService(t *testing.T) (
	service.TeamsService,
	*repomocks.MockTeamRepository,
	*repomocks.MockUserRepository,
	*dbmocks.MockTxManager,
	*emailmocks.MockClient,
) {
	t.Helper()
	ctrl := gomock.NewController(t)
	teamRepo := repomocks.NewMockTeamRepository(ctrl)
	userRepo := repomocks.NewMockUserRepository(ctrl)
	txm := dbmocks.NewMockTxManager(ctrl)
	emailClient := emailmocks.NewMockClient(ctrl)
	svc := NewService(teamRepo, userRepo, txm, authz.New(teamRepo), emailClient)
	return svc, teamRepo, userRepo, txm, emailClient
}

// runTx — заглушка ReadCommit, исполняющая переданную функцию в том же контексте.
func runTx(txm *dbmocks.MockTxManager) {
	txm.EXPECT().ReadCommit(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, f db.Handler) error { return f(ctx) })
}
