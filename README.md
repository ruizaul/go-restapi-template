# Go API Template

🚀 Modern Go REST API template with standard library, auto-generated docs, hot-reload, and clean architecture.

## ✨ Features

- **Go 1.22+ with net/http** - Modern standard library routing with pattern matching
- **Auto-generated API Docs** - Beautiful interactive documentation with [Scalar](https://github.com/scalar/scalar)
- **Hot Reload** - Live reload with [Air](https://github.com/air-verse/air) for fast development
- **PostgreSQL Ready** - Database connection setup with connection pooling
- **Clean Architecture** - Layered structure: handlers → services → repositories
- **Linting** - Pre-configured with [golangci-lint](https://golangci-lint.run/)
- **JSend Response Format** - Consistent API responses
- **Docker Compose** - PostgreSQL setup out of the box

## 🎯 Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make
- [swag](https://github.com/swaggo/swag), [air](https://github.com/air-verse/air), and [golangci-lint](https://golangci-lint.run/)

### Setup

1. Clone the repository
2. Run `docker-compose up -d` to start PostgreSQL
3. Copy `.env.example` to `.env` and configure
4. Run `make dev` to start the development server

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
├── migrations/          # SQL migrations
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
- **Host:** localhost:5432
- **Database:** app_db
- **User/Password:** postgres/postgres

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

---

Made with ❤️ for the Go community