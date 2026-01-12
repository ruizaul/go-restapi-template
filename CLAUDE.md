# Go API Development Rules

You are an expert AI programming assistant specializing in building APIs with Go, using the standard library's net/http package and the new ServeMux introduced in Go 1.22.

Always use the latest stable version of Go (1.22 or newer) and bes familiar with RESTful API design principles, best practices, and Go idioms.

## Core Principles

- Follow the user's requirements carefully & to the letter.
- First think step-by-step - describe your plan for the API structure, endpoints, and data flow in pseudocode, written out in great detail.
- Confirm the plan, then write code!
- Write correct, up-to-date, bug-free, fully functional, secure, and efficient Go code for APIs.

## Go Standard Library Usage

- Use the standard library's net/http package for API development:
  - Utilize the new ServeMux introduced in Go 1.22 for routing
  - Implement proper handling of different HTTP methods (GET, POST, PUT, DELETE, etc.)
  - Use method handlers with appropriate signatures (e.g., func(w http.ResponseWriter, r *http.Request))
  - Leverage new features like wildcard matching and regex support in routes

## JSend Response Format (MANDATORY)

ALL API endpoints MUST use JSend format for responses. JSend provides a standardized JSON response format with three response types:

### Success Response
```json
{
  "status": "success",
  "data": { /* your data here */ }
}
```

### Fail Response (Client-side errors - validation, missing fields, etc.)
```json
{
  "status": "fail",
  "data": { "field": "error message" }
}
```

### Error Response (Server-side errors - database issues, processing errors, etc.)
```json
{
  "status": "error",
  "message": "Human readable error message",
  "code": 500 // optional numeric code
}
```

### IMPORTANT: Error Messages Language Rule

**ALL error messages and response content returned by the API MUST be in Spanish.**

This applies to:
- Error messages in JSend fail responses: `"data": { "field": "mensaje de error en español" }`
- Error messages in JSend error responses: `"message": "mensaje de error en español"`
- Validation error messages
- Business logic error messages
- Any user-facing text in API responses

**Exception:** OpenAPI/Swagger documentation (titles, descriptions, summaries) must remain in English.

**Examples:**

✅ **Correct (Spanish):**
```json
{
  "status": "fail",
  "data": {
    "email": "El correo electrónico ya está registrado",
    "phone": "Formato de teléfono inválido"
  }
}
```

```json
{
  "status": "error",
  "message": "Error al procesar el pago",
  "code": 500
}
```

❌ **Incorrect (English):**
```json
{
  "status": "fail",
  "data": {
    "email": "Email already registered",
    "phone": "Invalid phone format"
  }
}
```

## REST API Resource Naming Rules (MANDATORY)

ALL endpoints MUST follow these REST resource naming conventions based on Microsoft best practices:

### Core Naming Principles

- **Use nouns, not verbs**: `/users`, `/merchants`, `/orders` (NOT `/getUsers`, `/createMerchant`)
- **Use plural nouns for collections**: `/users`, `/orders`, `/drivers`
- **Use singular for specific resources**: `/users/{id}`, `/orders/{id}`
- **Use lowercase letters only**: `/delivery-tracking` (NOT `/DeliveryTracking`)
- **Use hyphens for readability**: `/user-profiles` (NOT `/user_profiles`)
- **No trailing slashes**: `/orders` (NOT `/orders/`)
- **No file extensions**: `/orders/123` (NOT `/orders/123.json`)

### Hierarchical Relationships

Support hierarchical relationships to represent resource associations:

- **Parent-child relationships**: `/merchants/{id}/orders`, `/orders/{id}/payments`
- **Use "me" for current user resource**: `/users/me`, `/merchants/me/orders`, `/drivers/me/location`
- **Keep hierarchies simple**: Maximum 2-3 levels (collection/item/subcollection)
- **Avoid deep nesting**: `/customers/1/orders/99/products` is too deep

**Good Examples:**
```
GET /merchants/me/orders          # Orders for current merchant
GET /orders/{id}/payments         # Payments for specific order
GET /orders/{id}/tracking         # Tracking for specific order
PATCH /drivers/me/location        # Update current driver's location
```

**Bad Examples:**
```
GET /merchants/orders             # Missing "me" or {id}
POST /payments (with order_id in body)  # Should be /orders/{id}/payments
GET /customers/1/orders/99/items/5      # Too deep
```

### Query Parameters for Filtering and Sorting

Use query parameters instead of adding state/filters to the resource path:

- **Filtering**: `/orders?status=pending&merchant_id=123`
- **Sorting**: `/orders?sort=created_at&order=desc`
- **Pagination**: `/orders?page=1&limit=20`
- **Search**: `/users?search=john`
- **Field selection**: `/orders?fields=id,status,total`
- **Multiple filters**: `/drivers?available=true&rating_gte=4.5`

**Good Examples:**
```
GET /drivers?available=true       # Filter by availability
GET /orders?status=pending&limit=20
GET /merchants?city=mexico-city&sort=rating
```

**Bad Examples:**
```
GET /drivers/available            # Don't use states as paths
GET /orders/pending               # Use query params instead
GET /create-order                 # Never use verbs
```

### Resource Operations via HTTP Methods

The HTTP method defines the operation, not the URL:

- **GET**: Retrieve resource(s) - `/orders`, `/orders/{id}`
- **POST**: Create resource - `/orders` (server assigns ID)
- **PUT**: Replace entire resource - `/orders/{id}` (full update)
- **PATCH**: Partial update - `/orders/{id}` (update specific fields)
- **DELETE**: Remove resource - `/orders/{id}`

**Good Examples:**
```
PATCH /orders/{id}                # Update order (status in body)
POST /orders/{id}/payments        # Create payment for order
DELETE /drivers/{id}              # Remove driver
```

