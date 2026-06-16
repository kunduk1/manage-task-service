package user

import (
	"context"

	"github.com/georgysavva/scany/v2/sqlscan"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	repomodel "github.com/kunduk1/manage-task-service/internal/repository/user/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func (r *repo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	q := db.Query{
		Name:     "user_repository.GetByEmail",
		QueryRaw: "SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE email = ?",
	}

	var u repomodel.User
	if err := r.db.DB().ScanOneContext(ctx, &u, q, email); err != nil {
		if sqlscan.NotFound(err) {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}
	return toServiceUser(&u), nil
}
