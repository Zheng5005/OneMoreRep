.PHONY: dev build test migrate-up migrate-down seed sqlc

# Development server with hot reload
dev:
	@air -c .air.toml

# Build production binary
build:
	@CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/api ./cmd/api

# Run all tests
test:
	@go test -v -race ./...

# Run database migrations up
migrate-up:
	@migrate -path migrations -database "${DATABASE_URL}" up

# Run database migrations down
migrate-down:
	@migrate -path migrations -database "${DATABASE_URL}" down 1

# Seed sample data
seed:
	@go run ./cmd/seed

# Generate Go code from SQL queries using sqlc
sqlc:
	@sqlc generate
