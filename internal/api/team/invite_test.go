package team

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp teamv1.InviteResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, int64(1), resp.TeamID)
	assert.Equal(t, int64(7), resp.UserID)
	assert.Equal(t, "member", resp.Role)
}

func TestInviteHandler_RoleHonored(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().
		Invite(gomock.Any(), model.InviteInput{TeamID: 3, ActorID: 42, InviteeID: 7, Role: model.RoleAdmin}).
		Return(nil)

	req := inviteRequest(t, `{"user_id":7,"role":"admin"}`, "3", 42, mgr)
	rec := serveAuthed(h.Invite, req, mgr)

	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestInviteHandler_BadTeamID(t *testing.T) {
	h, _ := newTeamHandler(t)
	mgr := newManager()

	req := inviteRequest(t, `{"user_id":7}`, "abc", 42, mgr)
	rec := serveAuthed(h.Invite, req, mgr)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestInviteHandler_Forbidden(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().Invite(gomock.Any(), gomock.Any()).Return(errors.ErrForbidden)

	req := inviteRequest(t, `{"user_id":7}`, "1", 42, mgr)
	rec := serveAuthed(h.Invite, req, mgr)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}
