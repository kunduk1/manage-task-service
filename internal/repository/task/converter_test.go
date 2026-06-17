package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/task/model"
)

func TestToServiceTask(t *testing.T) {
	require.Nil(t, toServiceTask(nil))

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

	assert.Equal(t, desc, got.Description)
	assert.Equal(t, model.StatusInProgress, got.Status)
	require.NotNil(t, got.AssigneeID)
	assert.Equal(t, assignee, *got.AssigneeID)
}

func TestToServiceTask_NullableFields(t *testing.T) {
	got := toServiceTask(&repomodel.Task{ID: 1, Status: "todo"})

	assert.Empty(t, got.Description)
	assert.Nil(t, got.AssigneeID)
}
