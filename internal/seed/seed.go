// Package seed наполняет dev-БД детерминированным набором данных (см. fixtures.go).
// Используется бинарём cmd/seed. Простые сущности (users/teams/members/comments)
// вставляются через репозитории; задачи и история — сырым SQL, т.к. им нужен явный
// created_at/changed_at (репозиторный Create полагается на DEFAULT CURRENT_TIMESTAMP),
// а аналитике нужен разброс по времени. Паттерн зеркалит seed-хелперы интеграционных
// тестов (test/integration/suite_test.go).
package seed

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/internal/repository"
	taskcommentrepo "github.com/kunduk1/manage-task-service/internal/repository/taskcomment"
	teamrepo "github.com/kunduk1/manage-task-service/internal/repository/team"
	userrepo "github.com/kunduk1/manage-task-service/internal/repository/user"
	apperrors "github.com/kunduk1/manage-task-service/pkg/errors"
)

// Seeder наполняет БД фикстурами. now — база для относительных таймстемпов
// (передаётся снаружи ради воспроизводимости).
type Seeder struct {
	db          db.Client
	userRepo    repository.UserRepository
	teamRepo    repository.TeamRepository
	commentRepo repository.TaskCommentRepository
	now         time.Time
}

// New собирает Seeder поверх готового db-клиента (репозитории создаёт сам).
func New(dbc db.Client, now time.Time) *Seeder {
	return &Seeder{
		db:          dbc,
		userRepo:    userrepo.NewRepository(dbc),
		teamRepo:    teamrepo.NewRepository(dbc),
		commentRepo: taskcommentrepo.NewRepository(dbc),
		now:         now,
	}
}

// Summary — счётчики вставленных строк (для финального отчёта CLI).
type Summary struct {
	Users    int
	Teams    int
	Members  int
	Tasks    int
	History  int
	Comments int
}

// AlreadySeeded сообщает, есть ли уже засеянные данные (по первому фикстур-пользователю).
func (s *Seeder) AlreadySeeded(ctx context.Context) (bool, error) {
	_, err := s.userRepo.GetByEmail(ctx, fixtureUsers[0].email)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, apperrors.ErrUserNotFound):
		return false, nil
	default:
		return false, err
	}
}

// resetStatements — полные строковые литералы (без конкатенации с именами таблиц),
// чтобы не триггерить gosec G201. Порядок не важен: FK-проверки выключены на время очистки.
// TRUNCATE сбрасывает AUTO_INCREMENT → детерминированные ID между прогонами.
var resetStatements = []string{
	"SET FOREIGN_KEY_CHECKS = 0",
	"TRUNCATE TABLE task_history",
	"TRUNCATE TABLE task_comments",
	"TRUNCATE TABLE tasks",
	"TRUNCATE TABLE team_members",
	"TRUNCATE TABLE teams",
	"TRUNCATE TABLE users",
	"SET FOREIGN_KEY_CHECKS = 1",
}

// Reset очищает все доменные таблицы. Полагается на единственное соединение пула
// (cmd/seed ставит MaxOpenConns=1): FOREIGN_KEY_CHECKS сессионная, и без этого
// TRUNCATE таблиц с внешними ключами отвалится.
func (s *Seeder) Reset(ctx context.Context) error {
	for _, raw := range resetStatements {
		if _, err := s.db.DB().ExecContext(ctx, db.Query{Name: "seed.reset", QueryRaw: raw}); err != nil {
			return fmt.Errorf("reset %q: %w", raw, err)
		}
	}
	return nil
}