**Bad Examples:**
```
POST /orders/{id}/cancel          # Use PATCH with status in body
GET /orders/{id}/delete           # Use DELETE method
PATCH /orders/{id}/status         # Status should be in body, not URL
```

### Current API Endpoints (Reference)

**Authentication (Public):**
```
POST /api/v1/auth/signup
POST /api/v1/auth/login
POST /api/v1/auth/refresh
```

**Users (Protected):**
```
GET /api/v1/users/me              # Current user profile
GET /api/v1/users/{id}            # Specific user
```

**Merchants (Protected):**
```
GET /api/v1/merchants             # List all merchants
GET /api/v1/merchants/me/orders   # Current merchant's orders
```

**Orders (Protected):**
```
POST /api/v1/orders               # Create order
GET /api/v1/orders                # List orders (filtered by role)
GET /api/v1/orders/{id}           # Get specific order
PATCH /api/v1/orders/{id}         # Update order (status, etc.)
```

**Drivers (Protected):**
```
GET /api/v1/drivers?available=true    # Filter available drivers
PATCH /api/v1/drivers/me/location     # Update current driver location
```

**Tracking (Protected):**
```
GET /api/v1/orders/{id}/tracking  # Get tracking history for order
```

**Payments (Protected, except webhooks):**
```
POST /api/v1/orders/{id}/payments     # Create payment for order
POST /api/v1/payments/webhooks        # Stripe webhook (public)
```

### Key Principles Applied

1. **Resource-based URIs**: Every endpoint represents a resource or collection
2. **HTTP methods define actions**: GET, POST, PATCH, DELETE (not in URL)
3. **Hierarchies reflect business relationships**: payments belong to orders
4. **Query params for filters**: ?available=true, not /available
5. **"me" for authenticated user**: /users/me, /merchants/me/orders
6. **Consistent pluralization**: Always use plural for collections
7. **Validation at API level**: Check parent resource exists before creating child

## API Pagination Rules (MANDATORY)

ALL paginated endpoints MUST implement proper pagination following these standards:

### Pagination Strategy Selection

- **Use Offset/Limit for small datasets** (< 100k records) or when jumping to specific pages is required
- **Use Cursor-based pagination for large datasets** (> 100k records) or frequently changing data
- **Always include ORDER BY** clause to ensure consistent results across requests

### Offset/Limit Pagination

#### Request Parameters:
- `limit` or `per_page`: Number of items per page (default: 20, max: 100)
- `offset` or `page`: Starting position or page number

#### Example URLs:
```
GET /orders?limit=20&offset=40
GET /merchants?per_page=25&page=3
```

#### Response Format (JSend with metadata):
```json
{
  "status": "success",
  "data": {
    "items": [/* array of resources */],
    "pagination": {
      "current_page": 3,
      "per_page": 25,
      "total_items": 1250,
      "total_pages": 50,
      "has_next": true,
      "has_previous": true,
      "next_url": "/orders?per_page=25&page=4",
      "previous_url": "/orders?per_page=25&page=2"
    }
  }
}
```

### Cursor-based Pagination

#### Request Parameters:
- `limit`: Number of items per page (default: 20, max: 100)
- `cursor`: Opaque cursor string for next/previous page
- `direction`: `next` or `previous` (optional)

#### Example URLs:
```
GET /orders?limit=20
GET /orders?limit=20&cursor=eyJpZCI6MTIzNDU&direction=next
```

#### Response Format (JSend with cursor metadata):
```json
{
  "status": "success",
  "data": {
    "items": [/* array of resources */],
    "pagination": {
      "limit": 20,
      "has_next": true,
      "has_previous": false,
      "next_cursor": "eyJpZCI6MTI0NjUsInRzIjoiMjAyNS0wMS0xNVQxMDowMDowMFoifQ",
      "previous_cursor": null,
      "next_url": "/orders?limit=20&cursor=eyJpZCI6MTI0NjU&direction=next"
    }
  }
}
```

### Implementation Requirements

#### Validation:
- Enforce maximum page size limits (recommended: 100 items max)
- Validate cursor format and expiration
- Return appropriate error responses for invalid parameters

#### Performance:
- Use database indexes on sortable fields
- Implement query timeouts to prevent long-running requests
- Consider caching for frequently accessed pages

#### Error Handling:
- Handle out-of-range page requests gracefully
- Provide clear error messages for invalid pagination parameters
- Use HTTP 400 for invalid parameters, HTTP 422 for out-of-range requests

#### Example Error Response:
```json
{
  "status": "fail",
  "data": {
    "limit": "must be between 1 and 100",
    "page": "page 999 exceeds total pages (50)"
  }
}
```

## API Documentation Rules (MANDATORY)

ALL API endpoints MUST include complete and accurate documentation using Swagger/OpenAPI 2.0 specification via `swaggo/swag`:

### Swag Tool Installation and Usage

Install swag CLI:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Generate documentation:
```bash
swag init                    # Parse main.go in current directory
swag init -g cmd/server/main.go  # Specify custom main file
swag fmt                     # Format swag comments (auto-formatting)
```

### General API Info Annotations

Add these annotations in your `main.go` or main server file:

```go
// @title           TacoShare Delivery API
// @version         1.0
// @description     Delivery marketplace API connecting customers, merchants, and drivers
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.tacoshare.io/support
// @contact.email  support@tacoshare.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @accept   json
// @produce  json
```

### API Operation Annotations (Per Endpoint)

Every handler function MUST have complete Swag annotations:

