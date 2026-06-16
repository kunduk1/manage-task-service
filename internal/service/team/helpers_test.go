package team

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	dbmocks "github.com/kunduk1/manage-task-service/internal/clients/db/mocks"
	repomocks "github.com/kunduk1/manage-task-service/internal/repository/mocks"
	"github.com/kunduk1/manage-task-service/internal/service"
)

// newTestService собирает сервис с gomock-моками репозиториев и менеджера транзакций.
func newTestService(t *testing.T) (
	service.TeamsService,
	*repomocks.MockTeamRepository,
	*repomocks.MockUserRepository,
	*dbmocks.MockTxManager,
) {
	t.Helper()
	ctrl := gomock.NewController(t)
	teamRepo := repomocks.NewMockTeamRepository(ctrl)
	userRepo := repomocks.NewMockUserRepository(ctrl)
	txm := dbmocks.NewMockTxManager(ctrl)
	svc := NewService(teamRepo, userRepo, txm)
	return svc, teamRepo, userRepo, txm
}

// runTx — заглушка ReadCommit, исполняющая переданную функцию в том же контексте.
func runTx(txm *dbmocks.MockTxManager) {
	txm.EXPECT().ReadCommit(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, f db.Handler) error { return f(ctx) })
}
