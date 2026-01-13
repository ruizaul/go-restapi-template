# Go API Template - Development Guide

## What This Is
Modern Go REST API template using standard library `net/http` with Go 1.22+ ServeMux and PostgreSQL.

## Project Structure
```
cmd/server/main.go    # Application entrypoint
database/             # Database connection setup
internal/             # Private application code (features)
  └── feature/        # Feature module example
      ├── handlers/   # HTTP request handlers
      ├── services/   # Business logic layer
      ├── repositories/ # Data access layer (optional)
      ├── models/     # Data structures
      └── routes.go   # Route registration
pkg/                  # Public shared utilities
  ├── config/         # Centralized configuration management
  ├── middleware/     # HTTP middleware (CORS, logging, recovery, rate limit)
  └── response/       # JSend response helpers
migrations/           # SQL database migrations (golang-migrate)
docs/                 # Auto-generated API docs (DO NOT EDIT)
```

## Development Commands

```bash
# Development
make dev      # 🔥 Hot reload with Air (recommended)
make run      # Run server once (regenerates docs)
make build    # Compile binary to bin/server
make swagger  # Regenerate API documentation only
make test     # Run all tests
make test-coverage  # Run tests with coverage report
make lint     # Code quality checks (REQUIRED before commit)
make clean    # Remove build artifacts

# Database Migrations
make migrate-up       # Run all pending migrations
make migrate-down     # Rollback last migration
make migrate-down-all # Rollback ALL migrations (DANGER!)
make migrate-status   # Show current migration version
make migrate-create NAME=create_orders  # Create new migration
make migrate-force VERSION=1            # Force set version (dirty state fix)
```

## Development Workflow

1. Start PostgreSQL: `docker-compose up -d`
2. Run migrations: `make migrate-up`
3. Start development server: `make dev`
4. Edit code - Air watches changes and auto-restarts
5. View docs at http://localhost:8080/docs
6. Before committing:
   - Run `make lint` - fix ALL errors (zero tolerance)
   - Run `make test` - ensure all tests pass

## Database Migrations

