package team

import (
	"context"
	"fmt"
	"strings"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// Create создаёт команду и атомарно добавляет создателя как владельца (owner).
func (s *serv) Create(ctx context.Context, in model.CreateTeamInput) (*model.Team, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: team name is required", errors.ErrValidation)
	}

	team := &model.Team{
		Name:        name,
		Description: strings.TrimSpace(in.Description),
		CreatedBy:   in.OwnerID,
	}

	// Создание команды и добавление владельца в team_members — в одной транзакции:
	// если вторая вставка упадёт, команда без владельца не останется.
	err := s.txManager.ReadCommit(ctx, func(txCtx context.Context) error {
		id, err := s.teamRepo.Create(txCtx, team)
		if err != nil {
			return err
		}
		team.ID = id

		return s.teamRepo.AddMember(txCtx, &model.TeamMember{
			TeamID: id,
			UserID: in.OwnerID,
			Role:   model.RoleOwner,
		})
	})
	if err != nil {
		return nil, err
	}

	// Перечитываем команду, чтобы вернуть проставленные БД таймстемпы.
	return s.teamRepo.GetByID(ctx, team.ID)
}
