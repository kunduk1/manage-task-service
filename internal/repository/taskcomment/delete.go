package taskcomment

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
)

// Delete удаляет комментарий по id
func (r *repo) Delete(ctx context.Context, id int64) error {
	q := db.Query{
		Name:     "task_comment_repository.Delete",
		QueryRaw: "DELETE FROM task_comments WHERE id = ?",
	}

	_, err := r.db.DB().ExecContext(ctx, q, id)
	return err
}
