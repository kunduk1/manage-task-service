package team

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

func TestCreateHandler_Success(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().
		Create(gomock.Any(), model.CreateTeamInput{Name: "Platform", Description: "core", OwnerID: 42}).
		Return(&model.Team{ID: 1, Name: "Platform", Description: "core", CreatedBy: 42}, nil)

	req := authedRequest(t, http.MethodPost, "/", `{"name":"Platform","description":"core"}`, 42, mgr)
	rec := serveAuthed(h.Create, req, mgr)

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp teamv1.TeamResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, int64(1), resp.ID)
	assert.Equal(t, "Platform", resp.Name)
	assert.Equal(t, int64(42), resp.CreatedBy)
}

func TestCreateHandler_Unauthorized(t *testing.T) {
	// Прямой вызов без auth-контекста: userID не находится → 401, сервис не вызывается.
	h, _ := newTeamHandler(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"X"}`))
	h.Create(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateHandler_MalformedJSON(t *testing.T) {
	h, _ := newTeamHandler(t)
	mgr := newManager()

	req := authedRequest(t, http.MethodPost, "/", `{`, 42, mgr)
	rec := serveAuthed(h.Create, req, mgr)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateHandler_ValidationError(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().Create(gomock.Any(), gomock.Any()).
		Return(nil, errors.ErrValidation)

	req := authedRequest(t, http.MethodPost, "/", `{"name":""}`, 42, mgr)
	rec := serveAuthed(h.Create, req, mgr)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
