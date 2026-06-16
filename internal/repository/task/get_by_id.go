package task

import (
	"context"

	"github.com/georgysavva/scany/v2/sqlscan"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/task/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func (r *repo) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	q := db.Query{
		Name:     "task_repository.GetByID",
		QueryRaw: "SELECT id, team_id, title, description, status, assignee_id, created_by, created_at, updated_at FROM tasks WHERE id = ?",
	}

	var t repomodel.Task
	if err := r.db.DB().ScanOneContext(ctx, &t, q, id); err != nil {
		if sqlscan.NotFound(err) {
			return nil, errors.ErrTaskNotFound
		}
		return nil, err
	}
	return toServiceTask(&t), nil
}
