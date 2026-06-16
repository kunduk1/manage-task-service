package task

import (
	"context"
	stderrors "errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

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
