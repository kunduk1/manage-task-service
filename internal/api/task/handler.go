package task

import (
	"net/http"
	"strconv"

	"github.com/kunduk1/manage-task-service/internal/service"
	"github.com/kunduk1/manage-task-service/internal/transport/middleware"
)

// defaultPageSize/maxPageSize дублируют клампинг репозитория (internal/repository/task/list.go),
// чтобы поля limit/offset в ответе отражали фактический размер страницы.
const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type Handler struct {
	svc service.TasksService
}

func NewHandler(svc service.TasksService) *Handler {
	return &Handler{svc: svc}
}

// userID извлекает id аутентифицированного пользователя из контекста (subject JWT — строка)
// и парсит его в int64.
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

// parseIntDefault парсит строку запроса в int; пустая/невалидная — значение по умолчанию.
func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func clampLimit(v int) int {
	if v <= 0 || v > maxPageSize {
		return defaultPageSize
	}
	return v
}

func clampOffset(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
