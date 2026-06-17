package task

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
)

func TestMisassigned_PassThrough(t *testing.T) {
	svc, taskRepo, _, _, _, _ := newTestService(t)
	want := []model.Task{
		{ID: 1, TeamID: 1, Title: "orphan", AssigneeID: ptrInt64(99)},
	}

	taskRepo.EXPECT().MisassignedTasks(gomock.Any()).Return(want, nil)

	got, err := svc.Misassigned(context.Background())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
