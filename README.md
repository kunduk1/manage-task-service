# manage-task-service

HTTP-сервис на Go для управления задачами внутри команд. Поддерживает регистрацию и
вход пользователей, создание команд с приглашением участников и работу с задачами,
включая историю изменений. Источник правды — MySQL, Redis используется для
refresh-токенов и счётчиков rate-limit.

## Возможности

- **Аутентификация** по JWT: access-токен (HS256) + opaque refresh-токен.
- **Команды**: создание, список своих команд, приглашение участников с ролями,
  статистика по командам и топ создателей задач за период.
- **Задачи**: создание, список с фильтрами и пагинацией, частичное обновление с
  автоматической записью истории изменений, просмотр истории, проверка задач,
  назначенных на не-участников команды.
- **Надёжность и наблюдаемость**: per-user rate limiting (Redis),
  метрики Prometheus, circuit breaker для отправки email-приглашений.

## Технологии

Go 1.25 · chi (роутер) · MySQL (основное хранилище) · Redis (токены + rate-limit) ·
JWT (HS256) · Prometheus · Swagger / OpenAPI 3.1.

## Быстрый старт

### Через Docker Compose (проще всего)

Поднимает сервис вместе с MySQL и Redis:

```bash
make compose-up      # запустить app + MySQL + Redis
make compose-down    # остановить
```

После старта доступны:
- API — http://localhost:8080
- Метрики — http://localhost:9090/metrics
- Swagger UI — http://localhost:8080/swagger/

### Локальный запуск

Нужны запущенные MySQL и Redis (можно поднять через `make compose-up` только их).

```bash
cp local.env.example local.env   # при необходимости поправить значения
make run                         # go run ./cmd -config-path local.env
```

Миграции базы данных применяются автоматически при старте приложения —
вручную ничего накатывать не нужно.

### Тестовые данные (seed)

Чтобы наполнить локальную БД готовым набором данных (для ручного тестирования,
Swagger, фронтенда или демо):

```bash
make compose-up   # поднять MySQL + Redis
make seed         # наполнить БД фикстурами (флаг -reset стирает прежние данные)
```

`make seed` создаёт 5 пользователей, 2 команды (с ролями owner/admin/member),
8 задач (все статусы, в т.ч. misassigned-задача и разброс дат для аналитики),
историю изменений и комментарии. У всех аккаунтов общий пароль `password123`:

| Email               | Имя          | Роль(и)                                  |
| ------------------- | ------------ | ---------------------------------------- |
| `alice@example.com` | Alice Owner  | owner команды Platform                   |
| `bob@example.com`   | Bob Admin    | admin команды Platform                   |
| `carol@example.com` | Carol Member | member команды Platform                  |
| `dave@example.com`  | Dave Lead    | owner команды Mobile                     |
| `erin@example.com`  | Erin Member  | member команды Mobile                    |

```bash
# залогиниться засеянным пользователем
curl -s localhost:8080/api/v1/login \
  -d '{"email":"alice@example.com","password":"password123"}'
```

Реализация — `cmd/seed` + пакет `internal/seed` (детерминированные фикстуры).

## Конфигурация

Настройки читаются из env-файла (флаг `-config-path`, по умолчанию `local.env`).
Полный список переменных со значениями по умолчанию — в `local.env.example`.
Обязательна только `JWT_SECRET`, у всех остальных есть значения по умолчанию.

Ключевые переменные:

| Переменная           | По умолчанию          | Назначение                              |
| -------------------- | --------------------- | --------------------------------------- |
| `HTTP_PORT`          | `8080`                | Порт HTTP API                           |
| `METRICS_PORT`       | `9090`                | Порт сервера метрик Prometheus          |
| `MYSQL_HOST`/`_PORT` | `localhost`/`3306`    | Подключение к MySQL                      |
| `MYSQL_USER`/`_PASSWORD`/`_DATABASE` | `app`/`app`/`manage_task` | Учётные данные MySQL        |
| `REDIS_ADDR`         | `localhost:6379`      | Адрес Redis                             |
| `JWT_SECRET`         | — (обязательна)       | Секрет для подписи access-токенов        |
| `RATE_LIMIT_ENABLED` | `true`                | Включить per-user rate limiting          |
| `RATE_LIMIT_REQUESTS`/`_WINDOW` | `100`/`1m` | Лимит запросов на пользователя в окно    |

