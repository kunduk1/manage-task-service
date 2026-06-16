package task

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

func TestCreateHandler_Success(t *testing.T) {
	h, svc := newTaskHandler(t)
	mgr := newManager()

	svc.EXPECT().
		Create(gomock.Any(), model.CreateTaskInput{ActorID: 42, TeamID: 1, Title: "X"}).
		Return(&model.Task{ID: 5, TeamID: 1, Title: "X", Status: model.StatusTodo, CreatedBy: 42}, nil)

	req := authedRequest(t, http.MethodPost, "/", `{"team_id":1,"title":"X"}`, 42, mgr)
	rec := serveAuthed(h.Create, req, mgr)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%q)", rec.Code, rec.Body.String())
	}
	var resp taskv1.TaskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if resp.ID != 5 || resp.Status != "todo" || resp.CreatedBy != 42 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestCreateHandler_Unauthorized(t *testing.T) {
	h, _ := newTaskHandler(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"team_id":1,"title":"X"}`))
	h.Create(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestCreateHandler_MalformedJSON(t *testing.T) {
	h, _ := newTaskHandler(t)
	mgr := newManager()

	req := authedRequest(t, http.MethodPost, "/", `{`, 42, mgr)
	rec := serveAuthed(h.Create, req, mgr)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed JSON, got %d", rec.Code)
	}
}

func TestCreateHandler_Forbidden(t *testing.T) {
	h, svc := newTaskHandler(t)
	mgr := newManager()

	svc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.ErrForbidden)

	req := authedRequest(t, http.MethodPost, "/", `{"team_id":1,"title":"X"}`, 42, mgr)
	rec := serveAuthed(h.Create, req, mgr)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
