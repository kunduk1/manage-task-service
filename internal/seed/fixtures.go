package seed

import "github.com/kunduk1/manage-task-service/internal/model"

// Этот файл — детерминированный набор фикстур для dev-БД. Сущности ссылаются друг на
// друга символьными ключами (key), а реальные ID подставляются при вставке (см. Run).
// Набор подобран так, чтобы покрыть все фичи сервиса: роли/authz, статусы задач,
// аналитику (TeamStats.DoneLast7Days, TopCreators, MisassignedTasks) и историю задач.

// SeedPassword — единый пароль всех засеянных пользователей (вход сразу после сидирования).
const SeedPassword = "password123"

type userFix struct {
	key   string
	email string
	name  string
}

type memberFix struct {
	userKey string
	role    model.TeamRole
}

type teamFix struct {
	key      string
	name     string
	desc     string
	ownerKey string
	members  []memberFix // без владельца — он добавляется автоматически с ролью owner
}

type taskFix struct {
	key          string
	teamKey      string
	title        string
	desc         string
	status       model.TaskStatus
	assigneeKey  string // "" — задача без исполнителя
	createdByKey string
	ageDays      int // created_at = now - ageDays
}

type historyFix struct {
	taskKey      string
	changedByKey string
	field        string
	oldVal       string
	newVal       string
	ageDays      int // changed_at = now - ageDays
}

type commentFix struct {
	taskKey string
	userKey string
	body    string
}

// fixtureUsers — 5 пользователей с общим паролем SeedPassword.
var fixtureUsers = []userFix{
	{key: "alice", email: "alice@example.com", name: "Alice Owner"},
	{key: "bob", email: "bob@example.com", name: "Bob Admin"},
	{key: "carol", email: "carol@example.com", name: "Carol Member"},
	{key: "dave", email: "dave@example.com", name: "Dave Lead"},
	{key: "erin", email: "erin@example.com", name: "Erin Member"},
}

// fixtureTeams — 2 команды с разными ролями (покрывают owner/admin/member и authz).
var fixtureTeams = []teamFix{
	{
		key: "platform", name: "Platform", desc: "Core platform team", ownerKey: "alice",
		members: []memberFix{
			{userKey: "bob", role: model.RoleAdmin},
			{userKey: "carol", role: model.RoleMember},
		},
	},
	{
		key: "mobile", name: "Mobile", desc: "Mobile apps team", ownerKey: "dave",
		members: []memberFix{
			{userKey: "erin", role: model.RoleMember},
		},
	},
}

// fixtureTasks — задачи с разбросом created_at (для TopCreators, окно 30 дней),
// всеми статусами и разными исполнителями. Задача "misassigned" назначена carol,
// которая НЕ состоит в Mobile, — её должен находить task.MisassignedTasks.
//
// Создатели задач (для TopCreators): Platform — alice ×3, bob ×2; Mobile — dave ×3.
var fixtureTasks = []taskFix{
	{key: "ci", teamKey: "platform", title: "Set up CI pipeline", desc: "GitHub Actions: lint+test+build", status: model.StatusDone, assigneeKey: "bob", createdByKey: "alice", ageDays: 20},
	{key: "docs", teamKey: "platform", title: "Write API docs", desc: "OpenAPI annotations for all handlers", status: model.StatusInProgress, assigneeKey: "carol", createdByKey: "alice", ageDays: 12},
	{key: "loginbug", teamKey: "platform", title: "Fix login bug", desc: "Invalid credentials returns 500", status: model.StatusTodo, assigneeKey: "alice", createdByKey: "bob", ageDays: 3},
	{key: "refactor", teamKey: "platform", title: "Refactor auth module", desc: "", status: model.StatusTodo, assigneeKey: "", createdByKey: "alice", ageDays: 1},
	{key: "migrate", teamKey: "platform", title: "Migrate DB to v2", desc: "Add query-performance indexes", status: model.StatusDone, assigneeKey: "bob", createdByKey: "bob", ageDays: 6},

	{key: "onboarding", teamKey: "mobile", title: "Design onboarding flow", desc: "3-screen first-run experience", status: model.StatusInProgress, assigneeKey: "erin", createdByKey: "dave", ageDays: 8},
	{key: "ship", teamKey: "mobile", title: "Ship v1.0 to stores", desc: "App Store + Google Play release", status: model.StatusDone, assigneeKey: "dave", createdByKey: "dave", ageDays: 5},
	{key: "misassigned", teamKey: "mobile", title: "Localize strings", desc: "Assignee is not a Mobile member (data-integrity demo)", status: model.StatusTodo, assigneeKey: "carol", createdByKey: "dave", ageDays: 4},
}

// fixtureHistory — записи аудита. Переходы в done в пределах 7 дней попадают в
// TeamStats.DoneLast7Days (migrate→Platform, ship→Mobile); ci сделан давно и не считается.
// Записи по docs обогащают GET /tasks/{id}/history разными полями.
var fixtureHistory = []historyFix{
	{taskKey: "ci", changedByKey: "bob", field: "status", oldVal: "in_progress", newVal: "done", ageDays: 19},
	{taskKey: "migrate", changedByKey: "bob", field: "status", oldVal: "in_progress", newVal: "done", ageDays: 3},
	{taskKey: "ship", changedByKey: "dave", field: "status", oldVal: "in_progress", newVal: "done", ageDays: 4},
	{taskKey: "docs", changedByKey: "alice", field: "title", oldVal: "Draft API docs", newVal: "Write API docs", ageDays: 11},
	{taskKey: "docs", changedByKey: "carol", field: "status", oldVal: "todo", newVal: "in_progress", ageDays: 10},
}

// fixtureComments — комментарии (репозиторий taskcomment уже есть; верхний слайс ещё не построен).
var fixtureComments = []commentFix{
	{taskKey: "docs", userKey: "carol", body: "Started drafting the endpoints section."},
	{taskKey: "docs", userKey: "alice", body: "Looks good — add auth examples."},
	{taskKey: "onboarding", userKey: "erin", body: "Need a design review before coding."},
}

// Account — публичные данные засеянного аккаунта (для вывода в CLI). Пароль — SeedPassword.
type Account struct {
	Email string
	Name  string
}

// Accounts возвращает засеянные аккаунты в порядке вставки (все с паролем SeedPassword).
func Accounts() []Account {
	out := make([]Account, len(fixtureUsers))
	for i, u := range fixtureUsers {
		out[i] = Account{Email: u.email, Name: u.name}
	}
	return out
}