## API

Базовый путь — `/api/v1`. Защищённые ручки требуют заголовок
`Authorization: Bearer <access_token>` (токен выдаёт `POST /login`).

| Метод | Путь                       | Описание                                            | Auth |
| ----- | -------------------------- | --------------------------------------------------- | ---- |
| POST  | `/register`                | Регистрация пользователя                            | —    |
| POST  | `/login`                   | Вход, выдаёт access + refresh токены                | —    |
| POST  | `/teams`                   | Создать команду                                     | ✅   |
| GET   | `/teams`                   | Список команд пользователя                          | ✅   |
| GET   | `/teams/stats`             | Статистика по командам (участники, задачи за 7 дней)| ✅   |
| GET   | `/teams/top-creators`      | Топ создателей задач по командам за период           | ✅   |
| POST  | `/teams/{id}/invite`       | Пригласить участника (только owner/admin)           | ✅   |
| POST  | `/tasks`                   | Создать задачу                                       | ✅   |
| GET   | `/tasks`                   | Список задач с фильтрами и пагинацией               | ✅   |
| GET   | `/tasks/misassigned`       | Задачи, назначенные на не-участников команды         | ✅   |
| PUT   | `/tasks/{id}`              | Обновить задачу (с записью истории изменений)       | ✅   |
| GET   | `/tasks/{id}/history`      | История изменений задачи                            | ✅   |

Дополнительно: `GET /health` — проверка живости, `GET /swagger/` — интерактивная
документация (спецификация — `GET /swagger/doc.json`).

## Структура проекта

```
cmd/            точка входа (main.go)
internal/
  api/          HTTP-обработчики (auth, team, task) + router.go
  app/          инициализация приложения и DI-контейнер
  service/      бизнес-логика (auth, team, task, authz)
  repository/   доступ к данным (user, team, task, ...)
  clients/      клиенты внешних систем (MySQL, Redis, email)
  config/       загрузка конфигурации по секциям
  transport/    middleware, утилиты запросов/ответов
  ...           token, metrics, ratelimit, circuitbreaker, migrator и др.
pkg/            публичные контракты: DTO (<domain>/v1) и доменные ошибки
migrations/     SQL-миграции (применяются автоматически на старте)
test/integration/  end-to-end тесты на testcontainers
docs/           сгенерированная OpenAPI-спецификация
```

## Команды разработки

| Команда                 | Что делает                                              |
| ----------------------- | ------------------------------------------------------ |
| `make run`              | Запустить сервис локально                              |
| `make seed`             | Наполнить dev-БД фикстурами (`cmd/seed`, с `-reset`)   |
| `make build`            | Собрать бинарь в `bin/manage-task-service`             |
| `make test`             | Юнит-тесты (`go test ./...`)                           |
| `make test-integration` | Интеграционные тесты (testcontainers, нужен Docker)    |
| `make lint`             | Запустить golangci-lint                                |
| `make cover`            | Тесты с покрытием и проверкой порога 85%               |
| `make swagger`          | Перегенерировать OpenAPI-спецификацию                  |
| `make mocks`            | Перегенерировать gomock-моки                           |
| `make compose-up` / `make compose-down` | Поднять / остановить MySQL + Redis + app |

## Тесты

- `make test` — быстрые юнит-тесты.
- `make test-integration` — интеграционные тесты, поднимают реальные MySQL и Redis
  в контейнерах (нужен запущенный Docker).
