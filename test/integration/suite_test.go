//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	goredis "github.com/redis/go-redis/v9"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"go.uber.org/zap/zapcore"

	"github.com/kunduk1/manage-task-service/internal/api"
	authapi "github.com/kunduk1/manage-task-service/internal/api/auth"
	taskapi "github.com/kunduk1/manage-task-service/internal/api/task"
	teamapi "github.com/kunduk1/manage-task-service/internal/api/team"
	"github.com/kunduk1/manage-task-service/internal/clients/cache"
	rediscache "github.com/kunduk1/manage-task-service/internal/clients/cache/redis"
	"github.com/kunduk1/manage-task-service/internal/clients/db"
	"github.com/kunduk1/manage-task-service/internal/clients/db/mysql"
	"github.com/kunduk1/manage-task-service/internal/clients/db/transaction"
	"github.com/kunduk1/manage-task-service/internal/config"
	"github.com/kunduk1/manage-task-service/internal/logger"
	"github.com/kunduk1/manage-task-service/internal/metrics"
	"github.com/kunduk1/manage-task-service/internal/migrator"
	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/internal/repository"
	taskrepo "github.com/kunduk1/manage-task-service/internal/repository/task"
	taskcacherepo "github.com/kunduk1/manage-task-service/internal/repository/taskcache"
	taskhistoryrepo "github.com/kunduk1/manage-task-service/internal/repository/taskhistory"
	teamrepo "github.com/kunduk1/manage-task-service/internal/repository/team"
	tokenrepo "github.com/kunduk1/manage-task-service/internal/repository/token"
	userrepo "github.com/kunduk1/manage-task-service/internal/repository/user"
	authsvc "github.com/kunduk1/manage-task-service/internal/service/auth"
	"github.com/kunduk1/manage-task-service/internal/service/authz"
	tasksvc "github.com/kunduk1/manage-task-service/internal/service/task"
	teamsvc "github.com/kunduk1/manage-task-service/internal/service/team"
	"github.com/kunduk1/manage-task-service/internal/token"
	"github.com/kunduk1/manage-task-service/migrations"
	authv1 "github.com/kunduk1/manage-task-service/pkg/auth/v1"
)

// taskCacheKeyPrefix зеркалит taskcache.keyPrefix (internal, поэтому дублируем):
// ключ Redis-хэша со списками задач команды.
const taskCacheKeyPrefix = "tasks:team:"

