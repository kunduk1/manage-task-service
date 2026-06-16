package auth

import "github.com/kunduk1/manage-task-service/internal/service"

type Handler struct {
	svc service.AuthService
}

func NewHandler(svc service.AuthService) *Handler {
	return &Handler{svc: svc}
}
