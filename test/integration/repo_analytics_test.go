//go:build integration

package integration_test

import (
	"time"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
)

// Эти тесты бьют по нетривиальному SQL напрямую через репозитории — у этих запросов
// нет HTTP-ручек, а в обычном покрытии слой repository/ исключён.

func (s *IntegrationSuite) TestTeamStats_CountsMembersAndRecentDone() {
	owner := s.seedUser("owner@example.com", "Owner")
	member := s.seedUser("member@example.com", "Member")
	teamID := s.seedTeam(owner, "Platform") // владелец уже участник
	s.seedMember(teamID, member, model.RoleMember)

	// две задачи; перевод в done: один недавно (в окне 7 дней), один давно (вне окна)
	t1 := s.seedTaskAt(teamID, owner, "done", time.Now())
	t2 := s.seedTaskAt(teamID, owner, "done", time.Now())
	s.seedHistoryAt(t1, owner, "status", "in_progress", "done", time.Now().Add(-2*24*time.Hour))
	s.seedHistoryAt(t2, owner, "status", "in_progress", "done", time.Now().Add(-30*24*time.Hour))

	stats, err := s.teamRepo.TeamStats(s.ctx)
	s.Require().NoError(err)

	st := findTeamStats(stats, teamID)
	s.Require().NotNil(st, "команда отсутствует в статистике")
	s.Equal(int64(2), st.MemberCount)
	s.Equal(int64(1), st.DoneLast7Days, "должен учитываться только недавний перевод в done")
}

func (s *IntegrationSuite) TestTopCreators_RanksWithinWindow() {
	u1 := s.seedUser("u1@example.com", "U1")
	u2 := s.seedUser("u2@example.com", "U2")
	teamID := s.seedTeam(u1, "T")
	s.seedMember(teamID, u2, model.RoleMember)

	now := time.Now()
	// u1: 3 задачи в окне; u2: 1 в окне + 1 вне окна (не должна считаться)
	for i := 0; i < 3; i++ {
		s.seedTaskAt(teamID, u1, "todo", now.Add(-time.Hour))
	}
	s.seedTaskAt(teamID, u2, "todo", now.Add(-time.Hour))
	s.seedTaskAt(teamID, u2, "todo", now.Add(-10*24*time.Hour))

	from := now.Add(-7 * 24 * time.Hour)
	to := now.Add(time.Hour)
	top, err := s.teamRepo.TopCreators(s.ctx, from, to)
	s.Require().NoError(err)

	rows := filterTopByTeam(top, teamID)
	s.Require().Len(rows, 2)
	s.Equal(u1, rows[0].UserID)
	s.Equal(int64(3), rows[0].CreatedCount)
	s.Equal(int64(1), rows[0].Rank)
	s.Equal(u2, rows[1].UserID)
	s.Equal(int64(1), rows[1].CreatedCount)
	s.Equal(int64(2), rows[1].Rank)
}

func (s *IntegrationSuite) TestMisassignedTasks_DetectsNonMemberAssignee() {
	owner := s.seedUser("owner@example.com", "Owner")
	member := s.seedUser("member@example.com", "Member")
	stranger := s.seedUser("stranger@example.com", "Stranger") // существует, но не в команде
	teamID := s.seedTeam(owner, "T")
	s.seedMember(teamID, member, model.RoleMember)

	okTask := s.seedTaskAssigned(teamID, owner, member)
	badTask := s.seedTaskAssigned(teamID, owner, stranger)

	res, err := s.taskRepo.MisassignedTasks(s.ctx)
	s.Require().NoError(err)

	ids := taskIDs(res)
	s.Contains(ids, badTask, "ожидалась мисассайн-задача в результате")
	s.NotContains(ids, okTask, "корректно назначенная задача не должна попадать в результат")
}

// --- сид-хелперы с явными таймстемпами (repo.Create ставит created_at/changed_at сам) ---

func (s *IntegrationSuite) seedTaskAt(teamID, createdBy int64, status string, createdAt time.Time) int64 {
	s.T().Helper()
	res, err := s.db.DB().ExecContext(s.ctx, db.Query{
		Name:     "integration.seedTask",
		QueryRaw: "INSERT INTO tasks (team_id, title, description, status, assignee_id, created_by, created_at) VALUES (?, 'task', '', ?, NULL, ?, ?)",
	}, teamID, status, createdBy, createdAt)
	s.Require().NoError(err, "seed task")
	id, err := res.LastInsertId()
	s.Require().NoError(err)
	return id
}

func (s *IntegrationSuite) seedTaskAssigned(teamID, createdBy, assignee int64) int64 {
	s.T().Helper()
	res, err := s.db.DB().ExecContext(s.ctx, db.Query{
		Name:     "integration.seedTaskAssigned",
		QueryRaw: "INSERT INTO tasks (team_id, title, description, status, assignee_id, created_by) VALUES (?, 'task', '', 'todo', ?, ?)",
	}, teamID, assignee, createdBy)
	s.Require().NoError(err, "seed assigned task")
	id, err := res.LastInsertId()
	s.Require().NoError(err)
	return id
}

func (s *IntegrationSuite) seedHistoryAt(taskID, changedBy int64, field, oldVal, newVal string, changedAt time.Time) {
	s.execSQL(
		"INSERT INTO task_history (task_id, changed_by, field, old_value, new_value, changed_at) VALUES (?, ?, ?, ?, ?, ?)",
		taskID, changedBy, field, oldVal, newVal, changedAt)
}

func findTeamStats(stats []model.TeamStats, teamID int64) *model.TeamStats {
	for i := range stats {
		if stats[i].TeamID == teamID {
			return &stats[i]
		}
	}
	return nil
}

func filterTopByTeam(top []model.TopCreator, teamID int64) []model.TopCreator {
	var out []model.TopCreator
	for _, c := range top {
		if c.TeamID == teamID {
			out = append(out, c)
		}
	}
	return out
}

func taskIDs(tasks []model.Task) []int64 {
	ids := make([]int64, 0, len(tasks))
	for _, t := range tasks {
		ids = append(ids, t.ID)
	}
	return ids
}
