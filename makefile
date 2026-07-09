SERVER_ADDR = :8080
GRPC_ADDR   = :3200
DB_DSN      = postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable
SECRET_KEY = super_secret_goph_keeper_key
CLIENT_DB_PATH = ./internal/client/db/client.db
CLIENT_MIGRATIONS_DIR = ./internal/client/migration

run-dev:
	go run ./cmd/server/main.go -a "$(SERVER_ADDR)" -g "$(GRPC_ADDR)" -d "$(DB_DSN)" -k "$(SECRET_KEY)"

db-up:
	docker compose up -d  

db-down:
	docker compose down

migrate-up-s:
	migrate -path ./internal/server/migrations -database "$(DB_DSN)" up	


migration:
ifndef name
	$(error var name is undefined! Use: make migration name=migraion_name)
endif
	migrate create -ext sql -dir ./internal/server/migrations -seq $(name)

gen-proto:
	protoc \
  --go_out=internal/shared/pb --go_opt=paths=source_relative \
  --go-grpc_out=internal/shared/pb --go-grpc_opt=paths=source_relative \
  --go_opt=default_api_level=API_OPAQUE \
  -I proto \
  gophkeeper.proto

gen-query:
	sqlc generate


# CLIENT COMMAND
login:
	go run ./cmd/client/main.go login  -a "$(SERVER_ADDR)" -g "$(GRPC_ADDR)"

register:
	go run ./cmd/client/main.go register  -a "$(SERVER_ADDR)" -g "$(GRPC_ADDR)"

migrate-up-c:
	migrate -path ./internal/client/migrations -database "sqlite3://$(CLIENT_DB_PATH)" up

migrate-create-c:
	migrate create -ext sql -dir $(CLIENT_MIGRATIONS_DIR) -seq $(NAME)