package task

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

func TestUpdateHandler_Success(t *testing.T) {
	h, svc := newTaskHandler(t)
	mgr := newManager()

	status := model.StatusDone
	svc.EXPECT().
		Update(gomock.Any(), model.UpdateTaskInput{ActorID: 42, TaskID: 10, Status: &status}).
		Return(&model.Task{ID: 10, TeamID: 1, Status: model.StatusDone, CreatedBy: 42}, nil)

	req := withURLParam(authedRequest(t, http.MethodPut, "/", `{"status":"done"}`, 42, mgr), "id", "10")
	rec := serveAuthed(h.Update, req, mgr)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp taskv1.TaskResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "done", resp.Status)
}

func TestUpdateHandler_BadID(t *testing.T) {
	h, _ := newTaskHandler(t)
	mgr := newManager()

	req := withURLParam(authedRequest(t, http.MethodPut, "/", `{"status":"done"}`, 42, mgr), "id", "abc")
	rec := serveAuthed(h.Update, req, mgr)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateHandler_NotFound(t *testing.T) {
	h, svc := newTaskHandler(t)
	mgr := newManager()

	svc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.ErrTaskNotFound)

	req := withURLParam(authedRequest(t, http.MethodPut, "/", `{"status":"done"}`, 42, mgr), "id", "10")
	rec := serveAuthed(h.Update, req, mgr)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
