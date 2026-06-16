package model

import "time"

// TaskStatus — статус задачи (значения совпадают с ENUM в tasks.status).
type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

// Task — доменная сущность задачи. AssigneeID — указатель, т.к. задача может быть неназначенной.
type Task struct {
	ID          int64
	TeamID      int64
	Title       string
	Description string
	Status      TaskStatus
	AssigneeID  *int64
	CreatedBy   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TaskFilter — параметры фильтрации/пагинации для выборки задач.
type TaskFilter struct {
	TeamID     *int64
	Status     *TaskStatus
	AssigneeID *int64
	Limit      int
	Offset     int
}

// CreateTaskInput — входные данные для создания задачи.
// ActorID — аутентифицированный инициатор (становится CreatedBy и проверяется на членство).
type CreateTaskInput struct {
	ActorID     int64
	TeamID      int64
	Title       string
	Description string
	Status      TaskStatus
	AssigneeID  *int64
}

// UpdateTaskInput — входные данные для частичного обновления задачи; nil-поле = не менять.
// ActorID — аутентифицированный инициатор (для проверки прав). AssigneeID == 0 снимает исполнителя.
type UpdateTaskInput struct {
	ActorID     int64
	TaskID      int64
	Title       *string
	Description *string
	Status      *TaskStatus
	AssigneeID  *int64
}

// TaskListQuery — параметры выборки списка задач. TeamID обязателен (для проверки членства).
type TaskListQuery struct {
	ActorID    int64
	TeamID     int64
	Status     *TaskStatus
	AssigneeID *int64
	Limit      int
	Offset     int
}

// TaskHistoryQuery — параметры выборки истории задачи. ActorID — инициатор (для проверки членства).
type TaskHistoryQuery struct {
	ActorID int64
	TaskID  int64
}
