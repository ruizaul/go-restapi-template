.PHONY: help dev build test lint swagger db-up db-down db-shell migrate-up migrate-down migrate-new deploy

help: ## Mostrar comandos disponibles
	@echo '🚀 Comandos disponibles:'
	@echo ''
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ============================================================================
# 🔧 DESARROLLO LOCAL
# ============================================================================

run: ## Iniciar servidor con hot-reload
	@~/go/bin/air

build: ## Compilar binario
	@mkdir -p bin
	@go build -o bin/server cmd/server/main.go
	@echo "✅ Compilado: bin/server"

test: ## Ejecutar tests
	@go test -v ./...

lint: ## Revisar código con linter
	@golangci-lint run

swagger: ## Regenerar documentación de API
	@~/go/bin/swag init -g cmd/server/main.go -o docs
	@echo "✅ Docs generados → /docs"

# ============================================================================
# 🐳 BASE DE DATOS (Docker)
# ============================================================================

db-up: ## Iniciar PostgreSQL en Docker
	@docker-compose up -d postgres
	@echo "⏳ Esperando PostgreSQL..."
	@sleep 5
	@echo "✅ PostgreSQL listo → localhost:5432"

db-down: ## Detener PostgreSQL
	@docker-compose down

db-shell: ## Conectar a PostgreSQL
	@docker-compose exec postgres psql -U postgres -d tacoshare_delivery

db-logs: ## Ver logs de PostgreSQL
	@docker-compose logs -f postgres

# ============================================================================
# 🗄️ MIGRACIONES
# ============================================================================

migrate-up: ## Aplicar migraciones pendientes
	@export $$(grep DATABASE_URL .env | xargs) && migrate -path migrations -database "$$DATABASE_URL" up
	@echo "✅ Migraciones aplicadas"

migrate-down: ## Revertir última migración
	@export $$(grep DATABASE_URL .env | xargs) && migrate -path migrations -database "$$DATABASE_URL" down 1
	@echo "✅ Migración revertida"

migrate-new: ## Crear nueva migración (uso: make migrate-new name=add_users)
	@if [ -z "$(name)" ]; then \
		echo "❌ Uso: make migrate-new name=nombre_migracion"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir migrations -seq $(name)
	@echo "✅ Migración creada en migrations/"

# ============================================================================
# 🚀 DEPLOY A PRODUCCIÓN
# ============================================================================

deploy: ## Deployar a Cloud Run
	@echo "🚀 Deploying a Google Cloud..."
	@read -p "¿Continuar? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		gcloud builds submit --region=us-west2 --project=delivery-93190 --config=cloudbuild.yaml; \
	else \
		echo "❌ Cancelado"; \
	fi
