

run-c:
	go run cmd/client/
run-s:
	go run cmd/server/main.go

gen-proto:
	protoc \
  --go_out=internal/shared/pb --go_opt=paths=source_relative \
  --go-grpc_out=internal/shared/pb --go-grpc_opt=paths=source_relative \
  --go_opt=default_api_level=API_OPAQUE \
  -I proto \
  gophkeeper.proto
