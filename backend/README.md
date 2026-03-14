# epn-killer-mvp

Бэкенд MVP сервиса виртуальных карт EPN Killer: API (Gorilla Mux), JWT, X-API-Key, PostgreSQL, карты, команды, рефералы, отчёты, Telegram-уведомления.

## Структура (по правилам репозитория)

- **Корень:** `go.mod`, `go.sum`, `Makefile`, `README.md`, `schema.sql`
- **cmd/** — точка входа: `main.go`
- **internal/** — весь остальной код: config, usecases, api, middleware, models, notification, repository, telegram, utils

## Требования

- Go 1.20+
- PostgreSQL (переменная `DATABASE_URL`)

## Сборка и запуск

```bash
make build
# или
go build -o bin/server ./cmd
``` 

Запуск (нужна переменная `DATABASE_URL`):

```bash
export DATABASE_URL="postgresql://user:pass@localhost:5432/db?sslmode=disable"
./bin/server
```

Порт по умолчанию: `8080` (переменная `PORT`).

## API

- **Health:** `GET /health`
- **Auth:** `POST /api/v1/auth/register`, `POST /api/v1/auth/login`
- **User:** `GET /api/v1/user/me`, `POST /api/v1/user/deposit`, карты, отчёт, API-ключ, грейд, команды, рефералы, настройки Telegram

Подробная документация — в репозитории **epn-killer-docs**.
