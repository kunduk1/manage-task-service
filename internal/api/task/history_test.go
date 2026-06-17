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

func TestHistoryHandler_Success(t *testing.T) {
	h, svc := newTaskHandler(t)
	mgr := newManager()

	old, new := "todo", "in_progress"
	svc.EXPECT().
		History(gomock.Any(), model.TaskHistoryQuery{ActorID: 42, TaskID: 10}).
		Return([]model.TaskHistoryEntry{{ID: 1, TaskID: 10, Field: "status", OldValue: &old, NewValue: &new}}, nil)

	req := withURLParam(authedRequest(t, http.MethodGet, "/", "", 42, mgr), "id", "10")
	rec := serveAuthed(h.History, req, mgr)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp taskv1.TaskHistoryResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Entries, 1)
	assert.Equal(t, "status", resp.Entries[0].Field)
}

func TestHistoryHandler_BadID(t *testing.T) {
	h, _ := newTaskHandler(t)
	mgr := newManager()

	req := withURLParam(authedRequest(t, http.MethodGet, "/", "", 42, mgr), "id", "abc")
	rec := serveAuthed(h.History, req, mgr)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
