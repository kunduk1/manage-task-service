package team

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
)

func TestList_PassThrough(t *testing.T) {
	svc, teamRepo, _, _, _ := newTestService(t)
	want := []model.Team{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}}

	teamRepo.EXPECT().ListByUser(gomock.Any(), int64(5)).Return(want, nil)

	got, err := svc.List(context.Background(), 5)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}
