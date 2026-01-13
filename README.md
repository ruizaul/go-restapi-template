# Go API Template

🚀 Production-ready Go REST API template with standard library, auto-generated docs, hot-reload, and clean architecture.

## ✨ Features

- **Go 1.22+ with net/http** - Modern standard library routing with pattern matching
- **Auto-generated API Docs** - Beautiful interactive documentation with [Scalar](https://github.com/scalar/scalar)
- **Hot Reload** - Live reload with [Air](https://github.com/air-verse/air) for fast development
- **PostgreSQL Ready** - Database connection setup with connection pooling
- **Database Migrations** - Version-controlled schema changes with [golang-migrate](https://github.com/golang-migrate/migrate)
- **Clean Architecture** - Layered structure: handlers → services → repositories
- **Linting** - Pre-configured with [golangci-lint](https://golangci-lint.run/)
- **JSend Response Format** - Consistent API responses
- **Docker Compose** - PostgreSQL setup out of the box

### Production-Ready Features

- **Graceful Shutdown** - Proper handling of SIGTERM/SIGINT signals
- **Structured Logging** - JSON logging with `log/slog`
- **Request ID Tracing** - Unique ID for each request (X-Request-ID header)
- **CORS Middleware** - Configurable cross-origin resource sharing
- **Rate Limiting** - In-memory rate limiter (token bucket algorithm)
- **Recovery Middleware** - Panic recovery to prevent server crashes
- **Health Checks** - Kubernetes-ready liveness and readiness probes
- **Centralized Configuration** - Environment-based configuration management
- **Comprehensive Tests** - Unit tests for all packages

## 🎯 Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make
- [swag](https://github.com/swaggo/swag), [air](https://github.com/air-verse/air), [golangci-lint](https://golangci-lint.run/), and [migrate](https://github.com/golang-migrate/migrate)

### Install Tools

```bash
# macOS
brew install swag air golangci-lint golang-migrate

# Go install (alternative)
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/air-verse/air@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Setup

```bash
# 1. Clone the repository
git clone <repo-url>
cd go-restapi-template

# 2. Start PostgreSQL
docker-compose up -d

# 3. Copy environment file
cp .env.example .env

# 4. Run database migrations
make migrate-up

# 5. Start development server
make dev
```

### Verify Installation

- **Health check:** http://localhost:8080/health
- **Liveness probe:** http://localhost:8080/health/live
- **Readiness probe:** http://localhost:8080/health/ready
- **Example endpoint:** http://localhost:8080/test/hello
- **API Documentation:** http://localhost:8080/docs

## 🛠️ Development Commands

| Command | Description |
|---------|-------------|
| `make dev` | Run server with hot reload (recommended) |
| `make run` | Run server once |
| `make build` | Build binary to `bin/server` |
| `make swagger` | Regenerate API documentation |
| `make test` | Run tests |
| `make test-coverage` | Run tests with coverage report |
| `make lint` | Check code quality |
| `make clean` | Clean build artifacts |

## 🗄️ Database Migrations

We use [golang-migrate](https://github.com/golang-migrate/migrate) for database schema management.

### Migration Commands

| Command | Description |
|---------|-------------|
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Rollback the last migration |
| `make migrate-down-all` | Rollback ALL migrations (⚠️ destructive) |
| `make migrate-status` | Show current migration version |
| `make migrate-create NAME=xxx` | Create new migration files |
| `make migrate-force VERSION=n` | Force set version (fix dirty state) |

### Creating a New Migration

```bash
# Create migration files
make migrate-create NAME=create_orders

# This creates:
# migrations/000002_create_orders.up.sql   <- Apply changes
# migrations/000002_create_orders.down.sql <- Rollback changes
```

### Custom Database URL

```bash
# Override default connection
DATABASE_URL="postgres://user:pass@host:5433/mydb?sslmode=disable" make migrate-up
```

## 📁 Project Structure

```
├── cmd/server/          # Application entrypoint
├── internal/            # Private application code (features)
│   └── feature/         # Feature modules
│       ├── handlers/    # HTTP handlers
│       ├── services/    # Business logic
│       ├── models/      # Data models
│       └── routes.go    # Route registration
├── pkg/                 # Public reusable libraries
│   ├── config/          # Centralized configuration
│   ├── middleware/      # HTTP middleware (CORS, logging, recovery, rate limit)
│   └── response/        # JSend response helpers
├── database/            # Database connection setup
├── migrations/          # SQL migrations (golang-migrate)
└── docs/                # Generated API documentation (don't edit)
```

## ⚙️ Configuration

Configuration is loaded from environment variables. See all options:

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `APP_ENV` | `development` | Environment (development/production) |
| `SERVER_READ_TIMEOUT` | `15s` | Read timeout |
| `SERVER_WRITE_TIMEOUT` | `15s` | Write timeout |
| `SERVER_SHUTDOWN_TIMEOUT` | `30s` | Graceful shutdown timeout |

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | - | Full connection string (overrides individual vars) |
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5433` | Database port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `app_db` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |

### Logging Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `LOG_FORMAT` | `json` | Log format (json/text) |

### CORS Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated origins |
| `CORS_ALLOW_CREDENTIALS` | `false` | Allow credentials |

### Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |
| `RATE_LIMIT_RATE` | `100` | Requests per window |
| `RATE_LIMIT_WINDOW` | `1m` | Time window |

## 📋 Code Standards

- **JSend Response Format** - All endpoints return `{status, data}` or `{status, message}`
- **REST Naming** - Use nouns (`/users`), not verbs (`/getUsers`)
- **Swagger Annotations** - Document all endpoints with `@Summary`, `@Tags`, `@Router`, etc.
- **Linting** - Run `make lint` before every commit (zero tolerance for errors)
- **Testing** - Run `make test` to ensure all tests pass

For detailed development guidelines, see [CLAUDE.md](CLAUDE.md).

## 🐳 Docker

Default PostgreSQL configuration in `docker-compose.yml`:
- **Host:** localhost:5433 (mapped to container port 5432)
- **Database:** app_db
- **User/Password:** postgres/postgres

> **Note:** Port 5433 is used to avoid conflicts with local PostgreSQL installations.

### Production Docker Build

```bash
# Build image
docker build -t my-api .

# Run container
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@host:5432/db?sslmode=require" \
  my-api
```

## 🔍 Health Checks

The API provides three health check endpoints:

| Endpoint | Purpose | Checks |
|----------|---------|--------|
| `/health` | Full health status | Server + Database |
| `/health/live` | Liveness probe | Server is running |
| `/health/ready` | Readiness probe | Ready to accept traffic (DB connected) |

Example response from `/health`:

```json
{
  "status": "success",
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "database": {
      "status": "healthy"
    }
  }
}
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test -v ./pkg/middleware/...

# Run with race detection
go test -race ./...
```

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

---

Made with ❤️ for the Go community