// Run вставляет весь граф фикстур в порядке зависимостей и возвращает счётчики.
// Реальные ID берутся из вставок и связываются по символьным ключам фикстур.
func (s *Seeder) Run(ctx context.Context) (Summary, error) {
	var sum Summary

	hash, err := bcrypt.GenerateFromPassword([]byte(SeedPassword), bcrypt.DefaultCost)
	if err != nil {
		return sum, fmt.Errorf("hash seed password: %w", err)
	}

	userIDs := make(map[string]int64, len(fixtureUsers))
	for _, u := range fixtureUsers {
		id, err := s.userRepo.Create(ctx, &model.User{Email: u.email, Name: u.name, PasswordHash: string(hash)})
		if err != nil {
			return sum, fmt.Errorf("create user %s: %w", u.email, err)
		}
		userIDs[u.key] = id
		sum.Users++
	}

	teamIDs := make(map[string]int64, len(fixtureTeams))
	for _, t := range fixtureTeams {
		ownerID := userIDs[t.ownerKey]
		teamID, err := s.teamRepo.Create(ctx, &model.Team{Name: t.name, Description: t.desc, CreatedBy: ownerID})
		if err != nil {
			return sum, fmt.Errorf("create team %s: %w", t.name, err)
		}
		teamIDs[t.key] = teamID
		sum.Teams++

		// Владелец — участник с ролью owner (как в service/team/create.go).
		if err := s.teamRepo.AddMember(ctx, &model.TeamMember{TeamID: teamID, UserID: ownerID, Role: model.RoleOwner}); err != nil {
			return sum, fmt.Errorf("add owner to team %s: %w", t.name, err)
		}
		sum.Members++

		for _, m := range t.members {
			if err := s.teamRepo.AddMember(ctx, &model.TeamMember{TeamID: teamID, UserID: userIDs[m.userKey], Role: m.role}); err != nil {
				return sum, fmt.Errorf("add member %s to team %s: %w", m.userKey, t.name, err)
			}
			sum.Members++
		}
	}

	taskIDs := make(map[string]int64, len(fixtureTasks))
	for _, t := range fixtureTasks {
		var assignee *int64
		if t.assigneeKey != "" {
			a := userIDs[t.assigneeKey]
			assignee = &a
		}
		taskID, err := s.insertTask(ctx, teamIDs[t.teamKey], t.title, t.desc, string(t.status), assignee, userIDs[t.createdByKey], s.daysAgo(t.ageDays))
		if err != nil {
			return sum, fmt.Errorf("create task %q: %w", t.title, err)
		}
		taskIDs[t.key] = taskID
		sum.Tasks++
	}

	for _, h := range fixtureHistory {
		if err := s.insertHistory(ctx, taskIDs[h.taskKey], userIDs[h.changedByKey], h.field, h.oldVal, h.newVal, s.daysAgo(h.ageDays)); err != nil {
			return sum, fmt.Errorf("create history for task %s: %w", h.taskKey, err)
		}
		sum.History++
	}

	for _, c := range fixtureComments {
		if _, err := s.commentRepo.Create(ctx, &model.TaskComment{TaskID: taskIDs[c.taskKey], UserID: userIDs[c.userKey], Body: c.body}); err != nil {
			return sum, fmt.Errorf("create comment on task %s: %w", c.taskKey, err)
		}
		sum.Comments++
	}

	return sum, nil
}

// insertTask вставляет задачу с явным created_at (репозиторный Create таймстемп не принимает).
func (s *Seeder) insertTask(ctx context.Context, teamID int64, title, desc, status string, assignee *int64, createdBy int64, createdAt time.Time) (int64, error) {
	q := db.Query{
		Name:     "seed.insertTask",
		QueryRaw: "INSERT INTO tasks (team_id, title, description, status, assignee_id, created_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
	}
	res, err := s.db.DB().ExecContext(ctx, q, teamID, title, desc, status, assignee, createdBy, createdAt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// insertHistory вставляет запись аудита с явным changed_at (нужно для окна 7 дней в TeamStats).
func (s *Seeder) insertHistory(ctx context.Context, taskID, changedBy int64, field, oldVal, newVal string, changedAt time.Time) error {
	q := db.Query{
		Name:     "seed.insertHistory",
		QueryRaw: "INSERT INTO task_history (task_id, changed_by, field, old_value, new_value, changed_at) VALUES (?, ?, ?, ?, ?, ?)",
	}
	_, err := s.db.DB().ExecContext(ctx, q, taskID, changedBy, field, oldVal, newVal, changedAt)
	return err
}

// daysAgo возвращает момент now - d суток.
func (s *Seeder) daysAgo(d int) time.Time {
	return s.now.AddDate(0, 0, -d)
}
