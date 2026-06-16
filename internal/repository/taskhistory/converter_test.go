package taskhistory

import (
	"testing"

	repomodel "github.com/kunduk1/manage-task-service/internal/repository/taskhistory/model"
)

func TestToServiceEntry(t *testing.T) {
	if toServiceEntry(nil) != nil {
		t.Fatal("expected nil for nil input")
	}

	oldV, newV := "todo", "done"
	got := toServiceEntry(&repomodel.Entry{
		ID: 1, TaskID: 2, ChangedBy: 3, Field: "status", OldValue: &oldV, NewValue: &newV,
	})
	if got.OldValue == nil || *got.OldValue != oldV || got.NewValue == nil || *got.NewValue != newV {
		t.Errorf("nullable values not preserved: %+v", got)
	}

	// NULL значения остаются nil-указателями.
	gotNil := toServiceEntry(&repomodel.Entry{ID: 2, Field: "assignee_id"})
	if gotNil.OldValue != nil || gotNil.NewValue != nil {
		t.Errorf("nil values should stay nil: old=%v new=%v", gotNil.OldValue, gotNil.NewValue)
	}
}
