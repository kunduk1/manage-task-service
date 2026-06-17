package task

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func TestCreate_Success(t *testing.T) {
	svc, taskRepo, _, teamRepo, _, cacheRepo := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleMember, nil)
	taskRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, tk *model.Task) (int64, error) {
			assert.Equal(t, "Write docs", tk.Title)
			assert.Equal(t, int64(42), tk.CreatedBy)
			assert.Equal(t, model.StatusTodo, tk.Status)
			return int64(5), nil
		})
	// Создание задачи инвалидирует кэш списка команды.
	cacheRepo.EXPECT().InvalidateTeam(gomock.Any(), int64(1)).Return(nil)
	taskRepo.EXPECT().GetByID(gomock.Any(), int64(5)).
		Return(&model.Task{ID: 5, TeamID: 1, Title: "Write docs", Status: model.StatusTodo, CreatedBy: 42}, nil)

	task, err := svc.Create(context.Background(), model.CreateTaskInput{
		ActorID: 42, TeamID: 1, Title: "  Write docs ",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(5), task.ID)
}

func TestCreate_AssigneeMustBeMember(t *testing.T) {
	svc, _, _, teamRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleMember, nil)
	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(7)).
		Return(model.TeamRole(""), errors.ErrNotTeamMember)

	_, err := svc.Create(context.Background(), model.CreateTaskInput{
		ActorID: 42, TeamID: 1, Title: "X", AssigneeID: ptrInt64(7),
	})
	assert.ErrorIs(t, err, errors.ErrValidation)
}

func TestCreate_Validation(t *testing.T) {
	// Невалидный ввод отсекается до обращения к репозиториям — никаких EXPECT.
	svc, _, _, _, _, _ := newTestService(t)

	cases := []model.CreateTaskInput{
		{ActorID: 1, TeamID: 1, Title: "   "},                                  // пустой title
		{ActorID: 1, TeamID: 0, Title: "X"},                                    // нет team_id
		{ActorID: 1, TeamID: 1, Title: "X", Status: model.TaskStatus("bogus")}, // неверный статус
	}
	for _, in := range cases {
		_, err := svc.Create(context.Background(), in)
		assert.ErrorIs(t, err, errors.ErrValidation)
	}
}

func TestCreate_ForbiddenNotMember(t *testing.T) {
	svc, _, _, teamRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(99)).
		Return(model.TeamRole(""), errors.ErrNotTeamMember)

	_, err := svc.Create(context.Background(), model.CreateTaskInput{ActorID: 99, TeamID: 1, Title: "X"})
	assert.ErrorIs(t, err, errors.ErrForbidden)
}
