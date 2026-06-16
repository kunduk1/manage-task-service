package task

import (
	"context"
	stderrors "errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/kunduk1/manage-task-service/internal/logger"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// List возвращает задачи команды с фильтрацией и пагинацией. Требует team_id и членство
// вызывающего в команде — иначе выборка без фильтра по команде утекала бы задачи чужих команд.
func (s *serv) List(ctx context.Context, in model.TaskListQuery) ([]model.Task, error) {
	if in.TeamID <= 0 {
		return nil, fmt.Errorf("%w: team_id is required", errors.ErrValidation)
	}
	if in.Status != nil && !validStatus(*in.Status) {
		return nil, fmt.Errorf("%w: invalid status", errors.ErrValidation)
	}
	if err := s.requireMembership(ctx, in.TeamID, in.ActorID); err != nil {
		return nil, err
	}

	teamID := in.TeamID
	filter := model.TaskFilter{
		TeamID:     &teamID,
		Status:     in.Status,
		AssigneeID: in.AssigneeID,
		Limit:      in.Limit,
		Offset:     in.Offset,
	}

	// Read-through: пробуем кэш; промах и любые ошибки кэша — не фатальны, идём в БД.
	cached, err := s.cacheRepo.GetTaskList(ctx, in.TeamID, filter)
	if err == nil {
		return cached, nil
	}
	if !stderrors.Is(err, errors.ErrCacheMiss) {
		logger.Warn("task list cache read failed", zap.Int64("team_id", in.TeamID), zap.Error(err))
	}

	tasks, err := s.taskRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err := s.cacheRepo.SetTaskList(ctx, in.TeamID, filter, tasks); err != nil {
		logger.Warn("task list cache write failed", zap.Int64("team_id", in.TeamID), zap.Error(err))
	}
	return tasks, nil
}
