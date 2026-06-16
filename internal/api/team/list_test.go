package team

import (
	"encoding/json"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

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
