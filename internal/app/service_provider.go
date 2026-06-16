package app

import (
	"context"
	"log"

	authapi "github.com/kunduk1/manage-task-service/internal/api/auth"
	taskapi "github.com/kunduk1/manage-task-service/internal/api/task"
	teamapi "github.com/kunduk1/manage-task-service/internal/api/team"
	"github.com/kunduk1/manage-task-service/internal/clients/cache"
	redisCache "github.com/kunduk1/manage-task-service/internal/clients/cache/redis"
	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/clients/db/mysql"
	"github.com/kunduk1/manage-task-service/internal/clients/db/transaction"
	"github.com/kunduk1/manage-task-service/internal/closer"
	"github.com/kunduk1/manage-task-service/internal/config"
	"github.com/kunduk1/manage-task-service/internal/metrics"
	"github.com/kunduk1/manage-task-service/internal/ratelimit"
	"github.com/kunduk1/manage-task-service/internal/repository"
	taskRepo "github.com/kunduk1/manage-task-service/internal/repository/task"
	taskCacheRepo "github.com/kunduk1/manage-task-service/internal/repository/taskcache"
	taskHistoryRepo "github.com/kunduk1/manage-task-service/internal/repository/taskhistory"
	teamRepo "github.com/kunduk1/manage-task-service/internal/repository/team"
	tokenRepo "github.com/kunduk1/manage-task-service/internal/repository/token"
	userRepo "github.com/kunduk1/manage-task-service/internal/repository/user"
	"github.com/kunduk1/manage-task-service/internal/service"
	authService "github.com/kunduk1/manage-task-service/internal/service/auth"
	"github.com/kunduk1/manage-task-service/internal/service/authz"
	taskService "github.com/kunduk1/manage-task-service/internal/service/task"
	teamService "github.com/kunduk1/manage-task-service/internal/service/team"
	"github.com/kunduk1/manage-task-service/internal/token"
)

// serviceProvider — ленивый DI-контейнер: создаёт зависимости по требованию и кэширует их.
type serviceProvider struct {
	cfg *config.Config

	dbClient    db.Client
	cacheClient cache.Client
	txManager   db.TxManager

	userRepository        repository.UserRepository
	tokenRepository       repository.TokenRepository
	teamRepository        repository.TeamRepository
	taskRepository        repository.TaskRepository
	taskCacheRepository   repository.TaskCacheRepository
	taskHistoryRepository repository.TaskHistoryRepository

	metrics     *metrics.Metrics
	rateLimiter *ratelimit.Limiter

	jwtManager  *token.Manager
	authorizer  *authz.Authorizer
	authService service.AuthService
	authHandler *authapi.Handler
	teamService service.TeamsService
	teamHandler *teamapi.Handler
	taskService service.TasksService
	taskHandler *taskapi.Handler
}

func newServiceProvider(cfg *config.Config) *serviceProvider {
	return &serviceProvider{cfg: cfg}
}

func (s *serviceProvider) DBClient(ctx context.Context) db.Client {
	if s.dbClient == nil {
		cl, err := mysql.NewClient(ctx, s.cfg.MySQL.DSN(), mysql.PoolConfig{
			MaxOpenConns:    s.cfg.MySQL.MaxOpenConns(),
			MaxIdleConns:    s.cfg.MySQL.MaxIdleConns(),
			ConnMaxLifetime: s.cfg.MySQL.ConnMaxLifetime(),
		})
		if err != nil {
			log.Fatalf("failed to create db client: %v", err)
		}
		closer.Add(cl.Close)
		s.dbClient = cl
	}
	return s.dbClient
}

func (s *serviceProvider) CacheClient(ctx context.Context) cache.Client {
	if s.cacheClient == nil {
		cl, err := redisCache.NewClient(ctx, s.cfg.Redis)
		if err != nil {
			log.Fatalf("failed to create cache client: %v", err)
		}
		closer.Add(cl.Close)
		s.cacheClient = cl
	}
	return s.cacheClient
}

// TxManager — менеджер транзакций (задел под будущие многошаговые операции).
func (s *serviceProvider) TxManager(ctx context.Context) db.TxManager {
	if s.txManager == nil {
		s.txManager = transaction.New(s.DBClient(ctx).DB())
	}
	return s.txManager
}

func (s *serviceProvider) UserRepository(ctx context.Context) repository.UserRepository {
	if s.userRepository == nil {
		s.userRepository = userRepo.NewRepository(s.DBClient(ctx))
	}
	return s.userRepository
}

