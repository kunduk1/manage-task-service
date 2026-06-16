package team

import (
	"context"
	stderrors "errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	dbmocks "github.com/kunduk1/manage-task-service/internal/clients/db/mocks"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomocks "github.com/kunduk1/manage-task-service/internal/repository/mocks"
	"github.com/kunduk1/manage-task-service/internal/service"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// newTestService собирает сервис с gomock-моками репозиториев и менеджера транзакций.
func newTestService(t *testing.T) (
	service.TeamsService,
	*repomocks.MockTeamRepository,
	*repomocks.MockUserRepository,
	*dbmocks.MockTxManager,
) {
	t.Helper()
	ctrl := gomock.NewController(t)
	teamRepo := repomocks.NewMockTeamRepository(ctrl)
	userRepo := repomocks.NewMockUserRepository(ctrl)
	txm := dbmocks.NewMockTxManager(ctrl)
	svc := NewService(teamRepo, userRepo, txm)
	return svc, teamRepo, userRepo, txm
}

// runTx — заглушка ReadCommit, исполняющая переданную функцию в том же контексте.
func runTx(txm *dbmocks.MockTxManager) {
	txm.EXPECT().ReadCommit(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, f db.Handler) error { return f(ctx) })
}

// --- Create ---

func TestCreate_Success(t *testing.T) {
	svc, teamRepo, _, txm := newTestService(t)

	runTx(txm)
	gomock.InOrder(
		teamRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, tm *model.Team) (int64, error) {
				if tm.Name != "Platform" {
					t.Errorf("expected trimmed name Platform, got %q", tm.Name)
				}
				if tm.CreatedBy != 42 {
					t.Errorf("expected created_by 42, got %d", tm.CreatedBy)
				}
				return int64(7), nil
			}),
		teamRepo.EXPECT().AddMember(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, m *model.TeamMember) error {
				if m.TeamID != 7 || m.UserID != 42 || m.Role != model.RoleOwner {
					t.Errorf("owner must be added to team: %+v", m)
				}
				return nil
			}),
	)
	teamRepo.EXPECT().GetByID(gomock.Any(), int64(7)).
		Return(&model.Team{ID: 7, Name: "Platform", CreatedBy: 42}, nil)

	team, err := svc.Create(context.Background(), model.CreateTeamInput{Name: "  Platform ", OwnerID: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team.ID != 7 {
		t.Errorf("expected id from repository, got %d", team.ID)
	}
}

func TestCreate_Validation(t *testing.T) {
	// Пустое имя отсекается до транзакции — никаких EXPECT.
	svc, _, _, _ := newTestService(t)

	_, err := svc.Create(context.Background(), model.CreateTeamInput{Name: "   ", OwnerID: 1})
	if !stderrors.Is(err, errors.ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestCreate_AddMemberError(t *testing.T) {
	// Падение второй вставки в транзакции пробрасывается, команда не перечитывается.
	svc, teamRepo, _, txm := newTestService(t)
	errDB := stderrors.New("insert failed")

	runTx(txm)
	gomock.InOrder(
		teamRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(7), nil),
		teamRepo.EXPECT().AddMember(gomock.Any(), gomock.Any()).Return(errDB),
	)

	_, err := svc.Create(context.Background(), model.CreateTeamInput{Name: "Platform", OwnerID: 42})
	if !stderrors.Is(err, errDB) {
		t.Errorf("expected propagated error, got %v", err)
	}
}

// --- List ---

func TestList_PassThrough(t *testing.T) {
	svc, teamRepo, _, _ := newTestService(t)
	want := []model.Team{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}}

	teamRepo.EXPECT().ListByUser(gomock.Any(), int64(5)).Return(want, nil)

	got, err := svc.List(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 teams, got %d", len(got))
	}
}

// --- Invite ---

func TestInvite_SuccessOwner(t *testing.T) {
	svc, teamRepo, userRepo, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)
	userRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(&model.User{ID: 7}, nil)
	teamRepo.EXPECT().AddMember(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, m *model.TeamMember) error {
			if m.TeamID != 1 || m.UserID != 7 || m.Role != model.RoleMember {
				t.Errorf("unexpected member: %+v", m)
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

func TestInvite_SuccessAdminGrantsAdmin(t *testing.T) {
	svc, teamRepo, userRepo, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(50)).Return(model.RoleAdmin, nil)
	userRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(&model.User{ID: 7}, nil)
	teamRepo.EXPECT().AddMember(gomock.Any(), gomock.Any()).Return(nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 50, InviteeID: 7, Role: model.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvite_ForbiddenNotMember(t *testing.T) {
	svc, teamRepo, _, _ := newTestService(t)

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
	svc, teamRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(8)).Return(model.RoleMember, nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 8, InviteeID: 7, Role: model.RoleMember,
	})
	if !stderrors.Is(err, errors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestInvite_InvalidRole(t *testing.T) {
	svc, teamRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 7, Role: model.TeamRole("boss"),
	})
	if !stderrors.Is(err, errors.ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestInvite_MissingUserID(t *testing.T) {
	svc, teamRepo, _, _ := newTestService(t)

	teamRepo.EXPECT().GetMemberRole(gomock.Any(), int64(1), int64(42)).Return(model.RoleOwner, nil)

	err := svc.Invite(context.Background(), model.InviteInput{
		TeamID: 1, ActorID: 42, InviteeID: 0, Role: model.RoleMember,
	})
	if !stderrors.Is(err, errors.ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestInvite_InviteeNotFound(t *testing.T) {
	svc, teamRepo, userRepo, _ := newTestService(t)

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
	svc, teamRepo, userRepo, _ := newTestService(t)

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
