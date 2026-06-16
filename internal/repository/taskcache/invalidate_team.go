package taskcache

import "context"

func (r *repo) InvalidateTeam(ctx context.Context, teamID int64) error {
	return r.client.Del(ctx, key(teamID))
}
