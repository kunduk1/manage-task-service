package taskcomment

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/taskcomment/model"
)

// ListByTask возвращает комментарии задачи в хронологическом порядке.
func (r *repo) ListByTask(ctx context.Context, taskID int64) ([]model.TaskComment, error) {
	q := db.Query{
		Name:     "task_comment_repository.ListByTask",
		QueryRaw: "SELECT id, task_id, user_id, body, created_at, updated_at FROM task_comments WHERE task_id = ? ORDER BY created_at ASC, id ASC",
	}

	var cs []repomodel.Comment
	if err := r.db.DB().ScanAllContext(ctx, &cs, q, taskID); err != nil {
		return nil, err
	}
	return toServiceComments(cs), nil
}
