package team

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/internal/token"
	"github.com/kunduk1/manage-task-service/pkg/errors"
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

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
