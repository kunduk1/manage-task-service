package task

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/model"
)

// History возвращает историю изменений задачи (новые сверху). Доступно членам команды задачи.
func (s *serv) History(ctx context.Context, in model.TaskHistoryQuery) ([]model.TaskHistoryEntry, error) {
	task, err := s.taskRepo.GetByID(ctx, in.TaskID)
	if err != nil {
		return nil, err // ErrTaskNotFound → 404
	}
	if err := s.requireMembership(ctx, task.TeamID, in.ActorID); err != nil {
		return nil, err
	}
	return s.historyRepo.ListByTask(ctx, in.TaskID)
}
