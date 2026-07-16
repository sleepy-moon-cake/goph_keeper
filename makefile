# ==============================================================================
# CONFIGURATION & VARIABLES
# ==============================================================================
SERVER_ADDR           ?= :8080
GRPC_ADDR             ?= :3200
DB_DSN                ?= postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable
SECRET_KEY            ?= super_secret_goph_keeper_key
CLIENT_DB_PATH        ?= ./internal/client/db/client.db
CLIENT_MIGRATIONS_DIR ?= ./internal/client/migrations

SERVER_BIN = ./bin/server
CLIENT_BIN = ./bin/gophkeeper

# ==============================================================================
# ENTRY POINT (START)
# ==============================================================================
.PHONY: start stop restart

# Полный автоматический запуск всего бэкенд-окружения
start: env-init db-up
	@echo "Waiting for database to be ready..."
	@sleep 2 # Небольшая пауза, чтобы PostgreSQL успел инициализироваться в Docker
	@echo "Running backend migrations..."
	@$(MAKE) migrate-up-s
	@echo "Starting GophKeeper backend server..."
	@$(MAKE) run-dev

# Полная остановка инфраструктуры
stop: db-down
	@echo "Backend infrastructure stopped."

# Перезапуск бэкенда
restart: stop start

# ==============================================================================
# DEVELOPMENT & ENVIRONMENT
# ==============================================================================
.PHONY: run-dev db-up db-down env-init

run-dev:
	go run ./cmd/server/main.go -a "$(SERVER_ADDR)" -g "$(GRPC_ADDR)" -d "$(DB_DSN)" -k "$(SECRET_KEY)"

db-up:
	docker compose up -d  

db-down:
	docker compose down

env-init:
	@mkdir -p ./internal/client/db
	@mkdir -p ./internal/client/migrations
	@mkdir -p ./internal/server/migrations
	@mkdir -p ./bin

# ==============================================================================
# CODE GENERATION & QUALITY
# ==============================================================================
.PHONY: gen-proto gen-query lint test test-cover

gen-proto:
	protoc \
		--go_out=internal/shared/pb --go_opt=paths=source_relative \
		--go-grpc_out=internal/shared/pb --go-grpc_opt=paths=source_relative \
		--go_opt=default_api_level=API_OPAQUE \
		-I proto \
		gophkeeper.proto

gen-query:
	@rm -rf ~/.cache/sqlc
	sqlc generate

lint:
	golangci-lint run ./...

test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# ==============================================================================
# DATABASE MIGRATIONS
# ==============================================================================
.PHONY: migrate-up-s migration-create-s migration-create-c cache-clean

# Накатить миграции бэкенда (PostgreSQL)
migrate-up-s:
	migrate -path ./internal/server/migrations -database "$(DB_DSN)" up	

# Создать новую миграцию для сервера
migration-create-s: env-init
ifndef name
	$(error var name is undefined! Use: make migration-create-s name=migration_name)
endif
	migrate create -ext sql -dir ./internal/server/migrations -seq $(name)

# Создать новую миграцию для клиента (Она применится автоматически при старте кода)
migration-create-c: env-init
ifndef name
	$(error var name is undefined! Use: make migration-create-c name=migration_name)
endif
	migrate create -ext sql -dir $(CLIENT_MIGRATIONS_DIR) -seq $(name)

# Очистка локального SQLite кэша при сбоях дебага
cache-clean:
	@rm -f $(CLIENT_DB_PATH)
	@echo "Local SQLite database cache has been wiped."

# ==============================================================================
# CLIENT CLI COMMANDS
# ==============================================================================
.PHONY: login register build cli

start-client: build
	@echo "Starting GophKeeper background agent..."
	@go run ./cmd/client/main.go start &
	@echo "Agent is running in background. You can now use 'make cli args=...'"

login:
	go run ./cmd/client/main.go login -a "$(SERVER_ADDR)" -g "$(GRPC_ADDR)"

register:
	go run ./cmd/client/main.go register -a "$(SERVER_ADDR)" -g "$(GRPC_ADDR)"

# Быстрый запуск скомпилированного CLI-клиента (Использование: make cli args="get --name summer")
cli: build
	@$(CLIENT_BIN) $(args)

build: env-init
	go build -o $(SERVER_BIN) ./cmd/server/main.go
	go build -o $(CLIENT_BIN) ./cmd/client/main.go
