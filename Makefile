.PHONY: help run build test test-coverage lint clean swagger dev migrate-up migrate-down migrate-create migrate-status migrate-force

# Database connection string for migrations
# Port 5433 to avoid conflict with local PostgreSQL (Docker maps 5433->5432)
DATABASE_URL ?= postgres://postgres:postgres@localhost:5433/app_db?sslmode=disable

help: ## Show available commands
	@echo 'Available commands:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# =============================================================================
# Development
# =============================================================================

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/server/main.go -o docs --quiet
	@echo "Converting to OpenAPI 3.0..."
	@swagger2openapi docs/swagger.json -o docs/openapi.json 2>/dev/null

dev: ## Run server with hot reload (Air)
	@air -c .air.toml

run: swagger ## Run server (regenerates docs)
	@go run cmd/server/main.go

build: swagger ## Build binary (regenerates docs)
	@mkdir -p bin
	@go build -o bin/server cmd/server/main.go

test: ## Run tests
	@go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "Coverage report generated: coverage.html"

lint: ## Check code quality
	@golangci-lint run

clean: ## Clean build artifacts
	@rm -rf bin/ tmp/

# =============================================================================
# Database Migrations (golang-migrate)
# =============================================================================

migrate-up: ## Run all pending migrations
	@echo "Running migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "Migrations completed!"

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@migrate -path migrations -database "$(DATABASE_URL)" down 1
	@echo "Rollback completed!"

migrate-down-all: ## Rollback ALL migrations (DANGER!)
	@echo "⚠️  Rolling back ALL migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" down -all
	@echo "All migrations rolled back!"

migrate-status: ## Show current migration version
	@migrate -path migrations -database "$(DATABASE_URL)" version

migrate-force: ## Force set migration version (use: make migrate-force VERSION=1)
	@echo "Forcing migration version to $(VERSION)..."
	@migrate -path migrations -database "$(DATABASE_URL)" force $(VERSION)

migrate-create: ## Create new migration (use: make migrate-create NAME=create_users)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=create_users"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	@migrate create -ext sql -dir migrations -seq $(NAME)
	@echo "Created migrations/XXXXXX_$(NAME).up.sql and migrations/XXXXXX_$(NAME).down.sql"
