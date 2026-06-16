package model

import "time"

// Entry — строка таблицы task_history (аудит изменения поля задачи).
type Entry struct {
	ID        int64     `db:"id"`
	TaskID    int64     `db:"task_id"`
	ChangedBy int64     `db:"changed_by"`
	Field     string    `db:"field"`
	OldValue  *string   `db:"old_value"` // NULL-able колонка
	NewValue  *string   `db:"new_value"` // NULL-able колонка
	ChangedAt time.Time `db:"changed_at"`
}
