package taskcomment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repomodel "github.com/kunduk1/manage-task-service/internal/repository/taskcomment/model"
)

func TestToServiceComment(t *testing.T) {
	require.Nil(t, toServiceComment(nil))

	got := toServiceComment(&repomodel.Comment{ID: 1, TaskID: 2, UserID: 3, Body: "hi"})
	assert.Equal(t, int64(2), got.TaskID)
	assert.Equal(t, int64(3), got.UserID)
	assert.Equal(t, "hi", got.Body)

	list := toServiceComments([]repomodel.Comment{{ID: 1, Body: "a"}, {ID: 2, Body: "b"}})
	require.Len(t, list, 2)
	assert.Equal(t, "a", list[0].Body)
	assert.Equal(t, "b", list[1].Body)
}
