// Package v1 — публичный контракт транспортного уровня для домена tasks (HTTP/JSON).
package v1

import "time"

type CreateTaskRequest struct {
	TeamID      int64  `json:"team_id"               example:"1"`
	Title       string `json:"title"                 example:"Write docs"`
	Description string `json:"description,omitempty" example:"Document the tasks API"`
	Status      string `json:"status,omitempty"      example:"todo"`
	AssigneeID  *int64 `json:"assignee_id,omitempty" example:"5"`
}

// UpdateTaskRequest — частичное обновление
type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"       example:"New title"`
	Description *string `json:"description,omitempty" example:"Updated description"`
	Status      *string `json:"status,omitempty"      example:"in_progress"`
	AssigneeID  *int64  `json:"assignee_id,omitempty" example:"5"`
}

type TaskResponse struct {
	ID          int64     `json:"id"                    example:"1"`
	TeamID      int64     `json:"team_id"               example:"1"`
	Title       string    `json:"title"                 example:"Write docs"`
	Description string    `json:"description"           example:"Document the tasks API"`
	Status      string    `json:"status"                example:"todo"`
	AssigneeID  *int64    `json:"assignee_id,omitempty" example:"5"`
	CreatedBy   int64     `json:"created_by"            example:"42"`
	CreatedAt   time.Time `json:"created_at"            example:"2026-06-16T12:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at"            example:"2026-06-16T12:00:00Z"`
}

type ListTasksResponse struct {
	Tasks  []TaskResponse `json:"tasks"`
	Limit  int            `json:"limit"  example:"20"`
	Offset int            `json:"offset" example:"0"`
}

type TaskHistoryEntryResponse struct {
	ID        int64     `json:"id"         example:"1"`
	TaskID    int64     `json:"task_id"    example:"1"`
	ChangedBy int64     `json:"changed_by" example:"42"`
	Field     string    `json:"field"      example:"status"`
	OldValue  *string   `json:"old_value"  example:"todo"`
	NewValue  *string   `json:"new_value"  example:"in_progress"`
	ChangedAt time.Time `json:"changed_at" example:"2026-06-16T12:00:00Z"`
}

type TaskHistoryResponse struct {
	Entries []TaskHistoryEntryResponse `json:"entries"`
}

type MisassignedTasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}
