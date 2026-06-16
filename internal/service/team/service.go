// Package team реализует бизнес-логику управления командами.
package team

import (
	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/repository"
	"github.com/kunduk1/manage-task-service/internal/service"
)

type serv struct {
	teamRepo  repository.TeamRepository
	userRepo  repository.UserRepository
	txManager db.TxManager
}

func NewService(
	teamRepo repository.TeamRepository,
	userRepo repository.UserRepository,
	txManager db.TxManager,
) service.TeamsService {
	return &serv{
		teamRepo:  teamRepo,
		userRepo:  userRepo,
		txManager: txManager,
	}
}
