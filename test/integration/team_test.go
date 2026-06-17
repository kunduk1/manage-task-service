//go:build integration

package integration_test

import (
	"fmt"
	"net/http"

	teamv1 "github.com/kunduk1/manage-task-service/pkg/team/v1"
)

func (s *IntegrationSuite) TestTeams_CreateAndList() {
	ownerID, access := s.registerAndLogin("owner@example.com", "Owner", "secret123")

	created := mustDo[teamv1.TeamResponse](s, http.MethodPost, "/api/v1/teams", access,
		teamv1.CreateTeamRequest{Name: "Platform", Description: "core"}, http.StatusCreated)
	s.NotZero(created.ID)
	s.Equal("Platform", created.Name)
	s.Equal(ownerID, created.CreatedBy)

	list := mustDo[teamv1.ListTeamsResponse](s, http.MethodGet, "/api/v1/teams", access, nil, http.StatusOK)
	s.Require().Len(list.Teams, 1)
	s.Equal(created.ID, list.Teams[0].ID)
}

func (s *IntegrationSuite) TestTeams_ListUnauthorized() {
	rsp := s.do(http.MethodGet, "/api/v1/teams", "", nil)
	s.requireStatus(rsp, http.StatusUnauthorized)
}

func (s *IntegrationSuite) TestInvite_DuplicateMember() {
	_, ownerTok := s.registerAndLogin("owner@example.com", "Owner", "secret123")
	team := mustDo[teamv1.TeamResponse](s, http.MethodPost, "/api/v1/teams", ownerTok,
		teamv1.CreateTeamRequest{Name: "T"}, http.StatusCreated)

	invitee := s.register("inv@example.com", "Inv", "secret123")
	path := fmt.Sprintf("/api/v1/teams/%d/invite", team.ID)

	s.requireStatus(s.do(http.MethodPost, path, ownerTok,
		teamv1.InviteRequest{UserID: invitee.ID, Role: "member"}), http.StatusCreated)

	rsp := s.do(http.MethodPost, path, ownerTok, teamv1.InviteRequest{UserID: invitee.ID, Role: "member"})
	s.requireError(rsp, http.StatusConflict, "user is already a team member")
}

func (s *IntegrationSuite) TestInvite_ByMemberForbidden() {
	_, ownerTok := s.registerAndLogin("owner@example.com", "Owner", "secret123")
	team := mustDo[teamv1.TeamResponse](s, http.MethodPost, "/api/v1/teams", ownerTok,
		teamv1.CreateTeamRequest{Name: "T"}, http.StatusCreated)

	// приглашаем обычного участника (member)
	memberID, memberTok := s.registerAndLogin("member@example.com", "Member", "secret123")
	path := fmt.Sprintf("/api/v1/teams/%d/invite", team.ID)
	s.requireStatus(s.do(http.MethodPost, path, ownerTok,
		teamv1.InviteRequest{UserID: memberID, Role: "member"}), http.StatusCreated)

	// member не owner/admin → не может приглашать
	third := s.register("third@example.com", "Third", "secret123")
	rsp := s.do(http.MethodPost, path, memberTok, teamv1.InviteRequest{UserID: third.ID, Role: "member"})
	s.requireError(rsp, http.StatusForbidden, "forbidden")
}

func (s *IntegrationSuite) TestInvite_TeamNotFound() {
	_, ownerTok := s.registerAndLogin("owner@example.com", "Owner", "secret123")
	invitee := s.register("inv@example.com", "Inv", "secret123")

	// несуществующая команда → инициатор не её участник → 403 forbidden
	rsp := s.do(http.MethodPost, "/api/v1/teams/999999/invite", ownerTok,
		teamv1.InviteRequest{UserID: invitee.ID, Role: "member"})
	s.requireStatus(rsp, http.StatusForbidden)
}
