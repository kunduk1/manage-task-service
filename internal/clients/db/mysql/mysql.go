package mysql

import (
	"context"
	"database/sql"
	"log"

	"github.com/georgysavva/scany/v2/sqlscan"

	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/clients/db/prettier"
)

type key string

const TxKey key = "tx"

type mysqlDB struct {
	db    *sql.DB
	debug bool
}

func NewDB(db *sql.DB, debug bool) db.DB {
	return &mysqlDB{db: db, debug: debug}
}

func (m *mysqlDB) ScanOneContext(ctx context.Context, dest interface{}, query db.Query, args ...interface{}) error {

	rows, err := m.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return sqlscan.ScanOne(dest, rows)
}

func (m *mysqlDB) ScanAllContext(ctx context.Context, dest interface{}, query db.Query, args ...interface{}) error {

	rows, err := m.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return sqlscan.ScanAll(dest, rows)
}

func (m *mysqlDB) ExecContext(ctx context.Context, query db.Query, args ...interface{}) (sql.Result, error) {
	m.logQuery(query, args...)

	tx, ok := ctx.Value(TxKey).(*sql.Tx)
	if ok {
		return tx.ExecContext(ctx, query.QueryRaw, args...)
	}

	return m.db.ExecContext(ctx, query.QueryRaw, args...)
}

func (m *mysqlDB) QueryContext(ctx context.Context, query db.Query, args ...interface{}) (*sql.Rows, error) {
	m.logQuery(query, args...)

	tx, ok := ctx.Value(TxKey).(*sql.Tx)
	if ok {
		return tx.QueryContext(ctx, query.QueryRaw, args...)
	}

	return m.db.QueryContext(ctx, query.QueryRaw, args...)
}

func (m *mysqlDB) QueryRowContext(ctx context.Context, query db.Query, args ...interface{}) *sql.Row {
	m.logQuery(query, args...)

	tx, ok := ctx.Value(TxKey).(*sql.Tx)
	if ok {
		return tx.QueryRowContext(ctx, query.QueryRaw, args...)
	}

	return m.db.QueryRowContext(ctx, query.QueryRaw, args...)
}

func (m *mysqlDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.db.BeginTx(ctx, opts)
}

func (m *mysqlDB) Close() {
	if err := m.db.Close(); err != nil {
		log.Printf("failed to close db: %v", err)
	}
}

func (m *mysqlDB) Ping(ctx context.Context) error {
	return m.db.PingContext(ctx)
}

func MakeContextTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, TxKey, tx)
}

// logQuery печатает запрос в лог только при включённом debug-режиме.
// Значения аргументов намеренно не подставляются (чтобы не светить данные в логах) —
// печатается сырой запрос с "?" и количество аргументов.
func (m *mysqlDB) logQuery(query db.Query, args ...interface{}) {
	if !m.debug {
		return
	}

	pretty := prettier.Pretty(query.QueryRaw, prettier.PlaceholderQuestion)
	log.Printf("sql: %s | query: %s | args: %d", query.Name, pretty, len(args))
}
