package model

import "time"

type Task struct {
	ID          int64     `db:"id"`
	TeamID      int64     `db:"team_id"`
	Title       string    `db:"title"`
	Description *string   `db:"description"`
	Status      string    `db:"status"`
	AssigneeID  *int64    `db:"assignee_id"`
	CreatedBy   int64     `db:"created_by"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
