package taskcache

import (
	"context"
	"encoding/json"

	"github.com/kunduk1/manage-task-service/internal/model"
)

// SetTaskList сохраняет список задач под поле фильтра f в хэше команды и продлевает TTL
// всего ключа. TTL общий на команду — каждое заполнение кэша сдвигает срок жизни хэша.
func (r *repo) SetTaskList(ctx context.Context, teamID int64, f model.TaskFilter, tasks []model.Task) error {
	raw, err := json.Marshal(tasks)
	if err != nil {
		return err
	}

	k := key(teamID)
	if err := r.client.HSet(ctx, k, map[string]any{field(f): string(raw)}); err != nil {
		return err
	}
	return r.client.Expire(ctx, k, defaultTTL)
}
