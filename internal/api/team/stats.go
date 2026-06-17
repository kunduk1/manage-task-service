package team

import (
	"net/http"

	"github.com/kunduk1/manage-task-service/internal/api/team/converter"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

// @Summary      Team stats
// @Description  Returns, for every team, its name, member count and number of tasks moved to done in the last 7 days.
// @Tags         teams
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  teamv1.TeamStatsResponse
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /teams/stats [get]
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	if _, ok := userID(r); !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	stats, err := h.svc.Stats(r.Context())
	if err != nil {
		response.ServiceError(w, err)
		return
	}

	resp := teamv1.TeamStatsResponse{Stats: converter.ToTeamStatsResponses(stats)}
	response.JSON(w, http.StatusOK, resp)
}
