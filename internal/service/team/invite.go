package team

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// Invite добавляет пользователя в команду. Приглашать может только owner или admin.
// Роль приглашаемого по умолчанию member; допускается также admin (но не owner).
func (s *serv) Invite(ctx context.Context, in model.InviteInput) error {
	// Проверка прав инициатора: он должен быть участником команды с ролью owner/admin.
	actorRole, err := s.teamRepo.GetMemberRole(ctx, in.TeamID, in.ActorID)
	if err != nil {
		if stderrors.Is(err, errors.ErrNotTeamMember) {
			return errors.ErrForbidden
		}
		return err
	}
	if actorRole != model.RoleOwner && actorRole != model.RoleAdmin {
		return errors.ErrForbidden
	}

	// Роль приглашаемого: по умолчанию member; owner назначить нельзя.
	role := in.Role
	if role == "" {
		role = model.RoleMember
	}
	if role != model.RoleMember && role != model.RoleAdmin {
		return fmt.Errorf("%w: role must be member or admin", errors.ErrValidation)
	}

	if in.InviteeID <= 0 {
		return fmt.Errorf("%w: user_id is required", errors.ErrValidation)
	}

	// Приглашаемый должен существовать (FK на users) — отдаём 404, а не ошибку БД.
	if _, err := s.userRepo.GetByID(ctx, in.InviteeID); err != nil {
		return err
	}

	return s.teamRepo.AddMember(ctx, &model.TeamMember{
		TeamID: in.TeamID,
		UserID: in.InviteeID,
		Role:   role,
	})
}
