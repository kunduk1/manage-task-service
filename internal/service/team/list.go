package team

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/model"
)

// List возвращает команды, в которых состоит пользователь.
func (s *serv) List(ctx context.Context, userID int64) ([]model.Team, error) {
	return s.teamRepo.ListByUser(ctx, userID)
}
