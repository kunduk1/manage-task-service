package task

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/internal/service/mocks"
	"github.com/kunduk1/manage-task-service/internal/token"
	"github.com/kunduk1/manage-task-service/internal/transport/middleware"
	"github.com/kunduk1/manage-task-service/pkg/errors"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

// newTaskHandler собирает хендлер с замоканным сервисом.
func newTaskHandler(t *testing.T) (*Handler, *mocks.MockTasksService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockTasksService(ctrl)
	return NewHandler(svc), svc
}

func newManager() *token.Manager {
	return token.NewManager("test-secret", time.Hour)
}

// authedRequest формирует запрос с валидным access-токеном пользователя uid.
func authedRequest(t *testing.T, method, target, body string, uid int64, mgr *token.Manager) *http.Request {
	t.Helper()
	tok, _, err := mgr.GenerateAccess(uid, "user@example.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	return req
}

// serveAuthed прогоняет запрос через middleware.Auth и хендлер.
func serveAuthed(h http.HandlerFunc, req *http.Request, mgr *token.Manager) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	middleware.Auth(mgr)(h).ServeHTTP(rec, req)
	return rec
}

// withURLParam добавляет к запросу chi-параметр {id}.
func withURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// --- Create ---

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

// --- List ---

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

// --- Update ---

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

// --- History ---

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
