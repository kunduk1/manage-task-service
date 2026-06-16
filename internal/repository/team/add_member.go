package team

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func (r *repo) AddMember(ctx context.Context, m *model.TeamMember) error {
	role := m.Role
	if role == "" {
		role = model.RoleMember
	}

	q := db.Query{
		Name:     "team_repository.AddMember",
		QueryRaw: "INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)",
	}

	if _, err := r.db.DB().ExecContext(ctx, q, m.TeamID, m.UserID, string(role)); err != nil {
		if db.IsDuplicateEntry(err) {
			return errors.ErrMemberExists
		}
		return err
	}
	return nil
}
