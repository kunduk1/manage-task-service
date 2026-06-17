package team

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/circuitbreaker"
	"github.com/kunduk1/manage-task-service/internal/clients/email"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func TestInvite_SuccessOwner(t *testing.T) {
	svc, teamRepo, userRepo, _, emailMock := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)
	userRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(&model.User{ID: 7}, nil)
	teamRepo.EXPECT().AddMember(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, m *model.TeamMember) error {
			assert.Equal(t, int64(1), m.TeamID)
			assert.Equal(t, int64(7), m.UserID)
			assert.Equal(t, model.RoleMember, m.Role)
			return nil
		})
	emailMock.EXPECT().SendInvite(gomock.Any(), gomock.Any()).Return(nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.RoleMember,
	})
	require.NoError(t, err)
}

func TestInvite_SuccessAdminGrantsAdmin(t *testing.T) {
	svc, teamRepo, userRepo, _, emailMock := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(50)).Return(model.RoleAdmin, nil)
	userRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(&model.User{ID: 7}, nil)
	teamRepo.EXPECT().AddMember(gomock.Any(), gomock.Any()).Return(nil)
	emailMock.EXPECT().SendInvite(gomock.Any(), gomock.Any()).Return(nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 50, InviteeID: 7, Role: model.RoleAdmin,
	})
	require.NoError(t, err)
}

// TestInvite_SendsInviteEmail проверяет, что письмо отправляется с корректным payload.
func TestInvite_SendsInviteEmail(t *testing.T) {
	svc, teamRepo, userRepo, _, emailMock := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)
	userRepo.EXPECT().GetByID(gomock.Any(), int64(7)).
		Return(&model.User{ID: 7, Email: "ann@example.com", Name: "Ann"}, nil)
	teamRepo.EXPECT().AddMember(gomock.Any(), gomock.Any()).Return(nil)
	emailMock.EXPECT().SendInvite(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, in email.Invite) error {
			assert.Equal(t, "ann@example.com", in.ToEmail)
			assert.Equal(t, "Ann", in.ToName)
			assert.Equal(t, int64(1), in.TeamID)
			assert.Equal(t, int64(42), in.InviterID)
			assert.Equal(t, string(model.RoleMember), in.Role)
			return nil
		})

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.RoleMember,
	})
	require.NoError(t, err)
}

// TestInvite_SucceedsWhenEmailFails: сбой отправки письма (в т.ч. разомкнутый
// брейкер) не должен ронять приглашение — оно best-effort.
func TestInvite_SucceedsWhenEmailFails(t *testing.T) {
	svc, teamRepo, userRepo, _, emailMock := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)
	userRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(&model.User{ID: 7}, nil)
	teamRepo.EXPECT().AddMember(gomock.Any(), gomock.Any()).Return(nil)
	emailMock.EXPECT().SendInvite(gomock.Any(), gomock.Any()).Return(circuitbreaker.ErrOpen)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.RoleMember,
	})
	require.NoError(t, err)
}

func TestInvite_ForbiddenNotMember(t *testing.T) {
	svc, teamRepo, _, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(99)).
		Return(model.TeamRole(""), errors.ErrNotTeamMember)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 99, InviteeID: 7, Role: model.RoleMember,
	})
	assert.ErrorIs(t, err, errors.ErrForbidden)
}

func TestInvite_ForbiddenPlainMember(t *testing.T) {
	svc, teamRepo, _, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(8)).Return(model.RoleMember, nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 8, InviteeID: 7, Role: model.RoleMember,
	})
	assert.ErrorIs(t, err, errors.ErrForbidden)
}

func TestInvite_InvalidRole(t *testing.T) {
	svc, teamRepo, _, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.TeamRole("boss"),
	})
	assert.ErrorIs(t, err, errors.ErrValidation)
}

func TestInvite_MissingUserID(t *testing.T) {
	svc, teamRepo, _, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 0, Role: model.RoleMember,
	})
	assert.ErrorIs(t, err, errors.ErrValidation)
}

func TestInvite_InviteeNotFound(t *testing.T) {
	svc, teamRepo, userRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)
	userRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(nil, errors.ErrUserNotFound)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.RoleMember,
	})
	assert.ErrorIs(t, err, errors.ErrUserNotFound)
}

func TestInvite_DuplicateMember(t *testing.T) {
	svc, teamRepo, userRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)
	userRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(&model.User{ID: 7}, nil)
	teamRepo.EXPECT().AddMember(gomock.Any(), gomock.Any()).Return(errors.ErrMemberExists)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.RoleMember,
	})
	assert.ErrorIs(t, err, errors.ErrMemberExists)
}
