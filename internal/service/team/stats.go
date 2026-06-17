package team

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/model"
)

// Stats возвращает аналитику по всем командам: число участников и число задач,
// переведённых в done за последние 7 дней.
func (s *serv) Stats(ctx context.Context) ([]model.TeamStats, error) {
	return s.teamRepo.TeamStats(ctx)
}
