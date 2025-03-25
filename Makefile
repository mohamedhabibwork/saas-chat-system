.PHONY: build run test lint clean swagger

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Generate Swagger documentation
swagger:
	swag init -g cmd/server/main.go

# Run database migrations
migrate:
	go run cmd/migrate/main.go

# Run database migrations down
migrate-down:
	go run cmd/migrate/main.go down

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run all checks (lint, test, build)
check: lint test build

# Run the application with hot reload
dev:
	air 