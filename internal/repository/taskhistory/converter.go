package taskhistory

import (
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/taskhistory/model"
)

func toServiceEntry(e *repomodel.Entry) *model.TaskHistoryEntry {
	if e == nil {
		return nil
	}
	return &model.TaskHistoryEntry{
		ID:        e.ID,
		TaskID:    e.TaskID,
		ChangedBy: e.ChangedBy,
		Field:     e.Field,
		OldValue:  e.OldValue,
		NewValue:  e.NewValue,
		ChangedAt: e.ChangedAt,
	}
}

func toServiceEntries(es []repomodel.Entry) []model.TaskHistoryEntry {
	out := make([]model.TaskHistoryEntry, 0, len(es))
	for i := range es {
		out = append(out, *toServiceEntry(&es[i]))
	}
	return out
}
