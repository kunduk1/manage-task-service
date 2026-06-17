package task

import (
	"net/http"

	"github.com/kunduk1/manage-task-service/internal/api/task/converter"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

// @Summary      Misassigned tasks
// @Description  Integrity check: returns tasks whose assignee is not a member of the task's team.
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  taskv1.MisassignedTasksResponse
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /tasks/misassigned [get]
func (h *Handler) Misassigned(w http.ResponseWriter, r *http.Request) {
	if _, ok := userID(r); !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tasks, err := h.svc.Misassigned(r.Context())
	if err != nil {
		response.ServiceError(w, err)
		return
	}

	resp := taskv1.MisassignedTasksResponse{Tasks: converter.ToTaskResponses(tasks)}
	response.JSON(w, http.StatusOK, resp)
}
