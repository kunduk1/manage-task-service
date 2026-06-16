package task

import (
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/task/model"
)

func toServiceTask(t *repomodel.Task) *model.Task {
	if t == nil {
		return nil
	}
	desc := ""
	if t.Description != nil {
		desc = *t.Description
	}
	return &model.Task{
		ID:          t.ID,
		TeamID:      t.TeamID,
		Title:       t.Title,
		Description: desc,
		Status:      model.TaskStatus(t.Status),
		AssigneeID:  t.AssigneeID,
		CreatedBy:   t.CreatedBy,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func toServiceTasks(ts []repomodel.Task) []model.Task {
	out := make([]model.Task, 0, len(ts))
	for i := range ts {
		out = append(out, *toServiceTask(&ts[i]))
	}
	return out
}
