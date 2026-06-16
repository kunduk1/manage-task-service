// Package migrator применяет встроенные SQL-миграции к MySQL на старте приложения.
package migrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	mysqlmig "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"

	// Регистрация драйвера mysql для database/sql.
	_ "github.com/go-sql-driver/mysql"

	"github.com/kunduk1/manage-task-service/internal/logger"
)

// Run применяет все непримененные миграции из fsys к базе по dsn.
//
// Открывает отдельное короткоживущее соединение (не из основного пула):
// для миграций нужен multiStatements=true, который основной пул сознательно не включает.
// Версии отслеживаются в таблице schema_migrations, конкурентный запуск защищён
// advisory-локом MySQL внутри golang-migrate.
func Run(ctx context.Context, dsn string, fsys fs.FS) error {
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}

	sqlDB, err := sql.Open("mysql", dsn+sep+"multiStatements=true")
	if err != nil {
		return fmt.Errorf("open migration db: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err = sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping migration db: %w", err)
	}

	src, err := iofs.New(fsys, ".")
	if err != nil {
		return fmt.Errorf("init migration source: %w", err)
	}
	defer func() { _ = src.Close() }()

	driver, err := mysqlmig.WithInstance(sqlDB, &mysqlmig.Config{})
	if err != nil {
		return fmt.Errorf("init migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}

	logger.Info("running db migrations")
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	version, dirty, verr := m.Version()
	if verr != nil && !errors.Is(verr, migrate.ErrNilVersion) {
		return fmt.Errorf("read migration version: %w", verr)
	}
	logger.Info("db migrations up to date", zap.Uint("version", version), zap.Bool("dirty", dirty))

	return nil
}