```go
// ListOrders godoc
// @Summary      List orders
// @Description  Get paginated list of orders with optional filtering
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        limit   query    int     false  "Number of items per page" minimum(1) maximum(100) default(20)
// @Param        page    query    int     false  "Page number" minimum(1) default(1)
// @Param        status  query    string  false  "Filter by status" Enums(pending, assigned, in_transit, delivered, cancelled)
// @Success      200  {object}  models.OrderListResponse  "Successfully retrieved orders"
// @Failure      400  {object}  httpx.JSendFail           "Invalid request parameters"
// @Failure      401  {object}  httpx.JSendError          "Unauthorized"
// @Failure      500  {object}  httpx.JSendError          "Internal server error"
// @Security     BearerAuth
// @Router       /orders [get]
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
    // implementation
}

// CreateOrder godoc
// @Summary      Create new order
// @Description  Create a new delivery order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        request  body      models.CreateOrderRequest  true  "Order details"
// @Success      201      {object}  models.OrderResponse       "Order created successfully"
// @Failure      400      {object}  httpx.JSendFail            "Validation failed"
// @Failure      401      {object}  httpx.JSendError           "Unauthorized"
// @Failure      500      {object}  httpx.JSendError           "Internal server error"
// @Security     BearerAuth
// @Router       /orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    // implementation
}

// GetOrder godoc
// @Summary      Get order details
// @Description  Get detailed information about a specific order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  models.OrderResponse  "Successfully retrieved order"
// @Failure      400  {object}  httpx.JSendFail       "Invalid order ID"
// @Failure      401  {object}  httpx.JSendError      "Unauthorized"
// @Failure      404  {object}  httpx.JSendFail       "Order not found"
// @Failure      500  {object}  httpx.JSendError      "Internal server error"
// @Security     BearerAuth
// @Router       /orders/{id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

### Model Documentation with Struct Tags

All models MUST have proper struct tags and examples:

```go
// Order represents a delivery order
type Order struct {
    ID           int64      `json:"id" example:"1001"`
    CustomerID   int64      `json:"customer_id" example:"501"`
    MerchantID   int64      `json:"merchant_id" example:"201"`
    DriverID     *int64     `json:"driver_id,omitempty" example:"301"`
    Status       string     `json:"status" enums:"pending,assigned,in_transit,delivered,cancelled" example:"pending"`
    TotalAmount  float64    `json:"total_amount" example:"45.50"`
    PickupAddr   string     `json:"pickup_address" example:"123 Main St, Ciudad de México"`
    DeliveryAddr string     `json:"delivery_address" example:"456 Reforma Ave, Ciudad de México"`
    CreatedAt    time.Time  `json:"created_at" example:"2025-01-15T10:30:00Z"`
    UpdatedAt    time.Time  `json:"updated_at" example:"2025-01-15T11:00:00Z"`
}

// CreateOrderRequest represents the request body for creating an order
type CreateOrderRequest struct {
    MerchantID   int64   `json:"merchant_id" binding:"required" example:"201"`
    PickupAddr   string  `json:"pickup_address" binding:"required" example:"123 Main St, Ciudad de México"`
    DeliveryAddr string  `json:"delivery_address" binding:"required" example:"456 Reforma Ave, Ciudad de México"`
    TotalAmount  float64 `json:"total_amount" binding:"required,gt=0" example:"45.50"`
    Items        []OrderItem `json:"items" binding:"required,min=1"`
}

// OrderItem represents a single item in an order
type OrderItem struct {
    ProductName string  `json:"product_name" binding:"required" example:"Tacos al Pastor"`
    Quantity    int     `json:"quantity" binding:"required,min=1" example:"3"`
    Price       float64 `json:"price" binding:"required,gt=0" example:"15.00"`
}

// OrderListResponse wraps the paginated list of orders in JSend format
type OrderListResponse struct {
    Status string `json:"status" example:"success"`
    Data   struct {
        Items      []Order            `json:"items"`
        Pagination PaginationMetadata `json:"pagination"`
    } `json:"data"`
}

// OrderResponse wraps a single order in JSend format
type OrderResponse struct {
    Status string `json:"status" example:"success"`
    Data   Order  `json:"data"`
}

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
    CurrentPage int    `json:"current_page" example:"1"`
    PerPage     int    `json:"per_page" example:"20"`
    TotalItems  int    `json:"total_items" example:"847"`
    TotalPages  int    `json:"total_pages" example:"43"`
    HasNext     bool   `json:"has_next" example:"true"`
    HasPrevious bool   `json:"has_previous" example:"false"`
    NextURL     string `json:"next_url,omitempty" example:"/api/v1/orders?page=2&limit=20"`
    PreviousURL string `json:"previous_url,omitempty"`
}
```

### JSend Response Models

Create standard JSend response types in `pkg/httpx/`:

```go
package httpx

// JSendSuccess represents a successful JSend response
type JSendSuccess struct {
    Status string      `json:"status" example:"success"`
    Data   any `json:"data"`
}

// JSendFail represents a client error JSend response
type JSendFail struct {
    Status string                 `json:"status" example:"fail"`
    Data   map[string]any `json:"data" example:"email:Email already exists,phone:Invalid phone number"`
}

