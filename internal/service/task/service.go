// Package task реализует бизнес-логику управления задачами.
package task

import (
	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/repository"
	"github.com/kunduk1/manage-task-service/internal/service"
	"github.com/kunduk1/manage-task-service/internal/service/authz"
)

type serv struct {
	taskRepo    repository.TaskRepository
	historyRepo repository.TaskHistoryRepository
	cacheRepo   repository.TaskCacheRepository
	txManager   db.TxManager
	authz       *authz.Authorizer
}

func NewService(
	taskRepo repository.TaskRepository,
	historyRepo repository.TaskHistoryRepository,
	cacheRepo repository.TaskCacheRepository,
	txManager db.TxManager,
	authorizer *authz.Authorizer,
) service.TasksService {
	return &serv{
		taskRepo:    taskRepo,
		historyRepo: historyRepo,
		cacheRepo:   cacheRepo,
		txManager:   txManager,
		authz:       authorizer,
	}
}
