package task

import (
	"encoding/json"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

func TestHistoryHandler_Success(t *testing.T) {
	h, svc := newTaskHandler(t)
	mgr := newManager()

	old, new := "todo", "in_progress"
	svc.EXPECT().
		History(gomock.Any(), model.TaskHistoryQuery{ActorID: 42, TaskID: 10}).
		Return([]model.TaskHistoryEntry{{ID: 1, TaskID: 10, Field: "status", OldValue: &old, NewValue: &new}}, nil)

	req := withURLParam(authedRequest(t, http.MethodGet, "/", "", 42, mgr), "id", "10")
	rec := serveAuthed(h.History, req, mgr)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%q)", rec.Code, rec.Body.String())
	}
	var resp taskv1.TaskHistoryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if len(resp.Entries) != 1 || resp.Entries[0].Field != "status" {
		t.Errorf("unexpected history: %+v", resp.Entries)
	}
}

func TestHistoryHandler_BadID(t *testing.T) {
	h, _ := newTaskHandler(t)
	mgr := newManager()

	req := withURLParam(authedRequest(t, http.MethodGet, "/", "", 42, mgr), "id", "abc")
	rec := serveAuthed(h.History, req, mgr)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid task id, got %d", rec.Code)
	}
}
