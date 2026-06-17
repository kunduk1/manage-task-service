package task

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
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

func TestMisassignedHandler_Success(t *testing.T) {
	h, svc := newTaskHandler(t)
	mgr := newManager()

	svc.EXPECT().Misassigned(gomock.Any()).
		Return([]model.Task{{ID: 1, TeamID: 1}, {ID: 2, TeamID: 1}}, nil)

	req := authedRequest(t, http.MethodGet, "/", "", 42, mgr)
	rec := serveAuthed(h.Misassigned, req, mgr)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp taskv1.MisassignedTasksResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Tasks, 2)
	assert.Equal(t, int64(1), resp.Tasks[0].ID)
	assert.Equal(t, int64(2), resp.Tasks[1].ID)
}

func TestMisassignedHandler_Unauthorized(t *testing.T) {
	h, _ := newTaskHandler(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.Misassigned(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMisassignedHandler_ServiceError(t *testing.T) {
	h, svc := newTaskHandler(t)
	mgr := newManager()

	svc.EXPECT().Misassigned(gomock.Any()).Return(nil, stderrors.New("boom"))

	req := authedRequest(t, http.MethodGet, "/", "", 42, mgr)
	rec := serveAuthed(h.Misassigned, req, mgr)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
