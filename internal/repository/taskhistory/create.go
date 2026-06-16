package taskhistory

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
)

func (r *repo) Create(ctx context.Context, e *model.TaskHistoryEntry) (int64, error) {
	q := db.Query{
		Name:     "task_history_repository.Create",
		QueryRaw: "INSERT INTO task_history (task_id, changed_by, field, old_value, new_value) VALUES (?, ?, ?, ?, ?)",
	}

	res, err := r.db.DB().ExecContext(ctx, q, e.TaskID, e.ChangedBy, e.Field, e.OldValue, e.NewValue)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
