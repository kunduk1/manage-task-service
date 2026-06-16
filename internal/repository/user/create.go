package user

import (
	"context"
	stderrors "errors"

	driver "github.com/go-sql-driver/mysql"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// mysqlDuplicateEntry — код ошибки MySQL для нарушения UNIQUE-ограничения.
const mysqlDuplicateEntry = 1062

func (r *repo) Create(ctx context.Context, user *model.User) (int64, error) {
	u := toRepoUser(user)

	q := db.Query{
		Name:     "user_repository.Create",
		QueryRaw: "INSERT INTO users (email, name, password_hash) VALUES (?, ?, ?)",
	}

	res, err := r.db.DB().ExecContext(ctx, q, u.Email, u.Name, u.PasswordHash)
	if err != nil {
		var mysqlErr *driver.MySQLError
		if stderrors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntry {
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
