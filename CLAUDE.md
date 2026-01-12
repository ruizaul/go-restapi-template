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
migrations/           # SQL database migrations
docs/                 # Auto-generated API docs (DO NOT EDIT)
```

## Development Commands

```bash
make dev      # 🔥 Hot reload with Air (recommended)
make run      # Run server once (regenerates docs)
make build    # Compile binary to bin/server
make swagger  # Regenerate API documentation only
make test     # Run all tests
make lint     # Code quality checks (REQUIRED before commit)
make clean    # Remove build artifacts
```

## Development Workflow

1. Start development server: `make dev`
2. Edit code - Air watches changes and auto-restarts
3. View docs at http://localhost:8080/docs
4. Before committing:
   - Run `make lint` - fix ALL errors (zero tolerance)
   - Run `make test` - ensure all tests pass

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

## Additional Resources

- [Standard Library ServeMux](https://pkg.go.dev/net/http#ServeMux)
- [Swagger Annotations](https://github.com/swaggo/swag#api-operation)
- [JSend Specification](https://github.com/omniti-labs/jsend)
- [Air Documentation](https://github.com/air-verse/air)
- [Scalar UI](https://github.com/scalar/scalar)