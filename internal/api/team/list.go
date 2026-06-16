package team

import (
	"net/http"

	"github.com/kunduk1/manage-task-service/internal/api/team/converter"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

// @Summary      List my teams
// @Description  Returns the teams the authenticated caller is a member of.
// @Tags         teams
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  teamv1.ListTeamsResponse
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /teams [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	teams, err := h.svc.List(r.Context(), uid)
	if err != nil {
		response.ServiceError(w, err)
		return
	}

	resp := teamv1.ListTeamsResponse{Teams: converter.ToTeamResponses(teams)}
	response.JSON(w, http.StatusOK, resp)
}
