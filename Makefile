generate:
	@go get github.com/99designs/gqlgen/codegen/config@v0.17.49
	@go get github.com/99designs/gqlgen@v0.17.49
	@go generate ./internal/graphQL/resolver.go

local_run: generate
	@go build ./cmd/server/main.go
	@./main

