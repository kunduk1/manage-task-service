package repository

//go:generate go tool mockgen -source=repository.go -destination=mocks/repository_mock.go -package=mocks

import (
	"context"
	"time"

	"github.com/kunduk1/manage-task-service/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (int64, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type TokenRepository interface {
	SaveRefresh(ctx context.Context, token string, userID int64, ttl time.Duration) error
	GetRefresh(ctx context.Context, token string) (int64, error)
	DeleteRefresh(ctx context.Context, token string) error
}

type TeamRepository interface {
	Create(ctx context.Context, team *model.Team) (int64, error)
	GetByID(ctx context.Context, id int64) (*model.Team, error)
	ListByUser(ctx context.Context, userID int64) ([]model.Team, error)
	AddMember(ctx context.Context, m *model.TeamMember) error
	// TeamStats — JOIN 3+ таблиц + агрегация: по каждой команде название,
	// число участников и число задач в done за последние 7 дней.
	TeamStats(ctx context.Context) ([]model.TeamStats, error)
	// TopCreators — оконная функция: топ-3 пользователя по числу созданных задач
	// в каждой команде за период [from, to).
	TopCreators(ctx context.Context, from, to time.Time) ([]model.TopCreator, error)
}

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) (int64, error)
	GetByID(ctx context.Context, id int64) (*model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	List(ctx context.Context, f model.TaskFilter) ([]model.Task, error)
	// MisassignedTasks — запрос с условием по связанным таблицам (валидация целостности):
	// задачи, где assignee не является членом команды этой задачи.
	MisassignedTasks(ctx context.Context) ([]model.Task, error)
}

type TaskHistoryRepository interface {
	Create(ctx context.Context, e *model.TaskHistoryEntry) (int64, error)
	ListByTask(ctx context.Context, taskID int64) ([]model.TaskHistoryEntry, error)
}

type TaskCommentRepository interface {
	Create(ctx context.Context, c *model.TaskComment) (int64, error)
	ListByTask(ctx context.Context, taskID int64) ([]model.TaskComment, error)
	Delete(ctx context.Context, id int64) error
}