We use [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

### Install migrate CLI

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Go install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Migration Commands

| Command | Description |
|---------|-------------|
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Rollback the last migration |
| `make migrate-down-all` | Rollback ALL migrations (⚠️ destructive) |
| `make migrate-status` | Show current migration version |
| `make migrate-create NAME=xxx` | Create new migration files |
| `make migrate-force VERSION=n` | Force set version (fix dirty state) |

### Creating Migrations

```bash
# Create a new migration
make migrate-create NAME=create_orders

# This creates two files:
# migrations/000002_create_orders.up.sql   <- Apply changes
# migrations/000002_create_orders.down.sql <- Rollback changes
```

### Migration Best Practices

1. **Always write both UP and DOWN migrations**
   - UP: Apply the change
   - DOWN: Revert the change completely

2. **Make migrations atomic**
   - One logical change per migration
   - Don't mix table creation with data migration

3. **Test rollbacks locally**
   ```bash
   make migrate-up      # Apply
   make migrate-down    # Rollback
   make migrate-up      # Re-apply (should work!)
   ```

4. **Use IF EXISTS / IF NOT EXISTS**
   ```sql
   -- UP
   CREATE TABLE IF NOT EXISTS users (...);
   CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
   
   -- DOWN
   DROP INDEX IF EXISTS idx_users_email;
   DROP TABLE IF EXISTS users;
   ```

5. **Never modify applied migrations**
   - Create a new migration instead
   - Exception: Only during development before merging

### Example Migration

```sql
-- migrations/000001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

```sql
-- migrations/000001_create_users_table.down.sql
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

### Fixing Dirty Database State

If a migration fails halfway, the database enters a "dirty" state:

```bash
# Check current state
make migrate-status
# Output: 2 (dirty)

# Force set to last successful version
make migrate-force VERSION=1

# Fix your migration SQL, then retry
make migrate-up
```

### Custom Database URL

Override the default connection string:

```bash
# Use different database
DATABASE_URL="postgres://user:pass@host:5433/mydb?sslmode=disable" make migrate-up

# Or export it
export DATABASE_URL="postgres://user:pass@host:5433/mydb?sslmode=disable"
make migrate-up
```

> **Note:** Default port is 5433 (not 5432) to avoid conflicts with local PostgreSQL installations.

## API Documentation

### Access Documentation
- **URL**: http://localhost:8080/docs
- **Engine**: Scalar UI (modern, interactive)
- **Format**: OpenAPI 3.0 spec
- **Generator**: swaggo/swag from code annotations

### Document Your Endpoints

Add Swagger annotations above handler functions:

```go
// GetUser godoc
// @Summary      Get user by ID
// @Description  Retrieve user information by unique identifier
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  models.UserResponse  "User found"
// @Failure      404  {object}  models.FailResponse  "User not found"
// @Failure      500  {object}  models.ErrorResponse "Server error"
// @Router       /users/{id} [get]
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

### Response Models (REQUIRED for Scalar UI)

For response bodies to appear in Scalar docs, you MUST define typed models in `models/`:

```go
// models/user_model.go

// UserData contains user information
type UserData struct {
    ID    string `json:"id" example:"123"`
    Name  string `json:"name" example:"John Doe"`
    Email string `json:"email" example:"john@example.com"`
}

// UserResponse - success response
type UserResponse struct {
    Status string   `json:"status" example:"success"`
    Data   UserData `json:"data"`
}

// FailResponse - client error (4xx)
type FailResponse struct {
    Status string            `json:"status" example:"fail"`
    Data   map[string]string `json:"data"`
}

// ErrorResponse - server error (5xx)
type ErrorResponse struct {
    Status  string `json:"status" example:"error"`
    Message string `json:"message" example:"Internal server error"`
    Code    int    `json:"code" example:"500"`
}
```

**Key rules:**
- Use `{object} models.YourType` in annotations, NOT `map[string]interface{}`
- Add `example:"value"` tags to show sample values in docs
- Define separate response types for success/fail/error

### Main API Metadata (cmd/server/main.go)

```go
// @title           Your API Title
// @version         1.0.0
// @description     API description here
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Type "Bearer" followed by JWT token
```

## Code Standards

### JSend Response Format (MANDATORY)

All endpoints MUST use JSend format:

**Success:**
```json
{
  "status": "success",
  "data": { "user": { "id": "123" } }
}
```

**Fail (client error - 4xx):**
```json
{
  "status": "fail",
  "data": { "email": "Invalid email format" }
}
```

**Error (server error - 5xx):**
```json
{
  "status": "error",
  "message": "Database connection failed",
  "code": 500
}
```

### REST API Naming Conventions

✅ **DO:**
- Use nouns: `/users`, `/orders`, `/products`
- Plural for collections: `/users` not `/user`
- Use HTTP methods: `GET /users`, `POST /users`, `DELETE /users/{id}`
- Nested resources: `/users/{id}/orders`
- Query params for filters: `/users?status=active&limit=10`

❌ **DON'T:**
- Use verbs in URLs: `/getUsers`, `/createOrder`
- Mix singular/plural: `/user`, `/orders`
- Custom actions as resources: `/users/search` (use `GET /users?q=search` instead)

### Feature Architecture

Each feature follows layered architecture:

```
internal/users/
├── handlers/          # HTTP layer (request/response handling)
│   └── user_handler.go
├── services/          # Business logic layer
│   └── user_service.go
├── repositories/      # Data access layer (if needed)
│   └── user_repository.go
├── models/           # Data structures
│   └── user.go
└── routes.go         # Route registration
```

### Struct Field Ordering

Order fields for optimal memory alignment:

```go
type User struct {
    ID        uuid.UUID  // 1. UUIDs first
    DeletedAt *time.Time // 2. Pointers
    CreatedAt time.Time  // 3. time.Time
    UpdatedAt time.Time
    Email     string     // 4. Strings
    Name      string
    Age       int        // 5. Integers
    Score     int64
    IsActive  bool       // 6. Booleans last
}
```

## Critical Rules

### 1. Linting - Zero Tolerance
- Run `make lint` after every change
- Fix ALL errors before committing
- No exceptions

### 2. Layer Separation
- **Handlers**: HTTP request/response only
- **Services**: Business logic, no HTTP knowledge
- **Repositories**: Database operations only
- **Models**: Data structures, no logic

### 3. Error Handling
```go
// Good: Return errors, let caller handle
func (s *Service) GetUser(id string) (*User, error) {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, err
    }
    return user, nil
}

