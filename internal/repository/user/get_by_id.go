package user

import (
	"context"

	"github.com/georgysavva/scany/v2/sqlscan"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/user/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func (r *repo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	q := db.Query{
		Name:     "user_repository.GetByID",
		QueryRaw: "SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE id = ?",
	}

	var u repomodel.User
	if err := r.db.DB().ScanOneContext(ctx, &u, q, id); err != nil {
		if sqlscan.NotFound(err) {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}
	return toServiceUser(&u), nil
}
