include .env

.PHONY: help build run run-race clean lint generate format test bin-deps \
        docker-up docker-down docker-prod-up \
        migrate-create migrate-up migrate-down migrate-force

help: ## Показать все команды
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Собрать бинарник
	go build -o bin/server ./cmd/main.go

run: ## Запустить локально
	go run ./cmd/main.go

run-race: ## Запустить с race detector
	go run --race ./cmd/main.go

clean: ## Очистить
	rm -rf bin/

lint: ## Запустить линтер
	golangci-lint run ./... -v

generate: ## go generate
	go generate ./...

format: ## Форматировать код
	gofmt -s -w .
	goimports -w -d .

test: ## Тесты
	go test ./internal/... -race -count=1

bin-deps: ## Установить goose и другие инструменты
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# === DOCKER ===
docker-up: ## Запустить dev-окружение (основной compose в корне)
	docker compose -f ./docker-compose.yaml --env-file .env up --build

docker-down: ## Остановить dev
	docker compose -f ./docker-compose.yaml down

docker-prod-up: ## Запустить прод-версию (из deployment/)
	docker compose -f ./deployment/docker-compose.prod.yaml --env-file ./deployment/.env up --build

# === MIGRATIONS ===
migrate-create: ## Создать миграцию: make migrate-create name=add_wallet_table
	goose postgres "$(POSTGRES_DSN)" create "$(filter-out $@,$(MAKECMDGOALS))" sql -dir ./migrations/postgres

migrate-up: ## Применить все миграции
	goose postgres "$(POSTGRES_DSN)" -dir ./migrations/postgres up

migrate-down: ## Откатить одну миграцию
	goose postgres "$(POSTGRES_DSN)" -dir ./migrations/postgres down

migrate-force: ## Откатить до конкретной версии: make migrate-force version=20260210
	goose postgres "$(POSTGRES_DSN)" -dir ./migrations/postgres down-to "$(filter-out $@,$(MAKECMDGOALS))"

%:
	@true