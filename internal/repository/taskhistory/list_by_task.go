package taskhistory

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/taskhistory/model"
)

// ListByTask возвращает историю изменений задачи в порядке от новых к старым.
func (r *repo) ListByTask(ctx context.Context, taskID int64) ([]model.TaskHistoryEntry, error) {
	q := db.Query{
		Name:     "task_history_repository.ListByTask",
		QueryRaw: "SELECT id, task_id, changed_by, field, old_value, new_value, changed_at FROM task_history WHERE task_id = ? ORDER BY changed_at DESC, id DESC",
	}

	var es []repomodel.Entry
	if err := r.db.DB().ScanAllContext(ctx, &es, q, taskID); err != nil {
		return nil, err
	}
	return toServiceEntries(es), nil
}
