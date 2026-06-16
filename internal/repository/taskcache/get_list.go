package taskcache

import (
	"context"
	"encoding/json"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// GetTaskList ищет в хэше команды поле, соответствующее фильтру f. HGetAll отдаёт
// плоский срез [field, value, ...]; отсутствие поля (в т.ч. пустой/несуществующий ключ)
// трактуется как промах кэша.
func (r *repo) GetTaskList(ctx context.Context, teamID int64, f model.TaskFilter) ([]model.Task, error) {
	pairs, err := r.client.HGetAll(ctx, key(teamID))
	if err != nil {
		return nil, err
	}

	want := field(f)
	for i := 0; i+1 < len(pairs); i += 2 {
		name, ok := pairs[i].(string)
		if !ok || name != want {
			continue
		}
		raw, ok := pairs[i+1].(string)
		if !ok {
			return nil, errors.ErrCacheMiss
		}
		var tasks []model.Task
		if err := json.Unmarshal([]byte(raw), &tasks); err != nil {
			return nil, err
		}
		return tasks, nil
	}

	return nil, errors.ErrCacheMiss
}
