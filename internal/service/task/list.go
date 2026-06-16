package task

import (
	"context"
	"fmt"

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
	return s.taskRepo.List(ctx, filter)
}
