package task

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/kunduk1/manage-task-service/internal/logger"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// Update частично обновляет задачу и пишет историю изменений в одной транзакции.
// Право на обновление: создатель задачи, её текущий исполнитель либо owner/admin команды.
func (s *serv) Update(ctx context.Context, in model.UpdateTaskInput) (*model.Task, error) {
	cur, err := s.taskRepo.GetByID(ctx, in.TaskID)
	if err != nil {
		return nil, err // ErrTaskNotFound пробрасываем → 404
	}

	if err := s.authorizeUpdate(ctx, cur, in.ActorID); err != nil {
		return nil, err
	}

	// Применяем переданные (ненулевые) поля к копии текущей задачи.
	updated := *cur

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return nil, fmt.Errorf("%w: title must not be empty", errors.ErrValidation)
		}
		updated.Title = title
	}
	if in.Description != nil {
		updated.Description = strings.TrimSpace(*in.Description)
	}
	if in.Status != nil {
		if !validStatus(*in.Status) {
			return nil, fmt.Errorf("%w: invalid status", errors.ErrValidation)
		}
		updated.Status = *in.Status
	}
	if in.AssigneeID != nil {
		if *in.AssigneeID <= 0 {
			// 0 (или отрицательное) — снять исполнителя.
			updated.AssigneeID = nil
		} else {
			if err := s.requireAssigneeMembership(ctx, cur.TeamID, *in.AssigneeID); err != nil {
				return nil, err
			}
			assignee := *in.AssigneeID
			updated.AssigneeID = &assignee
		}
	}

	entries := diffHistory(cur, &updated, in.ActorID)

	// Обновление задачи и запись истории изменений — атомарно.
	err = s.txManager.ReadCommit(ctx, func(txCtx context.Context) error {
		if err := s.taskRepo.Update(txCtx, &updated); err != nil {
			return err
		}
		for i := range entries {
			if _, err := s.historyRepo.Create(txCtx, &entries[i]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Задача в списке команды изменилась — сбрасываем кэш
	if err := s.cacheRepo.InvalidateTeam(ctx, cur.TeamID); err != nil {
		logger.Warn("task list cache invalidation failed", zap.Int64("team_id", cur.TeamID), zap.Error(err))
	}

	// Перечитываем задачу, чтобы вернуть обновлённый БД таймстемп updated_at.
	return s.taskRepo.GetByID(ctx, in.TaskID)
}

// authorizeUpdate разрешает обновление создателю задачи, её текущему исполнителю
// либо owner/admin команды; остальным — ErrForbidden.
func (s *serv) authorizeUpdate(ctx context.Context, t *model.Task, actorID int64) error {
	role, err := s.authz.RequireMember(ctx, t.TeamID, actorID)
	if err != nil {
		return err
	}
	if role == model.RoleOwner || role == model.RoleAdmin {
		return nil
	}
	if actorID == t.CreatedBy {
		return nil
	}
	if t.AssigneeID != nil && *t.AssigneeID == actorID {
		return nil
	}
	return errors.ErrForbidden
}
