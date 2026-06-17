package taskhistory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repomodel "github.com/kunduk1/manage-task-service/internal/repository/taskhistory/model"
)

func TestToServiceEntry(t *testing.T) {
	require.Nil(t, toServiceEntry(nil))

	oldV, newV := "todo", "done"
	got := toServiceEntry(&repomodel.Entry{
		ID: 1, TaskID: 2, ChangedBy: 3, Field: "status", OldValue: &oldV, NewValue: &newV,
	})
	require.NotNil(t, got.OldValue)
	assert.Equal(t, oldV, *got.OldValue)
	require.NotNil(t, got.NewValue)
	assert.Equal(t, newV, *got.NewValue)

	// NULL значения остаются nil-указателями.
	gotNil := toServiceEntry(&repomodel.Entry{ID: 2, Field: "assignee_id"})
	assert.Nil(t, gotNil.OldValue)
	assert.Nil(t, gotNil.NewValue)
}
