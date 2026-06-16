package authz

import (
	"context"
	stderrors "errors"
	"slices"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// teamRoleProvider — узкий интерфейс доступа к ролям участников команды.
type teamRoleProvider interface {
	GetMemberRole(ctx context.Context, teamID, userID int64) (model.TeamRole, error)
}

type Authorizer struct {
	teams teamRoleProvider
}

func New(teams teamRoleProvider) *Authorizer {
	return &Authorizer{teams: teams}
}

// MemberRole возвращает роль пользователя в команде
func (a *Authorizer) MemberRole(ctx context.Context, teamID, userID int64) (model.TeamRole, error) {
	return a.teams.GetMemberRole(ctx, teamID, userID)
}

// RequireMember требует, чтобы пользователь был участником команды, и возвращает его
// роль
func (a *Authorizer) RequireMember(ctx context.Context, teamID, userID int64) (model.TeamRole, error) {
	role, err := a.MemberRole(ctx, teamID, userID)
	if err != nil {
		if stderrors.Is(err, errors.ErrNotTeamMember) {
			return "", errors.ErrForbidden
		}
		return "", err
	}
	return role, nil
}

// RequireRole требует членства в команде с одной из допустимых ролей
func (a *Authorizer) RequireRole(ctx context.Context, teamID, userID int64, allowed ...model.TeamRole) error {
	role, err := a.RequireMember(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if slices.Contains(allowed, role) {
		return nil
	}
	return errors.ErrForbidden
}
