package task

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zapcore"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	dbmocks "github.com/kunduk1/manage-task-service/internal/clients/db/mocks"
	"github.com/kunduk1/manage-task-service/internal/logger"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomocks "github.com/kunduk1/manage-task-service/internal/repository/mocks"
	"github.com/kunduk1/manage-task-service/internal/service"
)

// Сервис best-effort логирует сбои кэша через глобальный логгер — в тестах он не
// инициализирован, поэтому ставим no-op, чтобы logger.Warn не паниковал на nil.
func init() {
	logger.NewGlobalLogger(zapcore.NewNopCore())
}

func newTestService(t *testing.T) (
	service.TasksService,
	*repomocks.MockTaskRepository,
	*repomocks.MockTaskHistoryRepository,
	*repomocks.MockTeamRepository,
	*dbmocks.MockTxManager,
	*repomocks.MockTaskCacheRepository,
) {
	t.Helper()
	ctrl := gomock.NewController(t)
	taskRepo := repomocks.NewMockTaskRepository(ctrl)
	historyRepo := repomocks.NewMockTaskHistoryRepository(ctrl)
	teamRepo := repomocks.NewMockTeamRepository(ctrl)
	cacheRepo := repomocks.NewMockTaskCacheRepository(ctrl)
	txm := dbmocks.NewMockTxManager(ctrl)
	svc := NewService(taskRepo, historyRepo, teamRepo, cacheRepo, txm)
	return svc, taskRepo, historyRepo, teamRepo, txm, cacheRepo
}

// runTx — заглушка ReadCommit, исполняющая переданную функцию в том же контексте.
func runTx(txm *dbmocks.MockTxManager) {
	txm.EXPECT().ReadCommit(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, f db.Handler) error { return f(ctx) })
}

func ptrStatus(s model.TaskStatus) *model.TaskStatus { return &s }
func ptrInt64(v int64) *int64                        { return &v }
