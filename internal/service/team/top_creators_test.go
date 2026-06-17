package team

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
)

func TestTopCreators_PassThrough(t *testing.T) {
	svc, teamRepo, _, _, _ := newTestService(t)
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	want := []model.TopCreator{
		{TeamID: 1, UserID: 42, UserName: "Alice", CreatedCount: 7, Rank: 1},
	}

	// Проверяем, что период доходит до репозитория без изменений.
	teamRepo.EXPECT().TopCreators(gomock.Any(), from, to).Return(want, nil)

	got, err := svc.TopCreators(context.Background(), from, to)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
