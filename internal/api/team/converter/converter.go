package converter

import (
	"github.com/kunduk1/manage-task-service/internal/model"
	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

func ToCreateInput(req teamv1.CreateTeamRequest, ownerID int64) model.CreateTeamInput {
	return model.CreateTeamInput{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
	}
}

func ToTeamResponse(t *model.Team) teamv1.TeamResponse {
	return teamv1.TeamResponse{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		CreatedBy:   t.CreatedBy,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func ToTeamResponses(teams []model.Team) []teamv1.TeamResponse {
	out := make([]teamv1.TeamResponse, 0, len(teams))
	for i := range teams {
		out = append(out, ToTeamResponse(&teams[i]))
	}
	return out
}

// ToInviteInput собирает доменный ввод из тела запроса, id команды (из пути) и инициатора (из токена).
// Пустая роль нормализуется в member, чтобы ответ отражал фактически назначенную роль.
func ToInviteInput(teamID, actorID int64, req teamv1.InviteRequest) model.InviteInput {
	role := model.TeamRole(req.Role)
	if role == "" {
		role = model.RoleMember
	}
	return model.InviteInput{
		TeamID:    teamID,
		ActorID:   actorID,
		InviteeID: req.UserID,
		Role:      role,
	}
}

func ToInviteResponse(in model.InviteInput) teamv1.InviteResponse {
	return teamv1.InviteResponse{
		TeamID: in.TeamID,
		UserID: in.InviteeID,
		Role:   string(in.Role),
	}
}

func ToTeamStatsResponses(stats []model.TeamStats) []teamv1.TeamStatItem {
	out := make([]teamv1.TeamStatItem, 0, len(stats))
	for i := range stats {
		out = append(out, teamv1.TeamStatItem{
			TeamID:        stats[i].TeamID,
			Name:          stats[i].Name,
			MemberCount:   stats[i].MemberCount,
			DoneLast7Days: stats[i].DoneLast7Days,
		})
	}
	return out
}

func ToTopCreatorResponses(creators []model.TopCreator) []teamv1.TopCreatorItem {
	out := make([]teamv1.TopCreatorItem, 0, len(creators))
	for i := range creators {
		out = append(out, teamv1.TopCreatorItem{
			TeamID:       creators[i].TeamID,
			UserID:       creators[i].UserID,
			UserName:     creators[i].UserName,
			CreatedCount: creators[i].CreatedCount,
			Rank:         creators[i].Rank,
		})
	}
	return out
}
