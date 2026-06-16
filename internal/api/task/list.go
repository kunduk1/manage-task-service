package task

import (
	"net/http"
	"strconv"

	"github.com/kunduk1/manage-task-service/internal/api/task/converter"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

// @Summary      List tasks
// @Description  Lists a team's tasks with optional status/assignee filters and pagination. Caller must be a team member.
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        team_id      query     int     true   "team id"
// @Param        status       query     string  false  "filter by status (todo, in_progress, done)"
// @Param        assignee_id  query     int     false  "filter by assignee id"
// @Param        limit        query     int     false  "page size (default 20, max 100)"
// @Param        offset       query     int     false  "page offset (default 0)"
// @Success      200          {object}  taskv1.ListTasksResponse
// @Failure      400          {object}  response.ErrorBody
// @Failure      401          {object}  response.ErrorBody
// @Failure      403          {object}  response.ErrorBody
// @Failure      500          {object}  response.ErrorBody
// @Router       /tasks [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()

	teamID, err := strconv.ParseInt(q.Get("team_id"), 10, 64)
	if err != nil || teamID <= 0 {
		response.Error(w, http.StatusBadRequest, "invalid or missing team_id")
		return
	}

	rawLimit := parseIntDefault(q.Get("limit"), 0)
	rawOffset := parseIntDefault(q.Get("offset"), 0)

	query := model.TaskListQuery{
		ActorID: uid,
		TeamID:  teamID,
		Limit:   rawLimit,
		Offset:  rawOffset,
	}

	if s := q.Get("status"); s != "" {
		st := model.TaskStatus(s)
		query.Status = &st
	}
	if a := q.Get("assignee_id"); a != "" {
		assigneeID, err := strconv.ParseInt(a, 10, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid assignee_id")
			return
		}
		query.AssigneeID = &assigneeID
	}

	tasks, err := h.svc.List(r.Context(), query)
	if err != nil {
		response.ServiceError(w, err)
		return
	}

	resp := taskv1.ListTasksResponse{
		Tasks:  converter.ToTaskResponses(tasks),
		Limit:  clampLimit(rawLimit),
		Offset: clampOffset(rawOffset),
	}
	response.JSON(w, http.StatusOK, resp)
}
