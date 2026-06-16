package team

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
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

// newTeamHandler собирает хендлер с замоканным сервисом.
func newTeamHandler(t *testing.T) (*Handler, *mocks.MockTeamsService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockTeamsService(ctrl)
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

// --- Create ---

func TestCreateHandler_Success(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().
		Create(gomock.Any(), model.CreateTeamInput{Name: "Platform", Description: "core", OwnerID: 42}).
		Return(&model.Team{ID: 1, Name: "Platform", Description: "core", CreatedBy: 42}, nil)

	req := authedRequest(t, http.MethodPost, "/", `{"name":"Platform","description":"core"}`, 42, mgr)
	rec := serveAuthed(h.Create, req, mgr)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%q)", rec.Code, rec.Body.String())
	}
	var resp teamv1.TeamResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if resp.ID != 1 || resp.Name != "Platform" || resp.CreatedBy != 42 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestCreateHandler_Unauthorized(t *testing.T) {
	// Прямой вызов без auth-контекста: userID не находится → 401, сервис не вызывается.
	h, _ := newTeamHandler(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"X"}`))
	h.Create(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestCreateHandler_MalformedJSON(t *testing.T) {
	h, _ := newTeamHandler(t)
	mgr := newManager()

	req := authedRequest(t, http.MethodPost, "/", `{`, 42, mgr)
	rec := serveAuthed(h.Create, req, mgr)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed JSON, got %d", rec.Code)
	}
}

func TestCreateHandler_ValidationError(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().Create(gomock.Any(), gomock.Any()).
		Return(nil, errors.ErrValidation)

	req := authedRequest(t, http.MethodPost, "/", `{"name":""}`, 42, mgr)
	rec := serveAuthed(h.Create, req, mgr)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- List ---

func TestListHandler_Success(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().List(gomock.Any(), int64(42)).
		Return([]model.Team{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}}, nil)

	req := authedRequest(t, http.MethodGet, "/", "", 42, mgr)
	rec := serveAuthed(h.List, req, mgr)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%q)", rec.Code, rec.Body.String())
	}
	var resp teamv1.ListTeamsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if len(resp.Teams) != 2 {
		t.Errorf("expected 2 teams, got %d", len(resp.Teams))
	}
}

// --- Invite ---

// inviteRequest добавляет к authedRequest chi-параметр {id}.
func inviteRequest(t *testing.T, body, id string, uid int64, mgr *token.Manager) *http.Request {
	t.Helper()
	req := authedRequest(t, http.MethodPost, "/", body, uid, mgr)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestInviteHandler_Success(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().
		Invite(gomock.Any(), model.InviteInput{TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.RoleMember}).
		Return(nil)

	req := inviteRequest(t, `{"user_id":7}`, "1", 42, mgr)
	rec := serveAuthed(h.Invite, req, mgr)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%q)", rec.Code, rec.Body.String())
	}
	var resp teamv1.InviteResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if resp.TeamID != 1 || resp.UserID != 7 || resp.Role != "member" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestInviteHandler_RoleHonored(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().
		Invite(gomock.Any(), model.InviteInput{TeamID: 3, ActorID: 42, InviteeID: 7, Role: model.RoleAdmin}).
		Return(nil)

	req := inviteRequest(t, `{"user_id":7,"role":"admin"}`, "3", 42, mgr)
	rec := serveAuthed(h.Invite, req, mgr)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%q)", rec.Code, rec.Body.String())
	}
}

func TestInviteHandler_BadTeamID(t *testing.T) {
	h, _ := newTeamHandler(t)
	mgr := newManager()

	req := inviteRequest(t, `{"user_id":7}`, "abc", 42, mgr)
	rec := serveAuthed(h.Invite, req, mgr)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid team id, got %d", rec.Code)
	}
}

func TestInviteHandler_Forbidden(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().Invite(gomock.Any(), gomock.Any()).Return(errors.ErrForbidden)

	req := inviteRequest(t, `{"user_id":7}`, "1", 42, mgr)
	rec := serveAuthed(h.Invite, req, mgr)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
