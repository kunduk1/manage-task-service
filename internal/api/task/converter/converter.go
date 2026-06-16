package converter

import (
	"github.com/kunduk1/manage-task-service/internal/model"
	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
)

func ToCreateInput(req taskv1.CreateTaskRequest, actorID int64) model.CreateTaskInput {
	return model.CreateTaskInput{
		ActorID:     actorID,
		TeamID:      req.TeamID,
		Title:       req.Title,
		Description: req.Description,
		Status:      model.TaskStatus(req.Status),
		AssigneeID:  req.AssigneeID,
	}
}

func ToUpdateInput(req taskv1.UpdateTaskRequest, taskID, actorID int64) model.UpdateTaskInput {
	in := model.UpdateTaskInput{
		ActorID:     actorID,
		TaskID:      taskID,
		Title:       req.Title,
		Description: req.Description,
		AssigneeID:  req.AssigneeID,
	}
	if req.Status != nil {
		st := model.TaskStatus(*req.Status)
		in.Status = &st
	}
	return in
}

func ToTaskResponse(t *model.Task) taskv1.TaskResponse {
	return taskv1.TaskResponse{
		ID:          t.ID,
		TeamID:      t.TeamID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		AssigneeID:  t.AssigneeID,
		CreatedBy:   t.CreatedBy,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func ToTaskResponses(tasks []model.Task) []taskv1.TaskResponse {
	out := make([]taskv1.TaskResponse, 0, len(tasks))
	for i := range tasks {
		out = append(out, ToTaskResponse(&tasks[i]))
	}
	return out
}

func ToHistoryResponses(entries []model.TaskHistoryEntry) []taskv1.TaskHistoryEntryResponse {
	out := make([]taskv1.TaskHistoryEntryResponse, 0, len(entries))
	for i := range entries {
		e := entries[i]
		out = append(out, taskv1.TaskHistoryEntryResponse{
			ID:        e.ID,
			TaskID:    e.TaskID,
			ChangedBy: e.ChangedBy,
			Field:     e.Field,
			OldValue:  e.OldValue,
			NewValue:  e.NewValue,
			ChangedAt: e.ChangedAt,
		})
	}
	return out
}
