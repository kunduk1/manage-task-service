package authz

import (
	"context"
	stderrors "errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	repomocks "github.com/kunduk1/manage-task-service/internal/repository/mocks"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

var errDB = stderrors.New("db boom")

func TestRequireMember(t *testing.T) {
	tests := []struct {
		name     string
		role     model.TeamRole
		repoErr  error
		wantRole model.TeamRole
		wantErr  error
	}{
		{name: "member returns role", role: model.RoleAdmin, wantRole: model.RoleAdmin},
		{name: "not member maps to forbidden", repoErr: errors.ErrNotTeamMember, wantErr: errors.ErrForbidden},
		{name: "other error propagated", repoErr: errDB, wantErr: errDB},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			teamRepo := repomocks.NewMockTeamRepository(ctrl)
			teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(2)).Return(tt.role, tt.repoErr)

			role, err := New(teamRepo).RequireMember(context.Background(), 1, 2)
			if !stderrors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && role != tt.wantRole {
				t.Fatalf("role = %v, want %v", role, tt.wantRole)
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	ownerOrAdmin := []model.TeamRole{model.RoleOwner, model.RoleAdmin}
	tests := []struct {
		name    string
		role    model.TeamRole
		repoErr error
		allowed []model.TeamRole
		wantErr error
	}{
		{name: "allowed role passes", role: model.RoleOwner, allowed: ownerOrAdmin},
		{name: "disallowed role forbidden", role: model.RoleMember, allowed: ownerOrAdmin, wantErr: errors.ErrForbidden},
		{name: "not member forbidden", repoErr: errors.ErrNotTeamMember, allowed: ownerOrAdmin, wantErr: errors.ErrForbidden},
		{name: "repo error propagated", repoErr: errDB, allowed: ownerOrAdmin, wantErr: errDB},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			teamRepo := repomocks.NewMockTeamRepository(ctrl)
			teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(2)).Return(tt.role, tt.repoErr)

			err := New(teamRepo).RequireRole(context.Background(), 1, 2, tt.allowed...)
			if !stderrors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// MemberRole — базовый примитив: ErrNotTeamMember не маппится, а пробрасывается,
// чтобы вызывающий мог применить свой маппинг (например ErrValidation).
func TestMemberRole_PropagatesNotMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	teamRepo := repomocks.NewMockTeamRepository(ctrl)
	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(2)).
		Return(model.TeamRole(""), errors.ErrNotTeamMember)

	_, err := New(teamRepo).MemberRole(context.Background(), 1, 2)
	if !stderrors.Is(err, errors.ErrNotTeamMember) {
		t.Fatalf("err = %v, want ErrNotTeamMember", err)
	}
}
