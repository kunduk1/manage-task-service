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

func TestList_Success(t *testing.T) {
	svc, taskRepo, _, teamRepo, _, cacheRepo := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleMember, nil)
	// Промах кэша → читаем из БД и затем кладём результат в кэш.
	cacheRepo.EXPECT().GetTaskList(gomock.Any(), int64(1), gomock.Any()).Return(nil, errors.ErrCacheMiss)
	taskRepo.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, f model.TaskFilter) ([]model.Task, error) {
			require.NotNil(t, f.TeamID)
			assert.Equal(t, int64(1), *f.TeamID)
			return []model.Task{{ID: 1}, {ID: 2}}, nil
		})
	cacheRepo.EXPECT().SetTaskList(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(nil)

	got, err := svc.List(context.Background(), model.TaskListQuery{ActorID: 42, TeamID: 1})
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestList_CacheHit(t *testing.T) {
	// Попадание в кэш: БД-репозиторий не вызывается (taskRepo.List без EXPECT).
	svc, _, _, teamRepo, _, cacheRepo := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleMember, nil)
	cacheRepo.EXPECT().GetTaskList(gomock.Any(), int64(1), gomock.Any()).
		Return([]model.Task{{ID: 1}, {ID: 2}, {ID: 3}}, nil)

	got, err := svc.List(context.Background(), model.TaskListQuery{ActorID: 42, TeamID: 1})
	require.NoError(t, err)
	assert.Len(t, got, 3)
}

func TestList_RequiresTeamID(t *testing.T) {
	svc, _, _, _, _, _ := newTestService(t)

	_, err := svc.List(context.Background(), model.TaskListQuery{ActorID: 42, TeamID: 0})
	assert.ErrorIs(t, err, errors.ErrValidation)
}

func TestList_ForbiddenNotMember(t *testing.T) {
	svc, _, _, teamRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(99)).
		Return(model.TeamRole(""), errors.ErrNotTeamMember)

	_, err := svc.List(context.Background(), model.TaskListQuery{ActorID: 99, TeamID: 1})
	assert.ErrorIs(t, err, errors.ErrForbidden)
}
