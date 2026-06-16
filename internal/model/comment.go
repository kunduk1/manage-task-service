package model

import "time"

type TaskComment struct {
	ID        int64
	TaskID    int64
	UserID    int64
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
