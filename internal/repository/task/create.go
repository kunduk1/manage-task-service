package task

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
)

func (r *repo) Create(ctx context.Context, task *model.Task) (int64, error) {
	status := task.Status
	if status == "" {
		status = model.StatusTodo
	}

	q := db.Query{
		Name:     "task_repository.Create",
		QueryRaw: "INSERT INTO tasks (team_id, title, description, status, assignee_id, created_by) VALUES (?, ?, ?, ?, ?, ?)",
	}

	res, err := r.db.DB().ExecContext(ctx, q,
		task.TeamID, task.Title, task.Description, string(status), task.AssigneeID, task.CreatedBy,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
