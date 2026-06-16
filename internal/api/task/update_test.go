package task

import (
	"encoding/json"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

func TestUpdateHandler_Success(t *testing.T) {
	h, svc := newTaskHandler(t)
	mgr := newManager()

	status := model.StatusDone
	svc.EXPECT().
		Update(gomock.Any(), model.UpdateTaskInput{ActorID: 42, TaskID: 10, Status: &status}).
		Return(&model.Task{ID: 10, TeamID: 1, Status: model.StatusDone, CreatedBy: 42}, nil)

	req := withURLParam(authedRequest(t, http.MethodPut, "/", `{"status":"done"}`, 42, mgr), "id", "10")
	rec := serveAuthed(h.Update, req, mgr)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%q)", rec.Code, rec.Body.String())
	}
	var resp taskv1.TaskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if resp.Status != "done" {
		t.Errorf("expected status done, got %q", resp.Status)
	}
}

func TestUpdateHandler_BadID(t *testing.T) {
	h, _ := newTaskHandler(t)
	mgr := newManager()

	req := withURLParam(authedRequest(t, http.MethodPut, "/", `{"status":"done"}`, 42, mgr), "id", "abc")
	rec := serveAuthed(h.Update, req, mgr)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid task id, got %d", rec.Code)
	}
}

func TestUpdateHandler_NotFound(t *testing.T) {
	h, svc := newTaskHandler(t)
	mgr := newManager()

	svc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.ErrTaskNotFound)

	req := withURLParam(authedRequest(t, http.MethodPut, "/", `{"status":"done"}`, 42, mgr), "id", "10")
	rec := serveAuthed(h.Update, req, mgr)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
