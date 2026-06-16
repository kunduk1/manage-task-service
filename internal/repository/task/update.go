package task

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
)

// Update перезаписывает изменяемые поля задачи
func (r *repo) Update(ctx context.Context, task *model.Task) error {
	q := db.Query{
		Name:     "task_repository.Update",
		QueryRaw: "UPDATE tasks SET title = ?, description = ?, status = ?, assignee_id = ? WHERE id = ?",
	}

	_, err := r.db.DB().ExecContext(ctx, q,
		task.Title, task.Description, string(task.Status), task.AssigneeID, task.ID,
	)
	return err
}
