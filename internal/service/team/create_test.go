package team

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func TestCreate_Success(t *testing.T) {
	svc, teamRepo, _, txm, _ := newTestService(t)

	runTx(txm)
	gomock.InOrder(
		teamRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, tm *model.Team) (int64, error) {
				assert.Equal(t, "Platform", tm.Name)
				assert.Equal(t, int64(42), tm.CreatedBy)
				return int64(7), nil
			}),
		teamRepo.EXPECT().AddMember(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, m *model.TeamMember) error {
				assert.Equal(t, int64(7), m.TeamID)
				assert.Equal(t, int64(42), m.UserID)
				assert.Equal(t, model.RoleOwner, m.Role)
				return nil
			}),
	)
	teamRepo.EXPECT().GetByID(gomock.Any(), int64(7)).
		Return(&model.Team{ID: 7, Name: "Platform", CreatedBy: 42}, nil)

	team, err := svc.Create(context.Background(), model.CreateTeamInput{Name: "  Platform ", OwnerID: 42})
	require.NoError(t, err)
	assert.Equal(t, int64(7), team.ID)
}

func TestCreate_Validation(t *testing.T) {
	// Пустое имя отсекается до транзакции — никаких EXPECT.
	svc, _, _, _, _ := newTestService(t)

	_, err := svc.Create(context.Background(), model.CreateTeamInput{Name: "   ", OwnerID: 1})
	assert.ErrorIs(t, err, errors.ErrValidation)
}

func TestCreate_AddMemberError(t *testing.T) {
	// Падение второй вставки в транзакции пробрасывается, команда не перечитывается.
	svc, teamRepo, _, txm, _ := newTestService(t)
	errDB := stderrors.New("insert failed")

	runTx(txm)
	gomock.InOrder(
		teamRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(7), nil),
		teamRepo.EXPECT().AddMember(gomock.Any(), gomock.Any()).Return(errDB),
	)

	_, err := svc.Create(context.Background(), model.CreateTeamInput{Name: "Platform", OwnerID: 42})
	assert.ErrorIs(t, err, errDB)
}
