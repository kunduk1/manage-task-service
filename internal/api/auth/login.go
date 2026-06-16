package auth

import (
	"net/http"

	"github.com/kunduk1/manage-task-service/internal/api/auth/converter"
	"github.com/kunduk1/manage-task-service/internal/transport/request"
	"github.com/kunduk1/manage-task-service/internal/transport/response"
	authv1 "github.com/kunduk1/manage-task-service/pkg/auth/v1"
)

// @Summary      Log in and obtain tokens
// @Description  Validates credentials and returns access/refresh tokens.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      authv1.LoginRequest  true  "login payload"
// @Success      200      {object}  authv1.LoginResponse
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req authv1.LoginRequest
	if !request.DecodeJSON(w, r, &req) {
		return
	}

	tokens, err := h.svc.Login(r.Context(), converter.ToLoginInput(req))
	if err != nil {
		response.ServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, converter.ToLoginResponse(tokens))
}
