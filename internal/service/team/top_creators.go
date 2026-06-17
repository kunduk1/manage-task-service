package team

import (
	"context"
	"time"

	"github.com/kunduk1/manage-task-service/internal/model"
)

// TopCreators возвращает топ-3 создателя задач в каждой команде за период [from, to).
func (s *serv) TopCreators(ctx context.Context, from, to time.Time) ([]model.TopCreator, error) {
	return s.teamRepo.TopCreators(ctx, from, to)
}
