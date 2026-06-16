package task

import (
	"context"
	stderrors "errors"
	"testing"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zapcore"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	dbmocks "github.com/kunduk1/manage-task-service/internal/clients/db/mocks"
	"github.com/kunduk1/manage-task-service/internal/logger"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomocks "github.com/kunduk1/manage-task-service/internal/repository/mocks"
	"github.com/kunduk1/manage-task-service/internal/service"
	"github.com/kunduk1/manage-task-service/pkg/errors"
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

// --- Create ---

func TestCreate_Success(t *testing.T) {
	svc, taskRepo, _, teamRepo, _, cacheRepo := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleMember, nil)
	taskRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, tk *model.Task) (int64, error) {
			if tk.Title != "Write docs" {
				t.Errorf("expected trimmed title, got %q", tk.Title)
			}
			if tk.CreatedBy != 42 {
				t.Errorf("expected created_by 42, got %d", tk.CreatedBy)
			}
			if tk.Status != model.StatusTodo {
				t.Errorf("expected default status todo, got %q", tk.Status)
			}
			return int64(5), nil
		})
	// Создание задачи инвалидирует кэш списка команды.
	cacheRepo.EXPECT().InvalidateTeam(gomock.Any(), int64(1)).Return(nil)
	taskRepo.EXPECT().GetByID(gomock.Any(), int64(5)).
		Return(&model.Task{ID: 5, TeamID: 1, Title: "Write docs", Status: model.StatusTodo, CreatedBy: 42}, nil)

	task, err := svc.Create(context.Background(), model.CreateTaskInput{
		ActorID: 42, TeamID: 1, Title: "  Write docs ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != 5 {
		t.Errorf("expected id from repository, got %d", task.ID)
	}
}

func TestCreate_AssigneeMustBeMember(t *testing.T) {
	svc, _, _, teamRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleMember, nil)
	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(7)).
		Return(model.TeamRole(""), errors.ErrNotTeamMember)

	_, err := svc.Create(context.Background(), model.CreateTaskInput{
		ActorID: 42, TeamID: 1, Title: "X", AssigneeID: ptrInt64(7),
	})
	if !stderrors.Is(err, errors.ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestCreate_Validation(t *testing.T) {
	// Невалидный ввод отсекается до обращения к репозиториям — никаких EXPECT.
	svc, _, _, _, _, _ := newTestService(t)

	cases := []model.CreateTaskInput{
		{ActorID: 1, TeamID: 1, Title: "   "},                                  // пустой title
		{ActorID: 1, TeamID: 0, Title: "X"},                                    // нет team_id
		{ActorID: 1, TeamID: 1, Title: "X", Status: model.TaskStatus("bogus")}, // неверный статус
	}
	for i, in := range cases {
		if _, err := svc.Create(context.Background(), in); !stderrors.Is(err, errors.ErrValidation) {
			t.Errorf("case %d: expected ErrValidation, got %v", i, err)
		}
	}
}

func TestCreate_ForbiddenNotMember(t *testing.T) {
	svc, _, _, teamRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(99)).
		Return(model.TeamRole(""), errors.ErrNotTeamMember)

	_, err := svc.Create(context.Background(), model.CreateTaskInput{ActorID: 99, TeamID: 1, Title: "X"})
	if !stderrors.Is(err, errors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// --- List ---

func TestList_Success(t *testing.T) {
	svc, taskRepo, _, teamRepo, _, cacheRepo := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleMember, nil)
	// Промах кэша → читаем из БД и затем кладём результат в кэш.
	cacheRepo.EXPECT().GetTaskList(gomock.Any(), int64(1), gomock.Any()).Return(nil, errors.ErrCacheMiss)
	taskRepo.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, f model.TaskFilter) ([]model.Task, error) {
			if f.TeamID == nil || *f.TeamID != 1 {
				t.Errorf("expected team filter 1, got %+v", f.TeamID)
			}
			return []model.Task{{ID: 1}, {ID: 2}}, nil
		})
	cacheRepo.EXPECT().SetTaskList(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(nil)

	got, err := svc.List(context.Background(), model.TaskListQuery{ActorID: 42, TeamID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(got))
	}
}

func TestList_CacheHit(t *testing.T) {
	// Попадание в кэш: БД-репозиторий не вызывается (taskRepo.List без EXPECT).
	svc, _, _, teamRepo, _, cacheRepo := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleMember, nil)
	cacheRepo.EXPECT().GetTaskList(gomock.Any(), int64(1), gomock.Any()).
		Return([]model.Task{{ID: 1}, {ID: 2}, {ID: 3}}, nil)

	got, err := svc.List(context.Background(), model.TaskListQuery{ActorID: 42, TeamID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 cached tasks, got %d", len(got))
	}
}

