package team

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

func TestStatsHandler_Success(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().Stats(gomock.Any()).Return([]model.TeamStats{
		{TeamID: 1, Name: "Platform", MemberCount: 5, DoneLast7Days: 12},
	}, nil)

	req := authedRequest(t, http.MethodGet, "/", "", 42, mgr)
	rec := serveAuthed(h.Stats, req, mgr)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp teamv1.TeamStatsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Stats, 1)
	assert.Equal(t, int64(1), resp.Stats[0].TeamID)
	assert.Equal(t, "Platform", resp.Stats[0].Name)
	assert.Equal(t, int64(5), resp.Stats[0].MemberCount)
	assert.Equal(t, int64(12), resp.Stats[0].DoneLast7Days)
}

func TestStatsHandler_Unauthorized(t *testing.T) {
	h, _ := newTeamHandler(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.Stats(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestStatsHandler_ServiceError(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().Stats(gomock.Any()).Return(nil, stderrors.New("boom"))

	req := authedRequest(t, http.MethodGet, "/", "", 42, mgr)
	rec := serveAuthed(h.Stats, req, mgr)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
