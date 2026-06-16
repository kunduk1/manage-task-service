package team

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/kunduk1/manage-task-service/internal/api/team/converter"
	"github.com/kunduk1/manage-task-service/internal/transport/request"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

// @Summary      Invite a user to a team
// @Description  Adds a user to the team. Only owner or admin may invite. Role defaults to member.
// @Tags         teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                   true  "team id"
// @Param        request  body      teamv1.InviteRequest  true  "invite payload"
// @Success      201      {object}  teamv1.InviteResponse
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      403      {object}  response.ErrorBody
// @Failure      404      {object}  response.ErrorBody
// @Failure      409      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /teams/{id}/invite [post]
func (h *Handler) Invite(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	teamID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid team id")
		return
	}

	var req teamv1.InviteRequest
	if !request.DecodeJSON(w, r, &req) {
		return
	}

	in := converter.ToInviteInput(teamID, uid, req)
	if err := h.svc.Invite(r.Context(), in); err != nil {
		response.ServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, converter.ToInviteResponse(in))
}
