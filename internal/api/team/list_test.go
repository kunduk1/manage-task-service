package team

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	require.Equal(t, http.StatusOK, rec.Code)
	var resp teamv1.ListTeamsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Teams, 2)
}