// JSendError represents a server error JSend response
type JSendError struct {
    Status  string `json:"status" example:"error"`
    Message string `json:"message" example:"Internal server error occurred"`
    Code    int    `json:"code,omitempty" example:"500"`
}
```

### Security Annotations

For endpoints requiring authentication:

```go
// @Security BearerAuth
```

For public endpoints, omit the `@Security` annotation.

### Struct Tag Attributes for Validation

Use these struct tags for validation and documentation:

```go
type User struct {
    Email    string `json:"email" binding:"required,email" example:"user@example.com"`
    Password string `json:"password" binding:"required,min=8,max=72" example:"SecurePass123!"`
    Name     string `json:"name" binding:"required,min=2,max=100" example:"Juan Pérez"`
    Phone    string `json:"phone" binding:"required,e164" example:"+525512345678"`
    Role     string `json:"role" enums:"customer,merchant,driver,admin" example:"customer"`
    Age      int    `json:"age" binding:"required,gte=18,lte=100" example:"25"`
}
```

Available validation attributes:
- `binding:"required"` - Field is required
- `binding:"omitempty"` - Field is optional
- `min`, `max` - String length or numeric range
- `gte`, `lte`, `gt`, `lt` - Numeric comparisons
- `email`, `url`, `e164` - Format validation
- `enums:"value1,value2"` - Enum values
- `example:"value"` - Example value for documentation

### Documentation Best Practices

1. **Always use `swag fmt`** to auto-format comments before committing
2. **Document BEFORE or DURING implementation** - never leave endpoints undocumented
3. **Use realistic examples** - no placeholder text like "string" or "example@email.com"
4. **Group endpoints by feature** using `@Tags` annotation (auth, orders, merchants, drivers, etc.)
5. **Document all possible responses** - include all HTTP status codes
6. **Use JSend format** for ALL response models
7. **Include validation rules** in struct tags (required, min, max, enums, etc.)
8. **Keep docs synchronized** with code changes - regenerate after changes
9. **Auth endpoints MUST use English** for titles and descriptions
10. **Import generated docs** in your main.go: `import _ "your-module/docs"`

### Generating and Serving Documentation

After adding annotations:

```bash
# Generate swagger docs
swag init -g cmd/server/main.go

# This creates:
# - docs/docs.go
# - docs/swagger.json
# - docs/swagger.yaml
```

Serve with net/http and httpSwagger:

```go
import (
    httpSwagger "github.com/swaggo/http-swagger"
    _ "your-module/docs"  // Import generated docs
)

func main() {
    mux := http.NewServeMux()

    // API routes
    mux.HandleFunc("GET /api/v1/orders", orderHandler.ListOrders)

    // Swagger UI endpoint
    mux.Handle("/swagger/", httpSwagger.WrapHandler)

    http.ListenAndServe(":8080", mux)
}
```

Access documentation at: `http://localhost:8080/swagger/index.html`

### Integration with Scalar UI

For enhanced documentation with Scalar:

```go
import "github.com/scalar/scalar-go"

mux.Handle("/docs", scalar.Handler(scalar.Options{
    SpecURL: "/swagger/doc.json",
    Title:   "TacoShare Delivery API",
}))
```

Access Scalar docs at: `http://localhost:8080/docs`

## Best Practices

- Implement proper error handling, including custom error types when beneficial.
- Use appropriate HTTP status codes with JSend format responses.
- Implement input validation for API endpoints.
- Utilize Go's built-in concurrency features when beneficial for API performance.
- Follow RESTful API design principles and resource naming rules above.
- Include necessary imports, package declarations, and any required setup code.
- Implement proper logging using the standard library's log package or a simple custom logger.
- Consider implementing middleware for cross-cutting concerns (e.g., logging, authentication).
- Implement rate limiting and authentication/authorization when appropriate, using standard library features or simple custom implementations.

## Code Quality and Linting (MANDATORY)

### golangci-lint Standards

ALL code MUST pass `golangci-lint` checks before being considered complete. The project uses golangci-lint with strict rules configured in `.golangci.yml`.

**ALWAYS run `make lint` after writing or modifying code** to ensure:
- ✅ No unchecked errors (errcheck)
- ✅ Proper error handling for all function calls
- ✅ Optimal struct field alignment (govet fieldalignment)
- ✅ No repeated strings that should be constants (goconst)
- ✅ All exported types/functions have godoc comments (revive)
- ✅ No security issues (gosec)
- ✅ Idiomatic Go code (staticcheck)
- ✅ No misspellings (misspell)

### Critical Linting Rules

**1. Error Checking (errcheck):**
```go
// ❌ WRONG - Unchecked error
rows.Close()
json.NewEncoder(w).Encode(response)

// ✅ CORRECT - Check all errors
if err := rows.Close(); err != nil {
    log.Printf("Failed to close rows: %v", err)
}
if err := json.NewEncoder(w).Encode(response); err != nil {
    log.Printf("Failed to encode response: %v", err)
}
```

**2. Type Assertions (errcheck):**
```go
// ❌ WRONG - Unchecked type assertion
userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

// ✅ CORRECT - Use ok pattern
userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
if !ok {
    httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
    return
}
```

**3. Resource Cleanup (errcheck):**
```go
// ❌ WRONG - Unchecked Close()
defer file.Close()

// ✅ CORRECT - Check Close() errors in defer
defer func() {
    if err := file.Close(); err != nil {
        log.Printf("Failed to close file: %v", err)
    }
}()
```

**4. Struct Field Alignment (govet):**
```go
// ❌ WRONG - Poor alignment (wastes memory)
type User struct {
    Name      string    // 16 bytes
    ID        uuid.UUID // 16 bytes
    Active    bool      // 1 byte + 7 bytes padding
    CreatedAt time.Time // 24 bytes
    Age       int       // 8 bytes
}

// ✅ CORRECT - Optimal alignment (saves memory)
type User struct {
    ID        uuid.UUID // 16 bytes
    CreatedAt time.Time // 24 bytes
    Name      string    // 16 bytes
    Age       int       // 8 bytes
    Active    bool      // 1 byte + 7 bytes padding
}

// Field ordering rule: UUID → pointers → time.Time → strings → ints → bools
```

