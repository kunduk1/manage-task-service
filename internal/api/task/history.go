package task

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/kunduk1/manage-task-service/internal/api/task/converter"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

// @Summary      Task change history
// @Description  Returns the task's field-level change history, newest first. Caller must be a team member.
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "task id"
// @Success      200  {object}  taskv1.TaskHistoryResponse
// @Failure      400  {object}  response.ErrorBody
// @Failure      401  {object}  response.ErrorBody
// @Failure      403  {object}  response.ErrorBody
// @Failure      404  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /tasks/{id}/history [get]
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
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

	entries, err := h.svc.History(r.Context(), model.TaskHistoryQuery{ActorID: uid, TaskID: taskID})
	if err != nil {
		response.ServiceError(w, err)
		return
	}

	resp := taskv1.TaskHistoryResponse{Entries: converter.ToHistoryResponses(entries)}
	response.JSON(w, http.StatusOK, resp)
}
