package team

import (
	"context"

	"github.com/georgysavva/scany/v2/sqlscan"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/team/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func (r *repo) GetByID(ctx context.Context, id int64) (*model.Team, error) {
	q := db.Query{
		Name:     "team_repository.GetByID",
		QueryRaw: "SELECT id, name, description, created_by, created_at, updated_at FROM teams WHERE id = ?",
	}

	var t repomodel.Team
	if err := r.db.DB().ScanOneContext(ctx, &t, q, id); err != nil {
		if sqlscan.NotFound(err) {
			return nil, errors.ErrTeamNotFound
		}
		return nil, err
	}
	return toServiceTeam(&t), nil
}
