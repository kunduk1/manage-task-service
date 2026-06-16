package team

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
