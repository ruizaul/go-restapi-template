.PHONY: help run build test lint clean swagger dev

help: ## Show available commands
	@echo 'Available commands:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

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

lint: ## Check code quality
	@golangci-lint run

clean: ## Clean build artifacts
	@rm -rf bin/ tmp/