// Handler converts to JSend
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.service.GetUser(id)
    if err != nil {
        // Return JSend fail/error response
    }
    // Return JSend success response
}
```

### 4. Never Edit Generated Files
- `docs/` folder is auto-generated
- Modify source code, run `make swagger`

### 5. Production-Ready Server
Always configure timeouts:

```go
server := &http.Server{
    Addr:              ":8080",
    Handler:           mux,
    ReadTimeout:       15 * time.Second,
    WriteTimeout:      15 * time.Second,
    IdleTimeout:       60 * time.Second,
    ReadHeaderTimeout: 5 * time.Second,
}
```

### 6. Database Connections
- Use connection pooling
- Set max open/idle connections
- Configure connection lifetime
- Handle connection errors gracefully

### 7. Migrations
- Always run `make migrate-up` after pulling changes
- Test `make migrate-down` before pushing new migrations
- Never modify migrations that are already in production

## Testing Guidelines

```go
// Unit tests: *_test.go files alongside code
func TestUserService_GetUser(t *testing.T) {
    // Arrange
    service := NewUserService()
    
    // Act
    user, err := service.GetUser("123")
    
    // Assert
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    if user.ID != "123" {
        t.Errorf("expected ID 123, got %s", user.ID)
    }
}
```

Run tests:
```bash
make test                    # All tests
go test -v ./internal/users/ # Specific package
go test -cover ./...         # With coverage
```

## Common Patterns

### Handler Pattern
```go
func (h *Handler) GetResource(w http.ResponseWriter, r *http.Request) {
    // 1. Extract parameters
    id := r.PathValue("id")
    
    // 2. Call service
    data, err := h.service.GetResource(id)
    
    // 3. Handle error
    if err != nil {
        response := map[string]any{
            "status": "fail",
            "data": map[string]string{"id": "Resource not found"},
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(response)
        return
    }
    
    // 4. Return success
    response := map[string]any{
        "status": "success",
        "data": data,
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}
```

### Route Registration Pattern
```go
// internal/users/routes.go
func RegisterRoutes(mux *http.ServeMux) {
    // Initialize dependencies
    repo := repositories.NewUserRepository(db)
    service := services.NewUserService(repo)
    handler := handlers.NewUserHandler(service)
    
    // Register routes
    mux.HandleFunc("GET /users", handler.ListUsers)
    mux.HandleFunc("GET /users/{id}", handler.GetUser)
    mux.HandleFunc("POST /users", handler.CreateUser)
    mux.HandleFunc("PATCH /users/{id}", handler.UpdateUser)
    mux.HandleFunc("DELETE /users/{id}", handler.DeleteUser)
}
```

## Quick Reference

| Task | Command |
|------|---------|
| Start dev | `make dev` |
| View docs | http://localhost:8080/docs |
| Health check | http://localhost:8080/health |
| Run tests | `make test` |
| Check code | `make lint` |
| Regen docs | `make swagger` |
| Build prod | `make build` |
| Run migrations | `make migrate-up` |
| Rollback migration | `make migrate-down` |
| Create migration | `make migrate-create NAME=xxx` |
| Migration status | `make migrate-status` |

## Additional Resources

- [Standard Library ServeMux](https://pkg.go.dev/net/http#ServeMux)
- [Swagger Annotations](https://github.com/swaggo/swag#api-operation)
- [JSend Specification](https://github.com/omniti-labs/jsend)
- [Air Documentation](https://github.com/air-verse/air)
- [Scalar UI](https://github.com/scalar/scalar)
- [golang-migrate](https://github.com/golang-migrate/migrate)