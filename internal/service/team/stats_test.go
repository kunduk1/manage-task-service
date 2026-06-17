package team

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
)

func TestStats_PassThrough(t *testing.T) {
	svc, teamRepo, _, _, _ := newTestService(t)
	want := []model.TeamStats{
		{TeamID: 1, Name: "A", MemberCount: 3, DoneLast7Days: 2},
		{TeamID: 2, Name: "B", MemberCount: 1, DoneLast7Days: 0},
	}

	teamRepo.EXPECT().TeamStats(gomock.Any()).Return(want, nil)

	got, err := svc.Stats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
