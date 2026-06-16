package transaction

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/clients/db/mysql"
)

type manager struct {
	db db.Transactor
}

func New(db db.Transactor) db.TxManager {
	return &manager{db: db}
}

func (m *manager) transaction(ctx context.Context, txOps *sql.TxOptions, f db.Handler) (err error) {
	// Вложенная транзакция: используем уже открытую из контекста, новую не стартуем.
	tx, ok := ctx.Value(mysql.TxKey).(*sql.Tx)
	if ok {
		return f(ctx)
	}

	tx, err = m.db.BeginTx(ctx, txOps)
	if err != nil {
		return errors.Wrap(err, "unable to begin transaction")
	}

	ctx = mysql.MakeContextTx(ctx, tx)

	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("%v", r)
		}

		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrapf(err, "errRollback: %v", errRollback)
			}
			return
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				err = errors.Wrap(err, "unable to commit transaction")
			}
		}
	}()

	if err = f(ctx); err != nil {
		err = errors.Wrap(err, "failed executing code inside transaction")
	}
	return err
}

func (m *manager) ReadCommit(ctx context.Context, f db.Handler) error {
	txOps := &sql.TxOptions{Isolation: sql.LevelReadCommitted}
	return m.transaction(ctx, txOps, f)
}
