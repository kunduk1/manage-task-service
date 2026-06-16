package team

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
)

func (r *repo) Create(ctx context.Context, team *model.Team) (int64, error) {
	q := db.Query{
		Name:     "team_repository.Create",
		QueryRaw: "INSERT INTO teams (name, description, created_by) VALUES (?, ?, ?)",
	}

	res, err := r.db.DB().ExecContext(ctx, q, team.Name, team.Description, team.CreatedBy)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
