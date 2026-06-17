// Command seed наполняет dev-БД детерминированными фикстурами (см. internal/seed).
// Использование:
//
//	go run ./cmd/seed -config-path local.env -reset
//	make seed
//
// Без -reset команда не трогает уже засеянную БД (печатает подсказку и выходит).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/kunduk1/manage-task-service/internal/clients/db/mysql"
	"github.com/kunduk1/manage-task-service/internal/config"
	"github.com/kunduk1/manage-task-service/internal/logger"
	"github.com/kunduk1/manage-task-service/internal/migrator"
	"github.com/kunduk1/manage-task-service/internal/seed"
	"github.com/kunduk1/manage-task-service/migrations"
)

func main() {
	var (
		configPath string
		reset      bool
	)
	flag.StringVar(&configPath, "config-path", "local.env", "path to env config file")
	flag.BoolVar(&reset, "reset", false, "wipe all domain tables before seeding")
	flag.Parse()

	ctx := context.Background()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// migrator и db-клиент пишут в глобальный логгер — без инициализации он nil и паникует.
	// Свой вывод печатаем через fmt, поэтому zap-логи глушим no-op core.
	logger.NewGlobalLogger(zapcore.NewNopCore())

	// MaxOpenConns=1 критично для seed.Reset: FOREIGN_KEY_CHECKS — сессионная переменная,
	// а пул иначе отдаёт произвольные соединения (см. internal/seed.Reset).
	dbClient, err := mysql.NewClient(ctx, cfg.MySQL.DSN(), mysql.PoolConfig{
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: cfg.MySQL.ConnMaxLifetime(),
	})
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	defer func() { _ = dbClient.Close() }()

	// Гарантируем схему — на случай запуска сразу после `make compose-up` (идемпотентно).
	if err := migrator.Run(ctx, cfg.MySQL.DSN(), migrations.FS); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	seeder := seed.New(dbClient, time.Now())

	if reset {
		if err := seeder.Reset(ctx); err != nil {
			log.Fatalf("reset: %v", err)
		}
		fmt.Println("✓ wiped existing data")
	} else {
		seeded, err := seeder.AlreadySeeded(ctx)
		if err != nil {
			log.Fatalf("check seeded: %v", err)
		}
		if seeded {
			fmt.Println("data already present — re-run with -reset to wipe and reseed")
			return
		}
	}

	sum, err := seeder.Run(ctx)
	if err != nil {
		log.Fatalf("seed: %v", err)
	}

	printSummary(sum)
}

func printSummary(sum seed.Summary) {
	fmt.Printf("✓ seeded: %d users, %d teams, %d members, %d tasks, %d history, %d comments\n",
		sum.Users, sum.Teams, sum.Members, sum.Tasks, sum.History, sum.Comments)
	fmt.Println()
	fmt.Printf("Login with any account below (password: %s):\n", seed.SeedPassword)
	for _, a := range seed.Accounts() {
		fmt.Printf("  %-20s %s\n", a.Email, a.Name)
	}
}
