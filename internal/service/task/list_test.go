package task

import (
	"context"
	stderrors "errors"
	"testing"

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
