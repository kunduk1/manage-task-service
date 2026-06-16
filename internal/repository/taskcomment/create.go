package taskcomment

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
)

func (r *repo) Create(ctx context.Context, c *model.TaskComment) (int64, error) {
	q := db.Query{
		Name:     "task_comment_repository.Create",
		QueryRaw: "INSERT INTO task_comments (task_id, user_id, body) VALUES (?, ?, ?)",
	}

	res, err := r.db.DB().ExecContext(ctx, q, c.TaskID, c.UserID, c.Body)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
