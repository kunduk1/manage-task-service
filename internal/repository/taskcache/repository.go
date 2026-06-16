package taskcache

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kunduk1/manage-task-service/internal/clients/cache"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/internal/repository"
)

// defaultTTL — время жизни кэша списка задач команды.
const defaultTTL = 5 * time.Minute

const keyPrefix = "tasks:team:"

type repo struct {
	client cache.Client
}

func NewRepository(client cache.Client) repository.TaskCacheRepository {
	return &repo{client: client}
}

// key — ключ Redis-хэша команды; все варианты фильтров живут полями внутри него,
// поэтому инвалидация команды — это удаление одного ключа.
func key(teamID int64) string {
	return keyPrefix + strconv.FormatInt(teamID, 10)
}

// field — детерминированное имя поля хэша под конкретный набор фильтров (без team_id,
// он уже в ключе). nil-фильтры кодируются как "-", чтобы поля не схлопывались.
func field(f model.TaskFilter) string {
	status := "-"
	if f.Status != nil {
		status = string(*f.Status)
	}
	assignee := "-"
	if f.AssigneeID != nil {
		assignee = strconv.FormatInt(*f.AssigneeID, 10)
	}
	return fmt.Sprintf("s=%s|a=%s|l=%d|o=%d", status, assignee, f.Limit, f.Offset)
}
