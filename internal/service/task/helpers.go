package task

import (
	"context"
	stderrors "errors"
	"fmt"
	"strconv"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// validStatus проверяет, что статус — одно из допустимых значений ENUM tasks.status.
func validStatus(s model.TaskStatus) bool {
	switch s {
	case model.StatusTodo, model.StatusInProgress, model.StatusDone:
		return true
	default:
		return false
	}
}

// requireMembership проверяет, что пользователь — член команды; иначе ErrForbidden.
func (s *serv) requireMembership(ctx context.Context, teamID, userID int64) error {
	_, err := s.authz.RequireMember(ctx, teamID, userID)
	return err
}

// requireAssigneeMembership проверяет, что назначаемый исполнитель — член команды задачи;
// иначе ErrValidation (нельзя назначить задачу не-участнику — защита от misassigned на записи).
// Это валидация входных данных, поэтому отсутствие членства маппится в ErrValidation, а не ErrForbidden.
func (s *serv) requireAssigneeMembership(ctx context.Context, teamID, assigneeID int64) error {
	if _, err := s.authz.MemberRole(ctx, teamID, assigneeID); err != nil {
		if stderrors.Is(err, errors.ErrNotTeamMember) {
			return fmt.Errorf("%w: assignee must be a team member", errors.ErrValidation)
		}
		return err
	}
	return nil
}

// diffHistory строит записи истории по изменившимся полям задачи.
func diffHistory(old, updated *model.Task, actorID int64) []model.TaskHistoryEntry {
	var entries []model.TaskHistoryEntry

	add := func(field string, oldVal, newVal *string) {
		entries = append(entries, model.TaskHistoryEntry{
			TaskID:    old.ID,
			ChangedBy: actorID,
			Field:     field,
			OldValue:  oldVal,
			NewValue:  newVal,
		})
	}

	if old.Title != updated.Title {
		add("title", strPtr(old.Title), strPtr(updated.Title))
	}
	if old.Description != updated.Description {
		add("description", strPtr(old.Description), strPtr(updated.Description))
	}
	if old.Status != updated.Status {
		add("status", strPtr(string(old.Status)), strPtr(string(updated.Status)))
	}
	if !sameAssignee(old.AssigneeID, updated.AssigneeID) {
		add("assignee_id", idPtr(old.AssigneeID), idPtr(updated.AssigneeID))
	}

	return entries
}

func strPtr(s string) *string {
	return &s
}

// idPtr форматирует assignee_id в строку; nil (нет исполнителя) сохраняется как nil.
func idPtr(id *int64) *string {
	if id == nil {
		return nil
	}
	s := strconv.FormatInt(*id, 10)
	return &s
}

// sameAssignee сравнивает двух исполнителей с учётом nil (неназначенная задача).
func sameAssignee(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