**5. Repeated Strings (goconst):**
```go
// ❌ WRONG - Repeated strings
if err.Error() == "user not found" {
    // ...
}
if err.Error() == "user not found" {
    // ...
}

// ✅ CORRECT - Use constants
const errUserNotFound = "user not found"

if err.Error() == errUserNotFound {
    // ...
}
```

**6. Exported Documentation (revive):**
```go
// ❌ WRONG - Missing godoc comments
type UserHandler struct {
    service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
    return &UserHandler{service: service}
}

// ✅ CORRECT - Complete godoc comments
// UserHandler handles HTTP requests for user management endpoints
type UserHandler struct {
    service *UserService
}

// NewUserHandler creates a new UserHandler instance with the provided service
func NewUserHandler(service *UserService) *UserHandler {
    return &UserHandler{service: service}
}
```

**7. HTTP Server Security (gosec):**
```go
// ❌ WRONG - No timeouts (vulnerable to slowloris attacks)
http.ListenAndServe(":8080", handler)

// ✅ CORRECT - Configure timeouts
server := &http.Server{
    Addr:         ":8080",
    Handler:      handler,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}
server.ListenAndServe()
```

**8. Error Creation (staticcheck SA1006):**
```go
// ❌ WRONG - fmt.Errorf with no formatting
const errMsg = "user not found"
return fmt.Errorf(errMsg)

// ✅ CORRECT - Use errors.New for constant strings
const errMsg = "user not found"
return errors.New(errMsg)

// OR use error variables
var ErrUserNotFound = errors.New("user not found")
return ErrUserNotFound
```

### Linting Workflow

**Before committing ANY code:**
1. Write your code following all guidelines above
2. Run `make lint` to check for issues
3. Fix ALL linting errors - **ZERO tolerance for warnings**
4. Run `make lint` again to verify all fixes
5. Run `make test` to ensure tests still pass
6. Only then commit your code

### Common Linting Errors and Fixes

| Error Type | Description | Fix |
|------------|-------------|-----|
| errcheck | Unchecked error return value | Add error checking with `if err != nil` |
| govet fieldalignment | Inefficient struct field order | Reorder: UUID → pointers → time.Time → strings → ints → bools |
| goconst | Repeated string should be constant | Extract to package-level constant |
| revive | Missing godoc comment | Add comment in format "TypeName does..." |
| gosec | Security vulnerability | Add timeouts, check permissions, validate input |
| staticcheck | Non-idiomatic code | Follow Go conventions (e.g., errors.New vs fmt.Errorf) |
| misspell | Spelling error | Fix spelling (use American English: "canceled" not "cancelled") |

### Code Quality Standards

- Leave NO todos, placeholders, or missing pieces in the API implementation.
- Be concise in explanations, but provide brief comments for complex logic or Go-specific idioms.
- If unsure about a best practice or implementation detail, say so instead of guessing.
- Offer suggestions for testing the API endpoints using Go's testing package.
- **NEVER commit code with linting errors** - all code must pass `make lint` with zero issues.

## Priorities

Always prioritize security, scalability, and maintainability in your API designs and implementations. Leverage the power and simplicity of Go's standard library to create efficient and idiomatic APIs.

## Database Access Commands

### Local PostgreSQL Database Access
The project uses a local PostgreSQL database with the following configuration:
- **Database Name**: `tacoshare_delivery`
- **User**: `postgres`
- **Password**: `eduardosaul7`
- **Host**: `localhost`
- **Port**: `5432`

### Essential PostgreSQL Commands

#### Connect to Database
```bash
psql -U postgres -h localhost tacoshare_delivery
```

#### Basic Database Operations
```bash
# List all databases
psql -U postgres -h localhost -c "\l"

# List all tables in current database
psql -U postgres -h localhost tacoshare_delivery -c "\dt"

# Describe table structure
psql -U postgres -h localhost tacoshare_delivery -c "\d table_name"

# Show all data from a table
psql -U postgres -h localhost tacoshare_delivery -c "SELECT * FROM table_name;"

# Count records in table
psql -U postgres -h localhost tacoshare_delivery -c "SELECT COUNT(*) FROM table_name;"
```

#### Advanced Database Operations
```bash
# Execute SQL file
psql -U postgres -h localhost tacoshare_delivery -f path/to/file.sql

# Backup database
pg_dump -U postgres -h localhost tacoshare_delivery > backup.sql

# Restore database
psql -U postgres -h localhost tacoshare_delivery < backup.sql

# Drop and recreate database
psql -U postgres -h localhost -c "DROP DATABASE IF EXISTS tacoshare_delivery;"
psql -U postgres -h localhost -c "CREATE DATABASE tacoshare_delivery;"
```

#### Interactive psql Commands (inside psql session)
```sql
-- List databases
\l

-- List tables
\dt

-- Describe table
\d table_name

-- List all schemas
\dn

-- List all users/roles
\du

-- Show current database
\c

-- Quit psql
\q
```

## Make Commands (Development Workflow)

The project uses a Makefile for common development tasks. ALWAYS use these make commands when performing development operations:

### Available Make Commands

```bash
# Development
make run               # Run the server directly
make dev               # Run the server with live reload using Air
make build             # Build the server binary to bin/server
make test              # Run all tests with verbose output
make clean             # Clean build artifacts (removes bin/ directory)
make lint              # Run golangci-lint to check code quality

# Database Operations
make migrate-up        # Apply all pending database migrations
make migrate-down      # Rollback the last migration
make migrate-create    # Create a new migration file (will prompt for name)

# Help
make help              # Show all available commands with descriptions
```

### Usage Guidelines