// Тесты идут последовательно и делят одни MySQL/Redis — t.Parallel() не использовать.
type IntegrationSuite struct {
	suite.Suite

	ctx    context.Context
	mysqlC *tcmysql.MySQLContainer
	redisC *tcredis.RedisContainer

	srv   *httptest.Server
	jwt   *token.Manager
	db    db.Client
	cache cache.Client
	redis *goredis.Client // админ-клиент: FLUSHDB и проверка ключей кэша

	userRepo        repository.UserRepository
	tokenRepo       repository.TokenRepository
	teamRepo        repository.TeamRepository
	taskRepo        repository.TaskRepository
	taskCacheRepo   repository.TaskCacheRepository
	taskHistoryRepo repository.TaskHistoryRepository
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

// SetupSuite поднимает контейнеры, прогоняет миграции и собирает реальный роутер — один раз.
func (s *IntegrationSuite) SetupSuite() {
	// migrator.Run и часть инфраструктуры пишут в глобальный логгер — без инициализации
	// logger.Info паникует на nil. Тестам логи не нужны → no-op core.
	logger.NewGlobalLogger(zapcore.NewNopCore())
	s.ctx = context.Background()
	r := s.Require()

	// 1. Контейнеры MySQL + Redis (модули testcontainers сами ждут готовности).
	mysqlC, err := tcmysql.Run(s.ctx, "mysql:8.4",
		tcmysql.WithDatabase("manage_task_test"),
		tcmysql.WithUsername("app"),
		tcmysql.WithPassword("app"),
	)
	r.NoError(err, "start mysql container (нужен запущенный Docker)")
	s.mysqlC = mysqlC

	redisC, err := tcredis.Run(s.ctx, "redis:7-alpine")
	r.NoError(err, "start redis container")
	s.redisC = redisC

	// 2. Координаты контейнеров → переменные окружения, которые читает config.Load.
	r.NoError(s.exportContainerEnv())

	// 3. Конфиг из окружения (пустой путь => без .env-файла).
	cfg, err := config.Load("")
	r.NoError(err, "load config")

	// 4. Миграции — та же продакшн-логика, что и на старте приложения.
	r.NoError(migrator.Run(s.ctx, cfg.MySQL.DSN(), migrations.FS), "run migrations")

	// 5. Клиенты.
	dbClient, err := mysql.NewClient(s.ctx, cfg.MySQL.DSN(), mysql.PoolConfig{
		MaxOpenConns:    cfg.MySQL.MaxOpenConns(),
		MaxIdleConns:    cfg.MySQL.MaxIdleConns(),
		ConnMaxLifetime: cfg.MySQL.ConnMaxLifetime(),
	})
	r.NoError(err, "db client")
	s.db = dbClient

	cacheClient, err := rediscache.NewClient(s.ctx, cfg.Redis)
	r.NoError(err, "cache client")
	s.cache = cacheClient
	s.redis = goredis.NewClient(&goredis.Options{Addr: cfg.Redis.Addr()})

	s.jwt = token.NewManager(cfg.JWT.Secret(), cfg.JWT.AccessTTL())

	// 6. Репозитории (общие для роутера и для сидов).
	s.userRepo = userrepo.NewRepository(dbClient)
	s.tokenRepo = tokenrepo.NewRepository(cacheClient)
	s.teamRepo = teamrepo.NewRepository(dbClient)
	s.taskRepo = taskrepo.NewRepository(dbClient)
	s.taskCacheRepo = taskcacherepo.NewRepository(cacheClient)
	s.taskHistoryRepo = taskhistoryrepo.NewRepository(dbClient)

	authorizer := authz.New(s.teamRepo)
	txMgr := transaction.New(dbClient.DB())

	// 7. Реальный роутер. rateLimiter=nil и emailClient=nil — поддержанные пути
	// «фича выключена»: тесты не троттлятся и не зависят от почты.
	handler := api.NewRouter(
		authapi.NewHandler(authsvc.NewService(s.userRepo, s.tokenRepo, s.jwt, cfg.JWT.RefreshTTL())),
		teamapi.NewHandler(teamsvc.NewService(s.teamRepo, s.userRepo, txMgr, authorizer, nil)),
		taskapi.NewHandler(tasksvc.NewService(s.taskRepo, s.taskHistoryRepo, s.taskCacheRepo, txMgr, authorizer)),
		s.jwt,
		metrics.New(),
		nil,
	)
	s.srv = httptest.NewServer(handler)
}

// TearDownSuite закрывает клиенты и гасит контейнеры.
func (s *IntegrationSuite) TearDownSuite() {
	if s.srv != nil {
		s.srv.Close()
	}
	if s.redis != nil {
		_ = s.redis.Close()
	}
	if s.cache != nil {
		_ = s.cache.Close()
	}
	if s.db != nil {
		_ = s.db.Close()
	}
	if s.redisC != nil {
		_ = s.redisC.Terminate(s.ctx)
	}
	if s.mysqlC != nil {
		_ = s.mysqlC.Terminate(s.ctx)
	}
}

// SetupTest приводит хранилища к пустому состоянию перед каждым тестом.
func (s *IntegrationSuite) SetupTest() {
	s.Require().NoError(s.reset())
}

// reset чистит БД и Redis. DELETE в порядке потомок→родитель уважает внешние ключи без
// отключения проверок (на FOREIGN_KEY_CHECKS полагаться нельзя: пул отдаёт произвольные
// соединения, а переменная сессионная). schema_migrations не трогаем.
func (s *IntegrationSuite) reset() error {
	tables := []string{"task_history", "task_comments", "tasks", "team_members", "teams", "users"}
	for _, name := range tables {
		q := db.Query{Name: "integration.reset", QueryRaw: "DELETE FROM " + name}
		if _, err := s.db.DB().ExecContext(s.ctx, q); err != nil {
			return err
		}
	}
	return s.redis.FlushDB(s.ctx).Err()
}

// exportContainerEnv прокидывает host:port контейнеров в переменные окружения,
// которые читают конфиг-секции (см. internal/config/*).
func (s *IntegrationSuite) exportContainerEnv() error {
	mHost, err := s.mysqlC.Host(s.ctx)
	if err != nil {
		return err
	}
	mPort, err := s.mysqlC.MappedPort(s.ctx, "3306/tcp")
	if err != nil {
		return err
	}
	rHost, err := s.redisC.Host(s.ctx)
	if err != nil {
		return err
	}
	rPort, err := s.redisC.MappedPort(s.ctx, "6379/tcp")
	if err != nil {
		return err
	}

	env := map[string]string{
		"MYSQL_HOST":         mHost,
		"MYSQL_PORT":         mPort.Port(),
		"MYSQL_USER":         "app",
		"MYSQL_PASSWORD":     "app",
		"MYSQL_DATABASE":     "manage_task_test",
		"REDIS_ADDR":         net.JoinHostPort(rHost, rPort.Port()),
		"JWT_SECRET":         "integration-secret",
		"RATE_LIMIT_ENABLED": "false",
		"EMAIL_ENABLED":      "false",
	}
	for k, v := range env {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

// --- транспорт ---

// apiResp — сырой ответ тестового сервера.
type apiResp struct {
	Status int
	Body   []byte
}

type errorBody struct {
	Error string `json:"error"`
}

// do выполняет HTTP-запрос к тестовому серверу. body: nil — без тела; []byte/string —
// отправляется как есть (для проверки «битого» JSON); иначе — JSON-маршалинг.
func (s *IntegrationSuite) do(method, path, bearer string, body any) apiResp {
	s.T().Helper()
	r := s.Require()

	var reader io.Reader
	switch b := body.(type) {
	case nil:
		reader = nil
	case []byte:
		reader = bytes.NewReader(b)
	case string:
		reader = strings.NewReader(b)
	default:
		raw, err := json.Marshal(b)
		r.NoError(err, "marshal body")
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, s.srv.URL+path, reader)
	r.NoError(err, "new request")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}

	resp, err := s.srv.Client().Do(req)
	r.NoErrorf(err, "do %s %s", method, path)
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	r.NoError(err, "read body")
	return apiResp{Status: resp.StatusCode, Body: data}
}

// requireStatus проверяет HTTP-статус (с телом в сообщении при несовпадении).
func (s *IntegrationSuite) requireStatus(rsp apiResp, want int) {
	s.T().Helper()
	s.Require().Equalf(want, rsp.Status, "body=%s", rsp.Body)
}

// requireError проверяет статус и тело-ошибку формата {"error":"..."}.
func (s *IntegrationSuite) requireError(rsp apiResp, wantStatus int, wantMsg string) {
	s.T().Helper()
	s.requireStatus(rsp, wantStatus)
	eb := decode[errorBody](s, rsp)
	s.Equal(wantMsg, eb.Error)
}

// decode разбирает JSON-тело ответа в T (свободная функция — методы не могут быть дженериками).
func decode[T any](s *IntegrationSuite, rsp apiResp) T {
	s.T().Helper()
	var out T
	s.Require().NoErrorf(json.Unmarshal(rsp.Body, &out), "decode body %s", rsp.Body)
	return out
}

// mustDo делает запрос, проверяет статус и декодирует тело в T.
func mustDo[T any](s *IntegrationSuite, method, path, bearer string, body any, want int) T {
	s.T().Helper()
	rsp := s.do(method, path, bearer, body)
	s.requireStatus(rsp, want)
	return decode[T](s, rsp)
}

// --- auth fast-paths: через реальные HTTP-ручки (путь под тестом всегда «настоящий») ---

func (s *IntegrationSuite) register(email, name, password string) authv1.RegisterResponse {
	s.T().Helper()
	return mustDo[authv1.RegisterResponse](s, http.MethodPost, "/api/v1/register", "",
		authv1.RegisterRequest{Email: email, Name: name, Password: password}, http.StatusCreated)
}

func (s *IntegrationSuite) login(email, password string) authv1.LoginResponse {
	s.T().Helper()
	return mustDo[authv1.LoginResponse](s, http.MethodPost, "/api/v1/login", "",
		authv1.LoginRequest{Email: email, Password: password}, http.StatusOK)
}

// registerAndLogin регистрирует пользователя и возвращает его id и access-токен.
func (s *IntegrationSuite) registerAndLogin(email, name, password string) (int64, string) {
	s.T().Helper()
	reg := s.register(email, name, password)
	tok := s.login(email, password)
	return reg.ID, tok.AccessToken
}

// --- seed fast-paths: прямой insert через repo для предусловий, которые не под тестом ---

// seedUser вставляет пользователя напрямую (минуя HTTP). Пароль не задаётся —
// для сценариев, где вход не нужен (напр. аналитика). Для входа используйте register.
func (s *IntegrationSuite) seedUser(email, name string) int64 {
	s.T().Helper()
	id, err := s.userRepo.Create(s.ctx, &model.User{Email: email, Name: name, PasswordHash: "x"})
	s.Require().NoError(err, "seed user")
	return id
}

// seedTeam создаёт команду и добавляет владельца участником с ролью owner.
func (s *IntegrationSuite) seedTeam(ownerID int64, name string) int64 {
	s.T().Helper()
	id, err := s.teamRepo.Create(s.ctx, &model.Team{Name: name, CreatedBy: ownerID})
	s.Require().NoError(err, "seed team")
	s.seedMember(id, ownerID, model.RoleOwner)
	return id
}

func (s *IntegrationSuite) seedMember(teamID, userID int64, role model.TeamRole) {
	s.T().Helper()
	err := s.teamRepo.AddMember(s.ctx, &model.TeamMember{TeamID: teamID, UserID: userID, Role: role})
	s.Require().NoError(err, "seed member")
}

// execSQL выполняет произвольный SQL на тестовой БД (для сидов с явными таймстемпами).
func (s *IntegrationSuite) execSQL(raw string, args ...any) {
	s.T().Helper()
	_, err := s.db.DB().ExecContext(s.ctx, db.Query{Name: "integration.seed", QueryRaw: raw}, args...)
	s.Require().NoErrorf(err, "exec sql %q", raw)
}

// taskCacheExists сообщает, есть ли в Redis ключ кэша списка задач команды.
func (s *IntegrationSuite) taskCacheExists(teamID int64) bool {
	s.T().Helper()
	n, err := s.redis.Exists(s.ctx, taskCacheKeyPrefix+strconv.FormatInt(teamID, 10)).Result()
	s.Require().NoError(err, "redis exists")
	return n == 1
}
