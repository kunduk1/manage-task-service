package team

import (
	"net/http"
	"strconv"

	"github.com/kunduk1/manage-task-service/internal/service"
	"github.com/kunduk1/manage-task-service/internal/transport/middleware"
)

type Handler struct {
	svc service.TeamsService
}

func NewHandler(svc service.TeamsService) *Handler {
	return &Handler{svc: svc}
}

// userID извлекает id аутентифицированного пользователя из контекста (subject JWT — строка)
// и парсит его в int64
func userID(r *http.Request) (int64, bool) {
	sub, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}
