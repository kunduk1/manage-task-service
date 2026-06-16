package team

import (
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/team/model"
)

func toServiceTeam(t *repomodel.Team) *model.Team {
	if t == nil {
		return nil
	}
	desc := ""
	if t.Description != nil {
		desc = *t.Description
	}
	return &model.Team{
		ID:          t.ID,
		Name:        t.Name,
		Description: desc,
		CreatedBy:   t.CreatedBy,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func toServiceTeams(ts []repomodel.Team) []model.Team {
	out := make([]model.Team, 0, len(ts))
	for i := range ts {
		out = append(out, *toServiceTeam(&ts[i]))
	}
	return out
}

func toServiceStats(rows []repomodel.TeamStats) []model.TeamStats {
	out := make([]model.TeamStats, 0, len(rows))
	for _, s := range rows {
		out = append(out, model.TeamStats{
			TeamID:        s.TeamID,
			Name:          s.Name,
			MemberCount:   s.MemberCount,
			DoneLast7Days: s.DoneLast7Days,
		})
	}
	return out
}

func toServiceTopCreators(rows []repomodel.TopCreator) []model.TopCreator {
	out := make([]model.TopCreator, 0, len(rows))
	for _, c := range rows {
		out = append(out, model.TopCreator{
			TeamID:       c.TeamID,
			UserID:       c.UserID,
			UserName:     c.UserName,
			CreatedCount: c.CreatedCount,
			Rank:         c.Rank,
		})
	}
	return out
}
