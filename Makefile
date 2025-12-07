.PHONY: help deps build run docker-run

help:
	@echo "Available commands:"
	@echo "  make deps         - Install dependencies"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make migrate-up   - Run database migrations"
	@echo "  make migrate-down - Rollback migrations"

deps:
	go mod download
	go mod tidy

build:
	go build -o bin/ ./cmd/server

run:
	go run ./cmd/server/main.go

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

docker-run:
	docker-compose up -d

clean:
	rm -rf bin/*
