package taskcache

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"testing"

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
	if got := key(42); got != "tasks:team:42" {
		t.Errorf("unexpected key: %q", got)
	}
}

func TestField(t *testing.T) {
	empty := model.TaskFilter{Limit: 20, Offset: 0}
	if got := field(empty); got != "s=-|a=-|l=20|o=0" {
		t.Errorf("unexpected field for empty filter: %q", got)
	}

	status := model.StatusInProgress
	assignee := int64(5)
	full := model.TaskFilter{Status: &status, AssigneeID: &assignee, Limit: 50, Offset: 10}
	if got := field(full); got != "s=in_progress|a=5|l=50|o=10" {
		t.Errorf("unexpected field for full filter: %q", got)
	}

	if field(empty) == field(full) {
		t.Error("different filters must produce different fields")
	}
}

func TestGetTaskList_Hit(t *testing.T) {
	repo, client := newTestRepo(t)
	f := model.TaskFilter{Limit: 20}
	raw, err := json.Marshal([]model.Task{{ID: 1}, {ID: 2}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// HGetAll отдаёт плоский [field, value, ...]; среди прочих полей наш — последним.
	client.EXPECT().HGetAll(gomock.Any(), key(1)).
		Return([]any{"s=done|a=-|l=20|o=0", "[]", field(f), string(raw)}, nil)

	got, err := repo.GetTaskList(context.Background(), 1, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(got))
	}
}

func TestGetTaskList_Miss(t *testing.T) {
	repo, client := newTestRepo(t)
	// Пустой/несуществующий хэш → промах кэша.
	client.EXPECT().HGetAll(gomock.Any(), key(1)).Return([]any{}, nil)

	_, err := repo.GetTaskList(context.Background(), 1, model.TaskFilter{Limit: 20})
	if !stderrors.Is(err, errors.ErrCacheMiss) {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

func TestGetTaskList_ClientError(t *testing.T) {
	repo, client := newTestRepo(t)
	boom := stderrors.New("boom")
	client.EXPECT().HGetAll(gomock.Any(), key(1)).Return(nil, boom)

	_, err := repo.GetTaskList(context.Background(), 1, model.TaskFilter{})
	if !stderrors.Is(err, boom) {
		t.Errorf("expected boom error, got %v", err)
	}
}

func TestSetTaskList(t *testing.T) {
	repo, client := newTestRepo(t)
	f := model.TaskFilter{Limit: 20}

	client.EXPECT().HSet(gomock.Any(), key(7), gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, values any) error {
			m, ok := values.(map[string]any)
			if !ok {
				t.Fatalf("expected map values, got %T", values)
			}
			if _, ok := m[field(f)]; !ok {
				t.Errorf("expected field %q in hset values, got %v", field(f), m)
			}
			return nil
		})
	client.EXPECT().Expire(gomock.Any(), key(7), defaultTTL).Return(nil)

	if err := repo.SetTaskList(context.Background(), 7, f, []model.Task{{ID: 1}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvalidateTeam(t *testing.T) {
	repo, client := newTestRepo(t)
	client.EXPECT().Del(gomock.Any(), key(3)).Return(nil)

	if err := repo.InvalidateTeam(context.Background(), 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