func TestList_RequiresTeamID(t *testing.T) {
	svc, _, _, _, _, _ := newTestService(t)

	_, err := svc.List(context.Background(), model.TaskListQuery{ActorID: 42, TeamID: 0})
	if !stderrors.Is(err, errors.ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestList_ForbiddenNotMember(t *testing.T) {
	svc, _, _, teamRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(99)).
		Return(model.TeamRole(""), errors.ErrNotTeamMember)

	_, err := svc.List(context.Background(), model.TaskListQuery{ActorID: 99, TeamID: 1})
	if !stderrors.Is(err, errors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// --- Update ---

func TestUpdate_StatusChangeRecordsHistory(t *testing.T) {
	svc, taskRepo, historyRepo, teamRepo, txm, cacheRepo := newTestService(t)

	cur := &model.Task{ID: 10, TeamID: 1, Title: "T", Status: model.StatusTodo, CreatedBy: 42}
	taskRepo.EXPECT().GetByID(gomock.Any(), int64(10)).Return(cur, nil)
	// Инициатор — создатель (роль member, но CreatedBy совпадает).
	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleMember, nil)

	runTx(txm)
	taskRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, tk *model.Task) error {
			if tk.Status != model.StatusInProgress {
				t.Errorf("expected status in_progress, got %q", tk.Status)
			}
			return nil
		})
	historyRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, e *model.TaskHistoryEntry) (int64, error) {
			if e.Field != "status" || e.ChangedBy != 42 {
				t.Errorf("unexpected history entry: %+v", e)
			}
			if e.OldValue == nil || *e.OldValue != "todo" || e.NewValue == nil || *e.NewValue != "in_progress" {
				t.Errorf("expected todo->in_progress, got old=%v new=%v", e.OldValue, e.NewValue)
			}
			return int64(1), nil
		})
	// Обновление задачи инвалидирует кэш списка команды.
	cacheRepo.EXPECT().InvalidateTeam(gomock.Any(), int64(1)).Return(nil)
	taskRepo.EXPECT().GetByID(gomock.Any(), int64(10)).
		Return(&model.Task{ID: 10, TeamID: 1, Title: "T", Status: model.StatusInProgress, CreatedBy: 42}, nil)

	task, err := svc.Update(context.Background(), model.UpdateTaskInput{
		ActorID: 42, TaskID: 10, Status: ptrStatus(model.StatusInProgress),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != model.StatusInProgress {
		t.Errorf("expected updated status, got %q", task.Status)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	svc, taskRepo, _, _, _, _ := newTestService(t)

	taskRepo.EXPECT().GetByID(gomock.Any(), int64(10)).Return(nil, errors.ErrTaskNotFound)

	_, err := svc.Update(context.Background(), model.UpdateTaskInput{ActorID: 42, TaskID: 10})
	if !stderrors.Is(err, errors.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestUpdate_ForbiddenPlainMember(t *testing.T) {
	// Рядовой член, не создатель и не исполнитель → 403.
	svc, taskRepo, _, teamRepo, _, _ := newTestService(t)

	cur := &model.Task{ID: 10, TeamID: 1, Status: model.StatusTodo, CreatedBy: 42, AssigneeID: ptrInt64(7)}
	taskRepo.EXPECT().GetByID(gomock.Any(), int64(10)).Return(cur, nil)
	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(99)).Return(model.RoleMember, nil)

	_, err := svc.Update(context.Background(), model.UpdateTaskInput{
		ActorID: 99, TaskID: 10, Status: ptrStatus(model.StatusDone),
	})
	if !stderrors.Is(err, errors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdate_AssigneeCanUpdate(t *testing.T) {
	// Исполнитель (роль member) вправе обновлять, история не пишется при отсутствии изменений.
	svc, taskRepo, _, teamRepo, txm, cacheRepo := newTestService(t)

	cur := &model.Task{ID: 10, TeamID: 1, Title: "T", Status: model.StatusTodo, CreatedBy: 42, AssigneeID: ptrInt64(7)}
	taskRepo.EXPECT().GetByID(gomock.Any(), int64(10)).Return(cur, nil)
	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(7)).Return(model.RoleMember, nil)

	runTx(txm)
	taskRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	cacheRepo.EXPECT().InvalidateTeam(gomock.Any(), int64(1)).Return(nil)
	taskRepo.EXPECT().GetByID(gomock.Any(), int64(10)).Return(cur, nil)

	// Передаём тот же статус — изменений нет, historyRepo.Create НЕ ожидается.
	_, err := svc.Update(context.Background(), model.UpdateTaskInput{
		ActorID: 7, TaskID: 10, Status: ptrStatus(model.StatusTodo),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdate_InvalidStatus(t *testing.T) {
	svc, taskRepo, _, teamRepo, _, _ := newTestService(t)

	cur := &model.Task{ID: 10, TeamID: 1, Status: model.StatusTodo, CreatedBy: 42}
	taskRepo.EXPECT().GetByID(gomock.Any(), int64(10)).Return(cur, nil)
	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)

	_, err := svc.Update(context.Background(), model.UpdateTaskInput{
		ActorID: 42, TaskID: 10, Status: ptrStatus(model.TaskStatus("bogus")),
	})
	if !stderrors.Is(err, errors.ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

// --- History ---

func TestHistory_Success(t *testing.T) {
	svc, taskRepo, historyRepo, teamRepo, _, _ := newTestService(t)

	taskRepo.EXPECT().GetByID(gomock.Any(), int64(10)).
		Return(&model.Task{ID: 10, TeamID: 1}, nil)
	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleMember, nil)
	historyRepo.EXPECT().ListByTask(gomock.Any(), int64(10)).
		Return([]model.TaskHistoryEntry{{ID: 1, Field: "status"}}, nil)

	got, err := svc.History(context.Background(), model.TaskHistoryQuery{ActorID: 42, TaskID: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 entry, got %d", len(got))
	}
}

func TestHistory_ForbiddenNotMember(t *testing.T) {
	svc, taskRepo, _, teamRepo, _, _ := newTestService(t)

	taskRepo.EXPECT().GetByID(gomock.Any(), int64(10)).Return(&model.Task{ID: 10, TeamID: 1}, nil)
	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(99)).
		Return(model.TeamRole(""), errors.ErrNotTeamMember)

	_, err := svc.History(context.Background(), model.TaskHistoryQuery{ActorID: 99, TaskID: 10})
	if !stderrors.Is(err, errors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
