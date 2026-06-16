package task

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/kunduk1/manage-task-service/internal/api/task/converter"
	"github.com/kunduk1/manage-task-service/internal/transport/request"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

// @Summary      Update a task
// @Description  Partially updates a task (only provided fields change; assignee_id=0 unassigns). Allowed for the creator, current assignee, or team owner/admin. Records field-level history.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                       true  "task id"
// @Param        request  body      taskv1.UpdateTaskRequest  true  "fields to update"
// @Success      200      {object}  taskv1.TaskResponse
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      403      {object}  response.ErrorBody
// @Failure      404      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /tasks/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req taskv1.UpdateTaskRequest
	if !request.DecodeJSON(w, r, &req) {
		return
	}

	task, err := h.svc.Update(r.Context(), converter.ToUpdateInput(req, taskID, uid))
	if err != nil {
		response.ServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, converter.ToTaskResponse(task))
}
