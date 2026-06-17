.PHONY: run build tidy test test-integration cover vet lint lint-fix swagger mocks generate bin-deps compose-up compose-down migrate-up docker-build

APP_NAME                  := manage-task-service
CONFIG                    := local.env
GOLANGCI_LINT_VERSION     := v2.12.2
GO_TEST_COVERAGE_VERSION  := v2.18.3

# Инструменты разработки ставятся в ./bin (per-project, без глобального $GOPATH/bin).
# Версия зашита в имя файла — при bump переменной make переустановит автоматически.
LOCAL_BIN                 := $(CURDIR)/bin
GOLANGCI_LINT             := $(LOCAL_BIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
GO_TEST_COVERAGE          := $(LOCAL_BIN)/go-test-coverage-$(GO_TEST_COVERAGE_VERSION)

COVER_FLAGS ?=

run:
	go run ./cmd -config-path $(CONFIG)

build:
	go build -o bin/$(APP_NAME) ./cmd

# Собирает Docker-образ сервиса (multi-stage, см. Dockerfile).
docker-build:
	docker build -t $(APP_NAME):latest .

tidy:
	go mod tidy

test:
	go test ./...

# Интеграционные тесты на testcontainers (нужен запущенный Docker). Поднимают
# MySQL+Redis в контейнерах, гоняют реальный роутер end-to-end. Тег integration
# исключает их из обычного `make test` и из покрытия (cover).
test-integration:
	go test -tags=integration -race -count=1 ./test/integration/...

# Прогоняет тесты с профилем покрытия и проверяет порог 85% (go-test-coverage).
cover: $(GO_TEST_COVERAGE)
	go test -coverprofile=coverage.out ./...
	$(GO_TEST_COVERAGE) --config=.testcoverage.yml $(COVER_FLAGS)

vet:
	go vet ./...

# Запускает golangci-lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

# То же, но с авто-исправлением (gofmt/goimports и прочие фиксы).
lint-fix: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix

# Ставит dev-инструменты в ./bin (idempotent — пропустит уже установленные нужной версии).
bin-deps: $(GOLANGCI_LINT) $(GO_TEST_COVERAGE)

# go install кладёт бинарь под именем пакета — переименовываем в версионированное имя.
$(GOLANGCI_LINT):
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@mv $(LOCAL_BIN)/golangci-lint $@

$(GO_TEST_COVERAGE):
	GOBIN=$(LOCAL_BIN) go install github.com/vladopajic/go-test-coverage/v2@$(GO_TEST_COVERAGE_VERSION)
	@mv $(LOCAL_BIN)/go-test-coverage $@

# Генерирует OpenAPI 3.1 спецификацию из аннотаций в docs/ (swag v2).
# swagfix чинит баг swag v2 (#2086): request body заворачивается в кривой oneOf
# с пустым {type: object} — из-за этого Postman подставляет пустое тело запроса.
swagger:
	go tool swag init -g cmd/main.go -o docs --v3.1 --parseDependency --parseInternal
	go run ./tools/swagfix docs

# Генерирует gomock-моки для интерфейсов
# Ограничено ./internal/..., чтобы не триггерить swag-директиву из cmd/main.go.
mocks:
	go generate ./internal/...

# Алиас: на текущий момент = только моки (swag живёт под целью swagger).
generate: mocks

compose-up:
	docker compose up -d

compose-down:
	docker compose down

# Применяет SQL-миграцию к MySQL из docker-compose (контейнер должен быть запущен: make compose-up).
# Ручной вариант без версионирования; приложение само поднимает миграции на старте (golang-migrate, см. internal/migrator).
migrate-up:
	docker compose exec -T mysql mysql -uapp -papp manage_task < migrations/0001_create_users.up.sql
