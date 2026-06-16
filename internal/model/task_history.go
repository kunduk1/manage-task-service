package model

import "time"

// TaskHistoryEntry — запись аудита изменения задачи
type TaskHistoryEntry struct {
	ID        int64
	TaskID    int64
	ChangedBy int64
	Field     string
	OldValue  *string
	NewValue  *string
	ChangedAt time.Time
}