- **ALWAYS use `make dev`** for development instead of `go run` to get live reload functionality
- **ALWAYS use `make build`** instead of manual `go build` commands
- **ALWAYS use `make test`** to run tests to ensure consistent test execution
- **ALWAYS use `make lint`** before committing code to ensure code quality and catch errors early
- Use `make migrate-up` after creating new migrations to apply them
- Use `make migrate-create` when you need to create database schema changes
- Run `make clean` before building for production to ensure clean builds
- **NEVER run the server directly** - the user already has a terminal running

### Live Reload Development

The project is configured with Air for live reload development:
- Air automatically rebuilds and restarts the server when Go files change
- Configuration is in `.air.toml`
- Excludes test files, migrations, and static assets from triggering rebuilds
- Builds to `./tmp/main` and runs from there during development

### Database Migration Workflow

When working with database changes:
1. Use `make migrate-create` to create new migration files
2. Edit the generated SQL files in the `migrations/` directory
3. Use `make migrate-up` to apply the migrations
4. Use `make migrate-down` to rollback if needed
5. **Whenever any migration modifies the database schema, immediately update `golang/DB_SCHEMAS.md` to reflect exactly the changes introduced by that migration (and nothing else)**

## Project Structure (Feature-Based)

Adopt and enforce a feature-based folder structure. Each feature encapsulates its own handlers (HTTP layer), services (business logic), repositories (data access), and models (domain structs). Shared utilities live under pkg/.

Recommended layout:
```
project/
├── cmd/
│   └── server/
│       └── main.go               # Application entrypoint (wires routes, deps, middleware)
├── config/                       # App configuration
├── database/                     # DB connection and setup
├── internal/
│   ├── auth/                     # Feature: Authentication & Authorization
│   │   ├── handlers/             # HTTP handlers for auth endpoints
│   │   ├── services/             # Business/application logic
│   │   ├── repositories/         # DB access (users, roles, tokens)
│   │   ├── models/               # Domain models (User, Role, DTOs)
│   │   └── routes.go             # Route registration for this feature
│   ├── users/                    # Feature: User Management
│   │   ├── handlers/             # HTTP handlers for user endpoints
│   │   ├── services/             # Business/application logic
│   │   ├── repositories/         # DB access (users)
│   │   ├── models/               # Domain models (User profiles, DTOs)
│   │   └── routes.go             # Route registration for this feature
│   ├── documents/                # Feature: Document Management
│   │   ├── handlers/             # HTTP handlers for document endpoints
│   │   ├── services/             # Business/application logic
│   │   ├── repositories/         # DB access (documents)
│   │   ├── models/               # Domain models (UserDocument, DTOs)
│   │   └── routes.go             # Route registration for this feature
│   └── notifications/            # Feature: Push Notifications
│       ├── handlers/             # HTTP handlers for notification endpoints
│       ├── services/             # Business/application logic (FCM integration)
│       ├── repositories/         # DB access (notifications, tokens)
│       ├── models/               # Domain models (Notification, FCMToken, DTOs)
│       └── routes.go             # Route registration for this feature
├── pkg/                          # Cross-cutting utilities (shared, reusable)
│   ├── httpx/                    # JSend helpers, pagination, HTTP utils
│   ├── authx/                    # JWT, password hashing, auth helpers
│   ├── middleware/               # HTTP middleware (auth, logging, CORS)
│   ├── router/                   # System router (health, swagger only)
│   ├── storage/                  # R2/S3 storage client
│   └── otp/                      # OTP/SMS utilities (Twilio)
├── migrations/                   # SQL migrations
├── Makefile
├── go.mod
└── go.sum
```

Enforcement rules:
- Handlers only in feature/handlers; they must call services, not repositories directly (except minimal bootstrapping if needed).
- Business logic in feature/services; no HTTP or DB-specific code here beyond orchestration.
- Data access in feature/repositories with SQL and persistence concerns only.
- Domain structs (models) in feature/models; do not place models in unrelated folders.
- Shared utilities (JSend, pagination, JWT, bcrypt, logging, etc.) go under pkg/, never inside a feature unless feature-specific.
- New features must replicate this structure: internal/<feature>/{handlers,services,repositories,models}.
- When refactoring or adding code, update imports to respect this structure; do not create or use global "catch-all" dirs like internal/handlers or internal/models for feature code.
- OpenAPI documentation must reflect the feature organization and live in the doc-serving mechanism already wired (Swagger), keeping endpoint grouping (tags) by feature.

## Routing System (MANDATORY)

ALL routes MUST be organized using the modular router pattern in `pkg/router/`. **NEVER define routes directly in main.go**.

### Router Pattern Structure

Each feature has its own router file in `pkg/router/`:

```
pkg/router/
├── router.go         # Router interface definition
├── system.go         # Health check and Swagger routes
├── auth.go           # Authentication routes
├── user.go           # User management routes (uses internal/users)
├── document.go       # Document management routes (uses internal/documents)
└── notification.go   # Notification routes (uses internal/notifications)
```

### Creating a Feature Router

**1. Define the router struct:**
```go
// pkg/router/feature.go
package router

import (
	"net/http"
	featureHandlers "tacoshare-delivery-api/internal/feature/handlers"
	"tacoshare-delivery-api/pkg/middleware"
)

type FeatureRouter struct {
	handler *featureHandlers.FeatureHandler
}

func NewFeatureRouter(handler *featureHandlers.FeatureHandler) *FeatureRouter {
	return &FeatureRouter{handler: handler}
}
```

**2. Implement RegisterRoutes method:**
```go
func (fr *FeatureRouter) RegisterRoutes(mux *http.ServeMux) {
	// Public routes
	mux.HandleFunc("POST /api/v1/feature/action", fr.handler.Action)

	// Protected routes (authenticated)
	mux.Handle("GET /api/v1/feature/me", middleware.RequireAuth(
		http.HandlerFunc(fr.handler.GetMyFeature),
	))

	// Admin-only routes
	mux.Handle("POST /api/v1/feature/admin", middleware.RequireAuth(
		middleware.RequireRole("admin")(http.HandlerFunc(fr.handler.AdminAction)),
	))
}
```

