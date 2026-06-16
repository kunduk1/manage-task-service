package taskcomment

import (
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/taskcomment/model"
)

func toServiceComment(c *repomodel.Comment) *model.TaskComment {
	if c == nil {
		return nil
	}
	return &model.TaskComment{
		ID:        c.ID,
		TaskID:    c.TaskID,
		UserID:    c.UserID,
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func toServiceComments(cs []repomodel.Comment) []model.TaskComment {
	out := make([]model.TaskComment, 0, len(cs))
	for i := range cs {
		out = append(out, *toServiceComment(&cs[i]))
	}
	return out
}
