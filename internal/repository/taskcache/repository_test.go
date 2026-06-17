package taskcache

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	cachemocks "github.com/kunduk1/manage-task-service/internal/clients/cache/mocks"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/internal/repository"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func newTestRepo(t *testing.T) (repository.TaskCacheRepository, *cachemocks.MockClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	client := cachemocks.NewMockClient(ctrl)
	return NewRepository(client), client
}

func TestKey(t *testing.T) {
	assert.Equal(t, "tasks:team:42", key(42))
}

func TestField(t *testing.T) {
	empty := model.TaskFilter{Limit: 20, Offset: 0}
	assert.Equal(t, "s=-|a=-|l=20|o=0", field(empty))

	status := model.StatusInProgress
	assignee := int64(5)
	full := model.TaskFilter{Status: &status, AssigneeID: &assignee, Limit: 50, Offset: 10}
	assert.Equal(t, "s=in_progress|a=5|l=50|o=10", field(full))

	assert.NotEqual(t, field(full), field(empty))
}

func TestGetTaskList_Hit(t *testing.T) {
	repo, client := newTestRepo(t)
	f := model.TaskFilter{Limit: 20}
	raw, err := json.Marshal([]model.Task{{ID: 1}, {ID: 2}})
	require.NoError(t, err)

	// HGetAll отдаёт плоский [field, value, ...]; среди прочих полей наш — последним.
	client.EXPECT().HGetAll(gomock.Any(), key(1)).
		Return([]any{"s=done|a=-|l=20|o=0", "[]", field(f), string(raw)}, nil)

	got, err := repo.GetTaskList(context.Background(), 1, f)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestGetTaskList_Miss(t *testing.T) {
	repo, client := newTestRepo(t)
	// Пустой/несуществующий хэш → промах кэша.
	client.EXPECT().HGetAll(gomock.Any(), key(1)).Return([]any{}, nil)

	_, err := repo.GetTaskList(context.Background(), 1, model.TaskFilter{Limit: 20})
	assert.ErrorIs(t, err, errors.ErrCacheMiss)
}

func TestGetTaskList_ClientError(t *testing.T) {
	repo, client := newTestRepo(t)
	boom := stderrors.New("boom")
	client.EXPECT().HGetAll(gomock.Any(), key(1)).Return(nil, boom)

	_, err := repo.GetTaskList(context.Background(), 1, model.TaskFilter{})
	assert.ErrorIs(t, err, boom)
}

func TestSetTaskList(t *testing.T) {
	repo, client := newTestRepo(t)
	f := model.TaskFilter{Limit: 20}

	client.EXPECT().HSet(gomock.Any(), key(7), gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, values any) error {
			m, ok := values.(map[string]any)
			require.True(t, ok)
			assert.Contains(t, m, field(f))
			return nil
		})
	client.EXPECT().Expire(gomock.Any(), key(7), defaultTTL).Return(nil)

	require.NoError(t, repo.SetTaskList(context.Background(), 7, f, []model.Task{{ID: 1}}))
}

func TestInvalidateTeam(t *testing.T) {
	repo, client := newTestRepo(t)
	client.EXPECT().Del(gomock.Any(), key(3)).Return(nil)

	require.NoError(t, repo.InvalidateTeam(context.Background(), 3))
}
