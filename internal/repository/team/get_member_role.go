package team

import (
	"context"

	"github.com/georgysavva/scany/v2/sqlscan"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func (r *repo) GetMemberRole(ctx context.Context, teamID, userID int64) (model.TeamRole, error) {
	q := db.Query{
		Name:     "team_repository.GetMemberRole",
		QueryRaw: "SELECT role FROM team_members WHERE team_id = ? AND user_id = ?",
	}

	var role string
	if err := r.db.DB().ScanOneContext(ctx, &role, q, teamID, userID); err != nil {
		if sqlscan.NotFound(err) {
			return "", errors.ErrNotTeamMember
		}
		return "", err
	}
	return model.TeamRole(role), nil
}