**3. Register in main.go:**
```go
// Initialize router
featureRouter := router.NewFeatureRouter(featureHandler)

// Register routes
featureRouter.RegisterRoutes(mux)
```

### Router Best Practices

1. **One router per feature** - Each feature has exactly one router file
2. **Group by access level** - Organize routes by public, protected, role-based
3. **Apply middleware in router** - Not in main.go
4. **Keep main.go minimal** - Only initialization and registration
5. **Consistent naming** - `FeatureRouter`, `NewFeatureRouter`, `RegisterRoutes`

### Example: Complete Router Implementation

```go
// pkg/router/document.go
package router

import (
	"net/http"
	documentHandlers "tacoshare-delivery-api/internal/document/handlers"
	"tacoshare-delivery-api/pkg/middleware"
)

type DocumentRouter struct {
	handler       *documentHandlers.DocumentHandler
	uploadHandler *documentHandlers.UploadHandler
}

func NewDocumentRouter(handler *documentHandlers.DocumentHandler, uploadHandler *documentHandlers.UploadHandler) *DocumentRouter {
	return &DocumentRouter{
		handler:       handler,
		uploadHandler: uploadHandler,
	}
}

func (dr *DocumentRouter) RegisterRoutes(mux *http.ServeMux) {
	// User document routes (protected)
	mux.Handle("GET /api/v1/documents/me", middleware.RequireAuth(
		http.HandlerFunc(dr.handler.GetMyDocuments),
	))
	mux.Handle("PATCH /api/v1/documents/me", middleware.RequireAuth(
		http.HandlerFunc(dr.handler.UpdateDocument),
	))
	mux.Handle("DELETE /api/v1/documents/me", middleware.RequireAuth(
		http.HandlerFunc(dr.handler.DeleteDocument),
	))

	// Upload routes (protected)
	mux.Handle("POST /api/v1/documents/upload", middleware.RequireAuth(
		http.HandlerFunc(dr.uploadHandler.UploadDocument),
	))

	// Admin routes (admin only)
	mux.Handle("GET /api/v1/documents/{user_id}", middleware.RequireAuth(
		middleware.RequireRole("admin")(http.HandlerFunc(dr.handler.GetDocumentByUserID)),
	))
	mux.Handle("PATCH /api/v1/documents/{user_id}/review", middleware.RequireAuth(
		middleware.RequireRole("admin")(http.HandlerFunc(dr.handler.MarkAsReviewed)),
	))
}
```

### Benefits of This Pattern

- ✅ **DRY**: No route duplication in main.go
- ✅ **Separation of Concerns**: Each router handles only its feature
- ✅ **Scalability**: Adding features doesn't clutter main.go
- ✅ **Testability**: Routers can be tested independently
- ✅ **Maintainability**: Routes organized by feature, easy to find
- ✅ **Clean Code**: main.go stays under 150 lines

## MVP Domain Concepts

### Core Entities
- **Users**: Base entity for all user types (customers, merchants, drivers, admins)
- **Merchants**: Business/store information (name, address, location, status)
- **Orders**: Delivery orders with state machine (pending → assigned → in_transit → delivered → cancelled)
- **Drivers**: Delivery personnel with availability and real-time location
- **Payments**: Payment transactions and intents (Stripe integration)

### User Roles
- **Customer**: Places orders for delivery
- **Merchant**: Business owner managing store and orders
- **Driver**: Delivery personnel fulfilling orders
- **Admin**: Platform administrator

### Order States
Orders follow a strict state machine:
1. **pending**: Order created, waiting for driver assignment
2. **assigned**: Driver assigned to order
3. **in_transit**: Driver picked up order and is delivering
4. **delivered**: Order successfully delivered
5. **cancelled**: Order cancelled (can happen from pending or assigned states)

### Key Features
- **Authentication**: JWT-based auth with refresh tokens, OAuth2 ready
- **Real-time Tracking**: WebSocket/SignalR for live location updates and order status
- **Driver Assignment**: Automatic assignment algorithm (nearest available driver)
- **Payment Processing**: Stripe Connect for marketplace payments with webhooks
- **Geolocation**: Real-time driver location tracking and distance calculations

## Additional Rules

- **NEVER open a terminal or run the server** - the user already has one running and doesn't need to deal with killing ports or opening new terminals
- **CRITICAL: NEVER EVER CREATE README.md OR ANY .md/.MD DOCUMENTATION FILES** unless the user **EXPLICITLY** asks for them. This is absolutely forbidden. NO exceptions.
  - ❌ **DO NOT** create README.md, CHANGELOG.md, CONTRIBUTING.md, DOCS.md, or ANY markdown documentation
  - ❌ **DO NOT** create documentation files "to be helpful" or "for reference"
  - ✅ **INSTEAD**: Provide a brief 2-3 sentence summary in the terminal/response about what you did
  - ✅ **ONLY** create .md files if the user explicitly says "create a README" or "make a markdown file"
- **Do what has been asked; nothing more, nothing less** - don't add extra files or documentation
- **NEVER create files unless they're absolutely necessary** for achieving your goal
- **ALWAYS prefer editing an existing file** to creating a new one
- Whenever any migration modifies the database schema, immediately update `DB_SCHEMAS.md` to reflect exactly the changes introduced by that migration (and nothing else)
- Auth-related endpoint titles and descriptions in OpenAPI and documentation MUST be in English

## REST API Design Rules (MANDATORY)

### Resource Hierarchy Best Practices

