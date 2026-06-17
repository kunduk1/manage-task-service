//go:build integration

package integration_test

import (
	"fmt"
	"net/http"

	taskv1 "github.com/kunduk1/manage-task-service/pkg/task/v1"
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

// TestTaskLifecycle_EndToEnd — золотой сквозной путь: register → login → создать команду →
// пригласить участника → создать задачу → список (наполняет кэш) → обновить → история →
// список снова (отражает обновление = кэш был инвалидирован).
func (s *IntegrationSuite) TestTaskLifecycle_EndToEnd() {
	ownerID, ownerTok := s.registerAndLogin("owner@example.com", "Owner", "secret123")

	team := mustDo[teamv1.TeamResponse](s, http.MethodPost, "/api/v1/teams", ownerTok,
		teamv1.CreateTeamRequest{Name: "Platform", Description: "core"}, http.StatusCreated)
	s.Require().Equal(ownerID, team.CreatedBy)

	// второй пользователь — будущий исполнитель, должен быть участником
	assigneeID, _ := s.registerAndLogin("dev@example.com", "Dev", "secret123")
	invitePath := fmt.Sprintf("/api/v1/teams/%d/invite", team.ID)
	s.requireStatus(s.do(http.MethodPost, invitePath, ownerTok,
		teamv1.InviteRequest{UserID: assigneeID, Role: "member"}), http.StatusCreated)

	created := mustDo[taskv1.TaskResponse](s, http.MethodPost, "/api/v1/tasks", ownerTok,
		taskv1.CreateTaskRequest{TeamID: team.ID, Title: "Write docs", Description: "doc the API", Status: "todo", AssigneeID: &assigneeID},
		http.StatusCreated)
	s.Equal("todo", created.Status)
	s.Equal(ownerID, created.CreatedBy)

	listPath := fmt.Sprintf("/api/v1/tasks?team_id=%d", team.ID)

	// первый листинг наполняет кэш
	list := mustDo[taskv1.ListTasksResponse](s, http.MethodGet, listPath, ownerTok, nil, http.StatusOK)
	s.Require().Len(list.Tasks, 1)
	s.Equal(created.ID, list.Tasks[0].ID)
	s.True(s.taskCacheExists(team.ID), "ожидался заполненный кэш списка задач после первого листинга")

	// обновляем статус и заголовок
	newTitle := "Write docs v2"
	newStatus := "in_progress"
	updated := mustDo[taskv1.TaskResponse](s, http.MethodPut, fmt.Sprintf("/api/v1/tasks/%d", created.ID), ownerTok,
		taskv1.UpdateTaskRequest{Title: &newTitle, Status: &newStatus}, http.StatusOK)
	s.Equal("in_progress", updated.Status)
	s.Equal(newTitle, updated.Title)
	// обновление инвалидирует кэш команды
	s.False(s.taskCacheExists(team.ID), "ожидалась инвалидация кэша списка задач после обновления")

	// история должна содержать изменения status и title
	hist := mustDo[taskv1.TaskHistoryResponse](s, http.MethodGet, fmt.Sprintf("/api/v1/tasks/%d/history", created.ID), ownerTok, nil, http.StatusOK)
	s.True(hasHistory(hist.Entries, "status", "todo", "in_progress"), "нет записи истории по status")
	s.True(hasHistory(hist.Entries, "title", "Write docs", newTitle), "нет записи истории по title")

	// повторный листинг отражает обновление (кэш был сброшен)
	list2 := mustDo[taskv1.ListTasksResponse](s, http.MethodGet, listPath, ownerTok, nil, http.StatusOK)
	s.Require().Len(list2.Tasks, 1)
	s.Equal("in_progress", list2.Tasks[0].Status)
	s.Equal(newTitle, list2.Tasks[0].Title)
}

func (s *IntegrationSuite) TestTask_CreateByNonMemberForbidden() {
	_, ownerTok := s.registerAndLogin("owner@example.com", "Owner", "secret123")
	team := mustDo[teamv1.TeamResponse](s, http.MethodPost, "/api/v1/teams", ownerTok,
		teamv1.CreateTeamRequest{Name: "T"}, http.StatusCreated)

	_, strangerTok := s.registerAndLogin("stranger@example.com", "Stranger", "secret123")
	rsp := s.do(http.MethodPost, "/api/v1/tasks", strangerTok,
		taskv1.CreateTaskRequest{TeamID: team.ID, Title: "x"})
	s.requireError(rsp, http.StatusForbidden, "forbidden")
}

func (s *IntegrationSuite) TestTask_UpdateMissingNotFound() {
	_, tok := s.registerAndLogin("u@example.com", "U", "secret123")

	title := "x"
	rsp := s.do(http.MethodPut, "/api/v1/tasks/999999", tok, taskv1.UpdateTaskRequest{Title: &title})
	s.requireError(rsp, http.StatusNotFound, "task not found")
}

func (s *IntegrationSuite) TestTask_ListMissingTeamIDBadRequest() {
	_, tok := s.registerAndLogin("u@example.com", "U", "secret123")

	rsp := s.do(http.MethodGet, "/api/v1/tasks", tok, nil)
	s.requireStatus(rsp, http.StatusBadRequest)
}

// hasHistory ищет запись истории с заданными field/old/new.
func hasHistory(entries []taskv1.TaskHistoryEntryResponse, field, oldVal, newVal string) bool {
	for _, e := range entries {
		if e.Field == field &&
			e.OldValue != nil && *e.OldValue == oldVal &&
			e.NewValue != nil && *e.NewValue == newVal {
			return true
		}
	}
	return false
}
