package team

import (
	"context"
	"time"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/team/model"
)

// TopCreators — аналитический запрос
// Топ-3 пользователя по числу созданных задач в каждой команде за период [from, to).
func (r *repo) TopCreators(ctx context.Context, from, to time.Time) ([]model.TopCreator, error) {
	q := db.Query{
		Name: "team_repository.TopCreators",
		QueryRaw: `SELECT r.team_id, r.user_id, u.name AS user_name, r.created_count, r.rnk
FROM (
    SELECT tk.team_id,
           tk.created_by AS user_id,
           COUNT(*)      AS created_count,
           ROW_NUMBER() OVER (
               PARTITION BY tk.team_id
               ORDER BY COUNT(*) DESC, tk.created_by ASC
           ) AS rnk
    FROM tasks tk
    WHERE tk.created_at >= ? AND tk.created_at < ?
    GROUP BY tk.team_id, tk.created_by
) r
JOIN users u ON u.id = r.user_id
WHERE r.rnk <= 3
ORDER BY r.team_id, r.rnk`,
	}

	var rows []repomodel.TopCreator
	if err := r.db.DB().ScanAllContext(ctx, &rows, q, from, to); err != nil {
		return nil, err
	}
	return toServiceTopCreators(rows), nil
}
