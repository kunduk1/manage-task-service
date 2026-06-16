package team

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
)

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
