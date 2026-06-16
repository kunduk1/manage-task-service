package taskcomment

import (
	"testing"

	repomodel "github.com/kunduk1/manage-task-service/internal/repository/taskcomment/model"
)

func TestToServiceComment(t *testing.T) {
	if toServiceComment(nil) != nil {
		t.Fatal("expected nil for nil input")
	}

	got := toServiceComment(&repomodel.Comment{ID: 1, TaskID: 2, UserID: 3, Body: "hi"})
	if got.TaskID != 2 || got.UserID != 3 || got.Body != "hi" {
		t.Errorf("unexpected mapping: %+v", got)
	}

	list := toServiceComments([]repomodel.Comment{{ID: 1, Body: "a"}, {ID: 2, Body: "b"}})
	if len(list) != 2 || list[0].Body != "a" || list[1].Body != "b" {
		t.Errorf("unexpected list mapping: %+v", list)
	}
}
