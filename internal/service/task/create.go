package task

import (
	"context"
	"fmt"
	"strings"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// Create создаёт задачу. Создавать может только член команды; если указан исполнитель,
// он тоже должен быть членом этой команды.
func (s *serv) Create(ctx context.Context, in model.CreateTaskInput) (*model.Task, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", errors.ErrValidation)
	}
	if in.TeamID <= 0 {
		return nil, fmt.Errorf("%w: team_id is required", errors.ErrValidation)
	}

	status := in.Status
	if status == "" {
		status = model.StatusTodo
	}
	if !validStatus(status) {
		return nil, fmt.Errorf("%w: invalid status", errors.ErrValidation)
	}

	// Неположительный assignee_id трактуем как «без исполнителя».
	assignee := in.AssigneeID
	if assignee != nil && *assignee <= 0 {
		assignee = nil
	}

	// Автор должен быть членом команды.
	if err := s.requireMembership(ctx, in.TeamID, in.ActorID); err != nil {
		return nil, err
	}

	// Если задан исполнитель — он тоже должен быть членом команды.
	if assignee != nil {
		if err := s.requireAssigneeMembership(ctx, in.TeamID, *assignee); err != nil {
			return nil, err
		}
	}

	task := &model.Task{
		TeamID:      in.TeamID,
		Title:       title,
		Description: strings.TrimSpace(in.Description),
		Status:      status,
		AssigneeID:  assignee,
		CreatedBy:   in.ActorID,
	}

	id, err := s.taskRepo.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	// Перечитываем задачу, чтобы вернуть проставленные БД таймстемпы.
	return s.taskRepo.GetByID(ctx, id)
}