func (s *serviceProvider) TokenRepository(ctx context.Context) repository.TokenRepository {
	if s.tokenRepository == nil {
		s.tokenRepository = tokenRepo.NewRepository(s.CacheClient(ctx))
	}
	return s.tokenRepository
}

func (s *serviceProvider) Metrics() *metrics.Metrics {
	if s.metrics == nil {
		s.metrics = metrics.New()
	}
	return s.metrics
}

// RateLimiter возвращает per-user лимитер запросов или nil, если фича выключена
// (в этом случае router не навешивает соответствующий middleware).
func (s *serviceProvider) RateLimiter(ctx context.Context) *ratelimit.Limiter {
	if !s.cfg.RateLimit.Enabled() {
		return nil
	}
	if s.rateLimiter == nil {
		s.rateLimiter = ratelimit.New(
			s.CacheClient(ctx),
			s.cfg.RateLimit.Requests(),
			s.cfg.RateLimit.Window(),
		)
	}
	return s.rateLimiter
}

func (s *serviceProvider) JWTManager() *token.Manager {
	if s.jwtManager == nil {
		s.jwtManager = token.NewManager(s.cfg.JWT.Secret(), s.cfg.JWT.AccessTTL())
	}
	return s.jwtManager
}

func (s *serviceProvider) AuthService(ctx context.Context) service.AuthService {
	if s.authService == nil {
		s.authService = authService.NewService(
			s.UserRepository(ctx),
			s.TokenRepository(ctx),
			s.JWTManager(),
			s.cfg.JWT.RefreshTTL(),
		)
	}
	return s.authService
}

func (s *serviceProvider) AuthHandler(ctx context.Context) *authapi.Handler {
	if s.authHandler == nil {
		s.authHandler = authapi.NewHandler(s.AuthService(ctx))
	}
	return s.authHandler
}

func (s *serviceProvider) TeamRepository(ctx context.Context) repository.TeamRepository {
	if s.teamRepository == nil {
		s.teamRepository = teamRepo.NewRepository(s.DBClient(ctx))
	}
	return s.teamRepository
}

func (s *serviceProvider) Authorizer(ctx context.Context) *authz.Authorizer {
	if s.authorizer == nil {
		s.authorizer = authz.New(s.TeamRepository(ctx))
	}
	return s.authorizer
}

func (s *serviceProvider) TeamService(ctx context.Context) service.TeamsService {
	if s.teamService == nil {
		s.teamService = teamService.NewService(
			s.TeamRepository(ctx),
			s.UserRepository(ctx),
			s.TxManager(ctx),
			s.Authorizer(ctx),
		)
	}
	return s.teamService
}

func (s *serviceProvider) TeamHandler(ctx context.Context) *teamapi.Handler {
	if s.teamHandler == nil {
		s.teamHandler = teamapi.NewHandler(s.TeamService(ctx))
	}
	return s.teamHandler
}

func (s *serviceProvider) TaskRepository(ctx context.Context) repository.TaskRepository {
	if s.taskRepository == nil {
		s.taskRepository = taskRepo.NewRepository(s.DBClient(ctx))
	}
	return s.taskRepository
}

func (s *serviceProvider) TaskCacheRepository(ctx context.Context) repository.TaskCacheRepository {
	if s.taskCacheRepository == nil {
		s.taskCacheRepository = taskCacheRepo.NewRepository(s.CacheClient(ctx))
	}
	return s.taskCacheRepository
}

func (s *serviceProvider) TaskHistoryRepository(ctx context.Context) repository.TaskHistoryRepository {
	if s.taskHistoryRepository == nil {
		s.taskHistoryRepository = taskHistoryRepo.NewRepository(s.DBClient(ctx))
	}
	return s.taskHistoryRepository
}

func (s *serviceProvider) TaskService(ctx context.Context) service.TasksService {
	if s.taskService == nil {
		s.taskService = taskService.NewService(
			s.TaskRepository(ctx),
			s.TaskHistoryRepository(ctx),
			s.TaskCacheRepository(ctx),
			s.TxManager(ctx),
			s.Authorizer(ctx),
		)
	}
	return s.taskService
}

func (s *serviceProvider) TaskHandler(ctx context.Context) *taskapi.Handler {
	if s.taskHandler == nil {
		s.taskHandler = taskapi.NewHandler(s.TaskService(ctx))
	}
	return s.taskHandler
}
