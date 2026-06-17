package task

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

func TestListHandler_Success(t *testing.T) {
	h, svc := newTaskHandler(t)
	mgr := newManager()

	status := model.StatusTodo
	svc.EXPECT().
		List(gomock.Any(), model.TaskListQuery{ActorID: 42, TeamID: 1, Status: &status, Limit: 5}).
		Return([]model.Task{{ID: 1}, {ID: 2}}, nil)

	req := authedRequest(t, http.MethodGet, "/?team_id=1&status=todo&limit=5", "", 42, mgr)
	rec := serveAuthed(h.List, req, mgr)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp taskv1.ListTasksResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Tasks, 2)
	assert.Equal(t, 5, resp.Limit)
	assert.Equal(t, 0, resp.Offset)
}

func TestListHandler_MissingTeamID(t *testing.T) {
	h, _ := newTaskHandler(t)
	mgr := newManager()

	req := authedRequest(t, http.MethodGet, "/", "", 42, mgr)
	rec := serveAuthed(h.List, req, mgr)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListHandler_InvalidAssignee(t *testing.T) {
	h, _ := newTaskHandler(t)
	mgr := newManager()

	req := authedRequest(t, http.MethodGet, "/?team_id=1&assignee_id=abc", "", 42, mgr)
	rec := serveAuthed(h.List, req, mgr)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
