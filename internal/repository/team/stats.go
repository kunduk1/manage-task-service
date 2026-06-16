package team

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/team/model"
)

// TeamStats — аналитический запрос
// По каждой команде: название, число участников и число задач, переведённых
// в статус done за последние 7 дней (по аудиту task_history).
func (r *repo) TeamStats(ctx context.Context) ([]model.TeamStats, error) {
	q := db.Query{
		Name: "team_repository.TeamStats",
		QueryRaw: `SELECT t.id   AS team_id,
       t.name AS name,
       COUNT(DISTINCT tm.user_id) AS member_count,
       COUNT(DISTINCT CASE WHEN th.id IS NOT NULL THEN tk.id END) AS done_last_7_days
FROM teams t
LEFT JOIN team_members tm ON tm.team_id = t.id
LEFT JOIN tasks tk        ON tk.team_id = t.id
LEFT JOIN task_history th ON th.task_id = tk.id
                         AND th.field = 'status'
                         AND th.new_value = 'done'
                         AND th.changed_at >= (NOW() - INTERVAL 7 DAY)
GROUP BY t.id, t.name
ORDER BY t.id`,
	}

	var rows []repomodel.TeamStats
	if err := r.db.DB().ScanAllContext(ctx, &rows, q); err != nil {
		return nil, err
	}
	return toServiceStats(rows), nil
}
