package user

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func (r *repo) Create(ctx context.Context, user *model.User) (int64, error) {
	u := toRepoUser(user)

	q := db.Query{
		Name:     "user_repository.Create",
		QueryRaw: "INSERT INTO users (email, name, password_hash) VALUES (?, ?, ?)",
	}

	res, err := r.db.DB().ExecContext(ctx, q, u.Email, u.Name, u.PasswordHash)
	if err != nil {
		if db.IsDuplicateEntry(err) {
			return 0, errors.ErrUserExists
		}
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
