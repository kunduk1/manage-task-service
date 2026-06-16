package task

import (
	"net/http"

	"github.com/kunduk1/manage-task-service/internal/api/task/converter"
	"github.com/kunduk1/manage-task-service/internal/transport/request"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

// @Summary      Create a task
// @Description  Creates a task in a team. Only a member of the team may create. Status defaults to todo.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      taskv1.CreateTaskRequest  true  "task payload"
// @Success      201      {object}  taskv1.TaskResponse
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      403      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /tasks [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req taskv1.CreateTaskRequest
	if !request.DecodeJSON(w, r, &req) {
		return
	}

	task, err := h.svc.Create(r.Context(), converter.ToCreateInput(req, uid))
	if err != nil {
		response.ServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, converter.ToTaskResponse(task))
}
