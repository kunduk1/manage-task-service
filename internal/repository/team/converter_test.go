package team

import (
	"testing"

	repomodel "github.com/kunduk1/manage-task-service/internal/repository/team/model"
)

func TestToServiceTeam(t *testing.T) {
	if toServiceTeam(nil) != nil {
		t.Fatal("expected nil for nil input")
	}

	desc := "the team"
	got := toServiceTeam(&repomodel.Team{ID: 1, Name: "alpha", Description: &desc, CreatedBy: 9})
	if got.Name != "alpha" || got.Description != desc || got.CreatedBy != 9 {
		t.Errorf("unexpected mapping: %+v", got)
	}

	gotNil := toServiceTeam(&repomodel.Team{ID: 2, Name: "beta"})
	if gotNil.Description != "" {
		t.Errorf("nil description should map to empty string, got %q", gotNil.Description)
	}
}

func TestToServiceStatsAndTopCreators(t *testing.T) {
	stats := toServiceStats([]repomodel.TeamStats{{TeamID: 1, Name: "a", MemberCount: 3, DoneLast7Days: 2}})
	if len(stats) != 1 || stats[0].MemberCount != 3 || stats[0].DoneLast7Days != 2 {
		t.Errorf("unexpected stats mapping: %+v", stats)
	}

	top := toServiceTopCreators([]repomodel.TopCreator{{TeamID: 1, UserID: 5, UserName: "u", CreatedCount: 4, Rank: 1}})
	if len(top) != 1 || top[0].UserName != "u" || top[0].Rank != 1 {
		t.Errorf("unexpected topCreators mapping: %+v", top)
	}
}