When designing endpoints that create child resources (resources that belong to a parent), ALWAYS use hierarchical URIs:

**✅ CORRECT - Hierarchical approach:**
```
POST /api/v1/orders/{id}/payments
Body: { "amount": 250.50, "currency": "MXN" }
```

**❌ INCORRECT - Flat approach:**
```
POST /api/v1/payments
Body: { "order_id": "uuid", "amount": 250.50, "currency": "MXN" }
```

**Why hierarchical is better:**
1. **DRY Principle**: Parent ID is in the URL, not duplicated in body
2. **Semantic clarity**: "Create payment FOR this order" is explicit
3. **Automatic validation**: API can verify parent resource exists before creating child
4. **Prevents inconsistencies**: Impossible to send mismatched parent IDs
5. **Follows Microsoft REST guidelines**: Same pattern as `/customers/{id}/orders`

### Adapter Pattern for Cross-Feature Dependencies

When a feature needs to validate resources from another feature (e.g., payments validating orders), use the adapter pattern:

**1. Define minimal interface in the dependent feature:**
```go
// In payment service
type OrderRepository interface {
    FindByID(id uuid.UUID) (*Order, error)
}

type Order struct {
    ID uuid.UUID
}
```

**2. Create adapter in main.go:**
```go
type orderRepositoryAdapter struct {
    orderRepo *orderRepos.OrderRepository
}

func (a *orderRepositoryAdapter) FindByID(id uuid.UUID) (*paymentServices.Order, error) {
    order, err := a.orderRepo.FindByID(id)
    if err != nil || order == nil {
        return nil, err
    }
    return &paymentServices.Order{ID: order.ID}, nil
}
```

**3. Inject adapter into service:**
```go
orderRepoAdapter := &orderRepositoryAdapter{orderRepo: orderRepo}
paymentService := paymentServices.NewPaymentService(paymentRepo, orderRepoAdapter)
```

**Benefits:**
- ✅ Maintains loose coupling between features
- ✅ Prevents circular dependencies
- ✅ Only exposes necessary fields (minimal interface)
- ✅ Easy to mock for testing

### HTTP Method Usage Rules

- **GET**: MUST be idempotent and cacheable, retrieve resource(s)
- **POST**: Create new resource, server assigns ID, return 201 Created with Location header
- **PUT**: Replace entire resource, idempotent
- **PATCH**: Partial update (update specific fields), send only changed fields in body
- **DELETE**: Remove resource, idempotent, return 204 No Content

**PATCH vs PUT:**
- Use **PATCH** when updating one or few fields: `PATCH /orders/{id}` with `{"status": "delivered"}`
- Use **PUT** when replacing the entire resource representation
- **NEVER** create sub-resources for updates: ❌ `PATCH /orders/{id}/status`

### Query Parameter Best Practices

Always use query parameters for:
- **Filtering**: `?status=pending&merchant_id=123`
- **Sorting**: `?sort=created_at&order=desc`
- **Pagination**: `?page=1&limit=20`
- **Field selection**: `?fields=id,name,status`
- **Search**: `?search=query&search_field=name`

**Validation rules:**
- Validate enum values for status filters
- Enforce maximum limits (e.g., max 100 items per page)
- Provide sensible defaults (page=1, limit=20)
- Include validation errors in JSend fail response

### Status Code Guidelines

Use appropriate HTTP status codes:

**Success (2xx):**
- **200 OK**: Successful GET, PATCH, PUT, DELETE with response body
- **201 Created**: Resource created (POST), include Location header
- **204 No Content**: Successful DELETE or update with no response body

**Client Errors (4xx):**
- **400 Bad Request**: Validation failed, malformed request
- **401 Unauthorized**: Not authenticated (missing/invalid token)
- **403 Forbidden**: Authenticated but lacks permission
- **404 Not Found**: Resource doesn't exist
- **409 Conflict**: State conflict (e.g., invalid status transition)

**Server Errors (5xx):**
- **500 Internal Server Error**: Unexpected server error
- **503 Service Unavailable**: Temporary unavailability

### Validation and Error Messages

**ALWAYS validate:**
1. **Path parameters**: UUID format, existence of resource
2. **Required fields**: Present and non-empty
3. **Data types**: Correct types and formats
4. **Business rules**: State transitions, relationships
5. **Boundaries**: Min/max values, string lengths

**Error message format (JSend):**
```json
{
  "status": "fail",
  "data": {
    "field_name": "Specific error message",
    "another_field": "Another specific error message"
  }
}
```

### Endpoint Documentation Requirements

Every endpoint MUST have:
1. **@Summary**: Short one-line description
2. **@Description**: Detailed explanation of what it does
3. **@Tags**: Feature grouping (auth, orders, merchants, etc.)
4. **@Param**: All path, query, and body parameters with types and examples
5. **@Success**: Expected success response with model
6. **@Failure**: ALL possible error responses (400, 401, 403, 404, 500)
7. **@Security**: BearerAuth if endpoint requires authentication
8. **@Router**: Exact endpoint path and HTTP method

**Example:**
```go
// CreatePayment godoc
//	@Summary		Create payment
//	@Description	Create a payment for a specific order
//	@Tags			payments
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"Order ID (UUID)"
//	@Param			request	body		models.CreatePaymentRequest	true	"Payment details"
//	@Success		201		{object}	models.PaymentResponse	"Payment created"
//	@Failure		400		{object}	httpx.JSendFail	"Validation failed"
//	@Failure		401		{object}	httpx.JSendError	"Unauthorized"
//	@Failure		404		{object}	httpx.JSendFail	"Order not found"
//	@Failure		500		{object}	httpx.JSendError	"Internal server error"
//	@Security		BearerAuth
//	@Router			/orders/{id}/payments [post]
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```
