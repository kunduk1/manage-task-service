package db

//go:generate go tool mockgen -source=db.go -destination=mocks/db_mock.go -package=mocks

import (
	"context"
	"database/sql"
)

// Handler - функция, которая выполняется в транзакции.
type Handler func(ctx context.Context) error

type Client interface {
	DB() DB
	Close() error
}

type Transactor interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type TxManager interface {
	ReadCommit(ctx context.Context, f Handler) error
}

// Query хранит имя запроса (для логирования/трейсинга) и сам SQL.
type Query struct {
	Name     string
	QueryRaw string
}

type SqlExecutor interface {
	NamedExecer
	QueryExecer
}

type NamedExecer interface {
	ScanOneContext(ctx context.Context, dest interface{}, query Query, args ...interface{}) error
	ScanAllContext(ctx context.Context, dest interface{}, query Query, args ...interface{}) error
}

type QueryExecer interface {
	ExecContext(ctx context.Context, query Query, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query Query, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query Query, args ...interface{}) *sql.Row
}

type Ping interface {
	Ping(ctx context.Context) error
}

type DB interface {
	SqlExecutor
	Ping
	Transactor
	Close()
}
