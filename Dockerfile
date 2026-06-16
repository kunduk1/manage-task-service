# --- build stage -------------------------------------------------------------
FROM golang:1.25-alpine AS build

WORKDIR /app

# Кэшируем загрузку зависимостей отдельным слоем.
COPY go.mod go.sum ./
RUN go mod download

# Остальной исходный код (migrations/ и docs/ обязательны — они go:embed'ятся в бинарь).
COPY . .

# Статичная сборка без CGO: пригодна для минимального alpine-образа.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/manage-task-service ./cmd

# --- runtime stage -----------------------------------------------------------
FROM alpine:3.21

# CA-сертификаты для исходящего TLS (например, managed-БД). busybox wget уже есть для healthcheck.
RUN apk add --no-cache ca-certificates

# Непривилегированный пользователь.
RUN adduser -D -u 10001 app

WORKDIR /app

# Логи пишутся в stdout и в файл (lumberjack, LOG_PATH по умолчанию logs/app.log).
# Каталог должен быть доступен пользователю app на запись.
RUN mkdir -p /app/logs && chown app:app /app/logs

COPY --from=build /out/manage-task-service /app/manage-task-service

USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health || exit 1

# Конфиг читается из переменных окружения; флаг -config-path не нужен
# (local.env в образе отсутствует — config.Load пропускает его через os.Stat).
ENTRYPOINT ["/app/manage-task-service"]
