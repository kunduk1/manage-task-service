package task

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/model"
)

// Misassigned возвращает задачи, где assignee не является членом команды задачи
func (s *serv) Misassigned(ctx context.Context) ([]model.Task, error) {
	return s.taskRepo.MisassignedTasks(ctx)
}
