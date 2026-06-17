package team

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

func TestTopCreatorsHandler_DefaultWindow(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	// from/to не заданы — хендлер подставляет окно по умолчанию (now-30d, now).
	svc.EXPECT().TopCreators(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]model.TopCreator{
			{TeamID: 1, UserID: 42, UserName: "Alice", CreatedCount: 7, Rank: 1},
		}, nil)

	req := authedRequest(t, http.MethodGet, "/", "", 42, mgr)
	rec := serveAuthed(h.TopCreators, req, mgr)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp teamv1.TopCreatorsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Creators, 1)
	assert.Equal(t, "Alice", resp.Creators[0].UserName)
	assert.True(t, resp.From.Before(resp.To))
}

func TestTopCreatorsHandler_ExplicitWindow(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	const from = "2026-05-01T00:00:00Z"
	const to = "2026-06-01T00:00:00Z"

	svc.EXPECT().TopCreators(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]model.TopCreator{}, nil)

	req := authedRequest(t, http.MethodGet, "/?from="+from+"&to="+to, "", 42, mgr)
	rec := serveAuthed(h.TopCreators, req, mgr)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp teamv1.TopCreatorsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	// Хендлер эхо-возвращает разобранное окно.
	assert.Equal(t, from, resp.From.Format(time.RFC3339))
	assert.Equal(t, to, resp.To.Format(time.RFC3339))
}

func TestTopCreatorsHandler_InvalidFrom(t *testing.T) {
	h, _ := newTeamHandler(t)
	mgr := newManager()

	req := authedRequest(t, http.MethodGet, "/?from=nope", "", 42, mgr)
	rec := serveAuthed(h.TopCreators, req, mgr)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTopCreatorsHandler_InvalidTo(t *testing.T) {
	h, _ := newTeamHandler(t)
	mgr := newManager()

	req := authedRequest(t, http.MethodGet, "/?to=nope", "", 42, mgr)
	rec := serveAuthed(h.TopCreators, req, mgr)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTopCreatorsHandler_FromNotBeforeTo(t *testing.T) {
	h, _ := newTeamHandler(t)
	mgr := newManager()

	// from позже to — невалидный диапазон.
	req := authedRequest(t, http.MethodGet,
		"/?from=2026-06-01T00:00:00Z&to=2026-05-01T00:00:00Z", "", 42, mgr)
	rec := serveAuthed(h.TopCreators, req, mgr)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTopCreatorsHandler_Unauthorized(t *testing.T) {
	h, _ := newTeamHandler(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.TopCreators(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestTopCreatorsHandler_ServiceError(t *testing.T) {
	h, svc := newTeamHandler(t)
	mgr := newManager()

	svc.EXPECT().TopCreators(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, stderrors.New("boom"))

	req := authedRequest(t, http.MethodGet, "/", "", 42, mgr)
	rec := serveAuthed(h.TopCreators, req, mgr)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
