package team

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/kunduk1/manage-task-service/internal/clients/email"
	"github.com/kunduk1/manage-task-service/internal/logger"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// Invite добавляет пользователя в команду. Приглашать может только owner или admin.
// Роль приглашаемого по умолчанию member; допускается также admin (но не owner).
func (s *serv) Invite(ctx context.Context, in model.InviteInput) error {
	// Проверка прав инициатора: он должен быть участником команды с ролью owner/admin.
	if err := s.authz.RequireRole(ctx, in.TeamID, in.ActorID, model.RoleOwner, model.RoleAdmin); err != nil {
		return err
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
	invitee, err := s.userRepo.GetByID(ctx, in.InviteeID)
	if err != nil {
		return err
	}

	if err := s.teamRepo.AddMember(ctx, &model.TeamMember{
		TeamID: in.TeamID,
		UserID: in.InviteeID,
		Role:   role,
	}); err != nil {
		return err
	}

	// Отправка письма — best-effort: сбой почты или разомкнутый брейкер
	// НЕ должны ронять приглашение. Логируем предупреждение и возвращаем nil.
	if s.emailClient != nil {
		if err := s.emailClient.SendInvite(ctx, email.Invite{
			ToEmail:   invitee.Email,
			ToName:    invitee.Name,
			TeamID:    in.TeamID,
			InviterID: in.ActorID,
			Role:      string(role),
		}); err != nil {
			logger.Warn("failed to send invitation email (best-effort)",
				zap.Int64("team_id", in.TeamID),
				zap.Int64("invitee_id", in.InviteeID),
				zap.Error(err),
			)
		}
	}

	return nil
}
