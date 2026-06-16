package response

import (
	stderrors "errors"
	"net/http"

	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// ServiceError мапит доменные ошибки сервисного слоя в HTTP-ответ с нужным статусом.
func ServiceError(w http.ResponseWriter, err error) {
	switch {
	case stderrors.Is(err, errors.ErrValidation):
		Error(w, http.StatusBadRequest, err.Error())
	case stderrors.Is(err, errors.ErrUserExists):
		Error(w, http.StatusConflict, "user already exists")
	case stderrors.Is(err, errors.ErrInvalidCredentials):
		Error(w, http.StatusUnauthorized, "invalid email or password")
	case stderrors.Is(err, errors.ErrForbidden):
		Error(w, http.StatusForbidden, "forbidden")
	case stderrors.Is(err, errors.ErrTeamNotFound):
		Error(w, http.StatusNotFound, "team not found")
	case stderrors.Is(err, errors.ErrTaskNotFound):
		Error(w, http.StatusNotFound, "task not found")
	case stderrors.Is(err, errors.ErrUserNotFound):
		Error(w, http.StatusNotFound, "user not found")
	case stderrors.Is(err, errors.ErrMemberExists):
		Error(w, http.StatusConflict, "user is already a team member")
	default:
		Error(w, http.StatusInternalServerError, "internal server error")
	}
}
