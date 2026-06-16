package auth

import (
	"net/http"

	"github.com/kunduk1/manage-task-service/internal/api/auth/converter"
	"github.com/kunduk1/manage-task-service/internal/transport/request"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
	authv1 "github.com/kunduk1/manage-task-service/pkg/auth/v1"
)

// @Summary      Register a new user
// @Description  Creates a user account and returns its public profile.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      authv1.RegisterRequest  true  "registration payload"
// @Success      201      {object}  authv1.RegisterResponse
// @Failure      400      {object}  response.ErrorBody
// @Failure      409      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req authv1.RegisterRequest
	if !request.DecodeJSON(w, r, &req) {
		return
	}

	user, err := h.svc.Register(r.Context(), converter.ToRegisterInput(req))
	if err != nil {
		response.ServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, converter.ToRegisterResponse(user))
}
