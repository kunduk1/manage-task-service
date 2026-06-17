package team

import (
	"context"
	stderrors "errors"
	"testing"

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
			if m.TeamID != 1 || m.UserID != 7 || m.Role != model.RoleMember {
				t.Errorf("unexpected member: %+v", m)
			}
			return nil
		})
	emailMock.EXPECT().SendInvite(gomock.Any(), gomock.Any()).Return(nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.RoleMember,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
			if in.ToEmail != "ann@example.com" || in.ToName != "Ann" {
				t.Errorf("unexpected recipient: %+v", in)
			}
			if in.TeamID != 1 || in.InviterID != 42 || in.Role != string(model.RoleMember) {
				t.Errorf("unexpected invite payload: %+v", in)
			}
			return nil
		})

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.RoleMember,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
	if err != nil {
		t.Fatalf("invite must succeed despite email failure, got: %v", err)
	}
}

func TestInvite_ForbiddenNotMember(t *testing.T) {
	svc, teamRepo, _, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(99)).
		Return(model.TeamRole(""), errors.ErrNotTeamMember)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 99, InviteeID: 7, Role: model.RoleMember,
	})
	if !stderrors.Is(err, errors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestInvite_ForbiddenPlainMember(t *testing.T) {
	svc, teamRepo, _, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(8)).Return(model.RoleMember, nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 8, InviteeID: 7, Role: model.RoleMember,
	})
	if !stderrors.Is(err, errors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestInvite_InvalidRole(t *testing.T) {
	svc, teamRepo, _, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.TeamRole("boss"),
	})
	if !stderrors.Is(err, errors.ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestInvite_MissingUserID(t *testing.T) {
	svc, teamRepo, _, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 0, Role: model.RoleMember,
	})
	if !stderrors.Is(err, errors.ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestInvite_InviteeNotFound(t *testing.T) {
	svc, teamRepo, userRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)
	userRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(nil, errors.ErrUserNotFound)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.RoleMember,
	})
	if !stderrors.Is(err, errors.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestInvite_DuplicateMember(t *testing.T) {
	svc, teamRepo, userRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)
	userRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(&model.User{ID: 7}, nil)
	teamRepo.EXPECT().AddMember(gomock.Any(), gomock.Any()).Return(errors.ErrMemberExists)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.RoleMember,
	})
	if !stderrors.Is(err, errors.ErrMemberExists) {
		t.Errorf("expected ErrMemberExists, got %v", err)
	}
}
