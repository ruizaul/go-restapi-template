# Go API Template

🚀 Modern Go REST API template with standard library, auto-generated docs, hot-reload, and clean architecture.

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

### Example Migration

```sql
-- migrations/000001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

```sql
-- migrations/000001_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

### Custom Database URL

```bash
# Override default connection
DATABASE_URL="postgres://user:pass@host:5432/mydb?sslmode=disable" make migrate-up
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
├── database/            # Database connection setup
├── migrations/          # SQL migrations (golang-migrate)
└── docs/                # Generated API documentation (don't edit)
```

## 📋 Code Standards

- **JSend Response Format** - All endpoints return `{status, data}` or `{status, message}`
- **REST Naming** - Use nouns (`/users`), not verbs (`/getUsers`)
- **Swagger Annotations** - Document all endpoints with `@Summary`, `@Tags`, `@Router`, etc.
- **Linting** - Run `make lint` before every commit (zero tolerance for errors)

For detailed development guidelines, see [CLAUDE.md](CLAUDE.md).

## 🐳 Docker

Default PostgreSQL configuration in `docker-compose.yml`:
- **Host:** localhost:5433 (mapped to container port 5432)
- **Database:** app_db
- **User/Password:** postgres/postgres

> **Note:** Port 5433 is used to avoid conflicts with local PostgreSQL installations.

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

---

Made with ❤️ for the Go community