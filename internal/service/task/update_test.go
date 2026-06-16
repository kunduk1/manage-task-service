package task

import (
	"context"
	stderrors "errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

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
