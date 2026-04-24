.PHONY: build run test migrate tidy lint

build:
	go build ./cmd/api/

run:
	go run ./cmd/api/

test:
	go test ./...

migrate:
	psql $(DATABASE_URL) -f migrations/001_initial_schema.sql

tidy:
	go mod tidy

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not found, skipping"
