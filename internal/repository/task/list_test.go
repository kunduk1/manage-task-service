package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kunduk1/manage-task-service/internal/model"
)

func ptr[T any](v T) *T { return &v }

func TestBuildListQuery(t *testing.T) {
	const base = "SELECT id, team_id, title, description, status, assignee_id, created_by, created_at, updated_at FROM tasks"

	tests := []struct {
		name      string
		filter    model.TaskFilter
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name:      "no filters uses defaults",
			filter:    model.TaskFilter{},
			wantQuery: base + " ORDER BY id DESC LIMIT ? OFFSET ?",
			wantArgs:  []interface{}{defaultPageSize, 0},
		},
		{
			name: "all filters set, preserves order team→status→assignee",
			filter: model.TaskFilter{
				TeamID:     ptr(int64(1)),
				Status:     ptr(model.StatusTodo),
				AssigneeID: ptr(int64(5)),
				Limit:      10,
				Offset:     20,
			},
			wantQuery: base + " WHERE team_id = ? AND status = ? AND assignee_id = ? ORDER BY id DESC LIMIT ? OFFSET ?",
			wantArgs:  []interface{}{int64(1), "todo", int64(5), 10, 20},
		},
		{
			name:      "single filter",
			filter:    model.TaskFilter{Status: ptr(model.StatusDone)},
			wantQuery: base + " WHERE status = ? ORDER BY id DESC LIMIT ? OFFSET ?",
			wantArgs:  []interface{}{"done", defaultPageSize, 0},
		},
		{
			name:      "oversized limit clamped to default, negative offset to zero",
			filter:    model.TaskFilter{Limit: maxPageSize + 1, Offset: -5},
			wantQuery: base + " ORDER BY id DESC LIMIT ? OFFSET ?",
			wantArgs:  []interface{}{defaultPageSize, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotArgs := buildListQuery(tt.filter)

			assert.Equal(t, tt.wantQuery, gotQuery)
			require.Len(t, gotArgs, len(tt.wantArgs))
			for i := range tt.wantArgs {
				assert.Equal(t, tt.wantArgs[i], gotArgs[i])
			}
		})
	}
}
