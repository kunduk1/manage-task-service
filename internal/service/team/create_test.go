package team

import (
	"context"
	stderrors "errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

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
