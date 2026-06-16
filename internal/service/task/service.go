// Package task реализует бизнес-логику управления задачами.
package task

import (
	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/repository"
	"github.com/kunduk1/manage-task-service/internal/service"
)

type serv struct {
	taskRepo    repository.TaskRepository
	historyRepo repository.TaskHistoryRepository
	teamRepo    repository.TeamRepository
	txManager   db.TxManager
}

func NewService(
	taskRepo repository.TaskRepository,
	historyRepo repository.TaskHistoryRepository,
	teamRepo repository.TeamRepository,
	txManager db.TxManager,
) service.TasksService {
	return &serv{
		taskRepo:    taskRepo,
		historyRepo: historyRepo,
		teamRepo:    teamRepo,
		txManager:   txManager,
	}
}
