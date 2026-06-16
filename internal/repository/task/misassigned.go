package task

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/task/model"
)

// MisassignedTasks — аналитический запрос: условие по связанным таблицам.
// Возвращает задачи, у которых назначенный исполнитель не состоит в команде задачи.
func (r *repo) MisassignedTasks(ctx context.Context) ([]model.Task, error) {
	q := db.Query{
		Name: "task_repository.MisassignedTasks",
		QueryRaw: `SELECT tk.id, tk.team_id, tk.title, tk.description, tk.status,
       tk.assignee_id, tk.created_by, tk.created_at, tk.updated_at
FROM tasks tk
WHERE tk.assignee_id IS NOT NULL
  AND NOT EXISTS (
        SELECT 1 FROM team_members tm
        WHERE tm.team_id = tk.team_id AND tm.user_id = tk.assignee_id
  )
ORDER BY tk.id`,
	}

	var ts []repomodel.Task
	if err := r.db.DB().ScanAllContext(ctx, &ts, q); err != nil {
		return nil, err
	}
	return toServiceTasks(ts), nil
}
