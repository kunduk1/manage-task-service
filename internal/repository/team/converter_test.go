package team

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repomodel "github.com/kunduk1/manage-task-service/internal/repository/team/model"
)

func TestToServiceTeam(t *testing.T) {
	require.Nil(t, toServiceTeam(nil))

	desc := "the team"
	got := toServiceTeam(&repomodel.Team{ID: 1, Name: "alpha", Description: &desc, CreatedBy: 9})
	assert.Equal(t, "alpha", got.Name)
	assert.Equal(t, desc, got.Description)
	assert.Equal(t, int64(9), got.CreatedBy)

	gotNil := toServiceTeam(&repomodel.Team{ID: 2, Name: "beta"})
	assert.Empty(t, gotNil.Description)
}

func TestToServiceStatsAndTopCreators(t *testing.T) {
	stats := toServiceStats([]repomodel.TeamStats{{TeamID: 1, Name: "a", MemberCount: 3, DoneLast7Days: 2}})
	require.Len(t, stats, 1)
	assert.Equal(t, int64(3), stats[0].MemberCount)
	assert.Equal(t, int64(2), stats[0].DoneLast7Days)

	top := toServiceTopCreators([]repomodel.TopCreator{{TeamID: 1, UserID: 5, UserName: "u", CreatedCount: 4, Rank: 1}})
	require.Len(t, top, 1)
	assert.Equal(t, "u", top[0].UserName)
	assert.Equal(t, int64(1), top[0].Rank)
}
