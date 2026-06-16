package task

import (
	"encoding/json"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

func TestListHandler_Success(t *testing.T) {
	h, svc := newTaskHandler(t)
	mgr := newManager()

	status := model.StatusTodo
	svc.EXPECT().
		List(gomock.Any(), model.TaskListQuery{ActorID: 42, TeamID: 1, Status: &status, Limit: 5}).
		Return([]model.Task{{ID: 1}, {ID: 2}}, nil)

	req := authedRequest(t, http.MethodGet, "/?team_id=1&status=todo&limit=5", "", 42, mgr)
	rec := serveAuthed(h.List, req, mgr)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%q)", rec.Code, rec.Body.String())
	}
	var resp taskv1.ListTasksResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if len(resp.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(resp.Tasks))
	}
	if resp.Limit != 5 || resp.Offset != 0 {
		t.Errorf("expected echoed limit=5 offset=0, got limit=%d offset=%d", resp.Limit, resp.Offset)
	}
}

func TestListHandler_MissingTeamID(t *testing.T) {
	h, _ := newTaskHandler(t)
	mgr := newManager()

	req := authedRequest(t, http.MethodGet, "/", "", 42, mgr)
	rec := serveAuthed(h.List, req, mgr)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing team_id, got %d", rec.Code)
	}
}

func TestListHandler_InvalidAssignee(t *testing.T) {
	h, _ := newTaskHandler(t)
	mgr := newManager()

	req := authedRequest(t, http.MethodGet, "/?team_id=1&assignee_id=abc", "", 42, mgr)
	rec := serveAuthed(h.List, req, mgr)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid assignee_id, got %d", rec.Code)
	}
}
