package team

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/team/model"
)

// ListByUser возвращает команды, в которых состоит пользователь
func (r *repo) ListByUser(ctx context.Context, userID int64) ([]model.Team, error) {
	q := db.Query{
		Name: "team_repository.ListByUser",
		QueryRaw: `SELECT t.id, t.name, t.description, t.created_by, t.created_at, t.updated_at
FROM teams t
JOIN team_members tm ON tm.team_id = t.id
WHERE tm.user_id = ?
ORDER BY t.id`,
	}

	var ts []repomodel.Team
	if err := r.db.DB().ScanAllContext(ctx, &ts, q, userID); err != nil {
		return nil, err
	}
	return toServiceTeams(ts), nil
}
