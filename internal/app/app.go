package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/kunduk1/manage-task-service/internal/api"
	"github.com/kunduk1/manage-task-service/internal/closer"
	"github.com/kunduk1/manage-task-service/internal/config"
	"github.com/kunduk1/manage-task-service/internal/logger"
	"github.com/kunduk1/manage-task-service/internal/migrator"
	"github.com/kunduk1/manage-task-service/migrations"
)

const shutdownTimeout = 10 * time.Second

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", "local.env", "path to env config file")
}

type App struct {
	cfg             *config.Config
	serviceProvider *serviceProvider
	httpServer      *http.Server
}

func NewApp(ctx context.Context) (*App, error) {
	a := &App{}
	if err := a.initDeps(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initConfig,
		a.initServiceProvider,
		a.initLogger,
		a.initMigrations,
		a.initHTTPServer,
	}
	for _, fn := range inits {
		if err := fn(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initConfig(_ context.Context) error {
	if !flag.Parsed() {
		flag.Parse()
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	a.cfg = cfg
	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.serviceProvider = newServiceProvider(a.cfg)
	return nil
}

func (a *App) initLogger(_ context.Context) error {
	logger.NewGlobalLogger(a.loggerCore(a.atomicLevel()))
	return nil
}

func (a *App) initMigrations(ctx context.Context) error {
	if err := migrator.Run(ctx, a.cfg.MySQL.DSN(), migrations.FS); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func (a *App) initHTTPServer(ctx context.Context) error {
	a.httpServer = &http.Server{
		Addr:              a.cfg.HTTP.Address(),
		Handler:           api.NewRouter(a.serviceProvider.AuthHandler(ctx)),
		ReadHeaderTimeout: 15 * time.Second,
	}
	return nil
}

// Run запускает HTTP-сервер и блокируется до фатальной ошибки или сигнала остановки.
func (a *App) Run() error {
	closer.Add(func() error {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return a.httpServer.Shutdown(shutdownCtx)
	})

	logger.Info("starting http server", zap.String("address", a.cfg.HTTP.Address()))

	errCh := make(chan error, 1)
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	runErr := <-errCh

	closer.CloseAll()
	closer.Wait()

	if runErr != nil {
		logger.Error("http server stopped with error", zap.Error(runErr))
	}
	return runErr
}

func (a *App) loggerCore(level zap.AtomicLevel) zapcore.Core {
	stdout := zapcore.AddSync(os.Stdout)
	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   a.cfg.Logger.LogPath(),
		MaxSize:    a.cfg.Logger.MaxSize(), // megabytes
		MaxBackups: a.cfg.Logger.MaxBackups(),
		MaxAge:     a.cfg.Logger.MaxAge(), // days
	})

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	return zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
		zapcore.NewCore(fileEncoder, file, level),
	)
}

func (a *App) atomicLevel() zap.AtomicLevel {
	var level zapcore.Level
	if err := level.Set(a.cfg.Logger.Level()); err != nil {
		log.Fatalf("failed to set log level: %v", err)
	}
	return zap.NewAtomicLevelAt(level)
}
