package team

import (
	"net/http"

	"github.com/kunduk1/manage-task-service/internal/api/team/converter"
	"github.com/kunduk1/manage-task-service/internal/transport/request"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

// @Summary      Create a team
// @Description  Creates a team; the authenticated caller becomes its owner.
// @Tags         teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      teamv1.CreateTeamRequest  true  "team payload"
// @Success      201      {object}  teamv1.TeamResponse
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /teams [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req teamv1.CreateTeamRequest
	if !request.DecodeJSON(w, r, &req) {
		return
	}

	team, err := h.svc.Create(r.Context(), converter.ToCreateInput(req, uid))
	if err != nil {
		response.ServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, converter.ToTeamResponse(team))
}
