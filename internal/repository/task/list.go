package task

import (
	"context"
	"strings"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/task/model"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// List возвращает задачи с опциональной фильтрацией и пагинацией.
// WHERE собирается только из заданных (не-nil) фильтров
func (r *repo) List(ctx context.Context, f model.TaskFilter) ([]model.Task, error) {
	query, args := buildListQuery(f)

	q := db.Query{Name: "task_repository.List", QueryRaw: query}

	var ts []repomodel.Task
	if err := r.db.DB().ScanAllContext(ctx, &ts, q, args...); err != nil {
		return nil, err
	}
	return toServiceTasks(ts), nil
}

// buildListQuery собирает SQL и упорядоченный список аргументов из фильтра.
func buildListQuery(f model.TaskFilter) (string, []interface{}) {
	var (
		conds []string
		args  []interface{}
	)

	if f.TeamID != nil {
		conds = append(conds, "team_id = ?")
		args = append(args, *f.TeamID)
	}
	if f.Status != nil {
		conds = append(conds, "status = ?")
		args = append(args, string(*f.Status))
	}
	if f.AssigneeID != nil {
		conds = append(conds, "assignee_id = ?")
		args = append(args, *f.AssigneeID)
	}

	query := "SELECT id, team_id, title, description, status, assignee_id, created_by, created_at, updated_at FROM tasks"
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY id DESC"

	limit := f.Limit
	if limit <= 0 || limit > maxPageSize {
		limit = defaultPageSize
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	return query, args
}
