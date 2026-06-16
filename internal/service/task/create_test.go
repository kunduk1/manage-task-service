package task

import (
	"context"
	stderrors "errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

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
