generate:
	@go get github.com/99designs/gqlgen/codegen/config@v0.17.49
	@go get github.com/99designs/gqlgen@v0.17.49
	@go generate ./internal/graphQL/resolver.go

local_run_memory_storage: generate
	@go build ./cmd/server/main.go
	@./main -db=mem

local_run_postgres_storage: generate
	@go build ./cmd/server/main.go
	@./main -db=sql

docker_run:
	@docker compose up

docker_stop:
	@docker compose down