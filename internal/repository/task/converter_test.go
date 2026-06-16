package task

import (
	"testing"

	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/task/model"
)

func TestToServiceTask(t *testing.T) {
	if toServiceTask(nil) != nil {
		t.Fatal("expected nil for nil input")
	}

	desc := "do the thing"
	assignee := int64(7)
	got := toServiceTask(&repomodel.Task{
		ID:          1,
		TeamID:      2,
		Title:       "title",
		Description: &desc,
		Status:      "in_progress",
		AssigneeID:  &assignee,
		CreatedBy:   3,
	})

	if got.Description != desc {
		t.Errorf("description: got %q, want %q", got.Description, desc)
	}
	if got.Status != model.StatusInProgress {
		t.Errorf("status: got %q, want %q", got.Status, model.StatusInProgress)
	}
	if got.AssigneeID == nil || *got.AssigneeID != assignee {
		t.Errorf("assignee: got %v, want %d", got.AssigneeID, assignee)
	}
}

func TestToServiceTask_NullableFields(t *testing.T) {
	got := toServiceTask(&repomodel.Task{ID: 1, Status: "todo"})

	if got.Description != "" {
		t.Errorf("nil description should map to empty string, got %q", got.Description)
	}
	if got.AssigneeID != nil {
		t.Errorf("nil assignee should stay nil, got %v", got.AssigneeID)
	}
}
