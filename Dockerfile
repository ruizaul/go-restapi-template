# ============================================================================
# DOCKERFILE - Optimized multi-stage build for TacoShare Delivery API
# ============================================================================
# References:
# - Multi-stage builds for minimal image size
# - Static binary compilation with CGO_ENABLED=0
# - Distroless base image for security
# - Non-root user execution
# ============================================================================

# ============================================================================
# BUILD STAGE - Compile Go binary
# ============================================================================
FROM --platform=linux/amd64 golang:1.25.1-alpine AS builder

# Install build dependencies (minimal set for Go compilation)
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    file \
    curl

# Install golang-migrate for database migrations
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Set working directory
WORKDIR /build

# ============================================================================
# DEPENDENCY CACHING OPTIMIZATION
# Copy go.mod and go.sum first to leverage Docker layer caching
# Dependencies are downloaded only when these files change
# ============================================================================
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ============================================================================
# SOURCE CODE COPY
# Copy only necessary directories (excludes files via .dockerignore)
# ============================================================================
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY config/ ./config/
COPY database/ ./database/
COPY docs/ ./docs/

# ============================================================================
# BUILD BINARY
# Optimizations applied:
# - CGO_ENABLED=0: Static binary (no libc dependency, works with distroless)
# - GOOS=linux GOARCH=amd64: Target Cloud Run architecture
# - -ldflags="-w -s": Strip debug info and symbol table (reduces size ~30%)
# - -trimpath: Remove file system paths (security hardening)
# ============================================================================
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -trimpath \
    -o server \
    ./cmd/server

# Verify binary was created and show size
RUN ls -lh /build/server && \
    echo "Binary architecture:" && \
    file /build/server

# ============================================================================
# RUNTIME STAGE - Alpine-based minimal image with shell support
# ============================================================================
# Using Alpine instead of distroless for migration script support:
# - Includes /bin/sh for running migrate.sh
# - Much smaller than full Debian (~5MB vs ~100MB)
# - Runs as non-root user for security
# ============================================================================
FROM --platform=linux/amd64 alpine:3.19

# Install required runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 65532 -S nonroot && \
    adduser -u 65532 -S nonroot -G nonroot

# Set working directory
WORKDIR /app

# Copy application binary from builder stage
COPY --from=builder /build/server /app/server

# Copy golang-migrate binary for database migrations
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Copy database migrations (required for runtime schema management)
COPY migrations/ /app/migrations/

# Copy universal entrypoint script (handles both server and migration modes)
COPY scripts/db_migration.sh /app/db_migration.sh

# Make script executable
RUN chmod +x /app/db_migration.sh

# Change ownership to nonroot user
RUN chown -R nonroot:nonroot /app

# Switch to non-root user
USER nonroot

# ============================================================================
# NETWORKING CONFIGURATION
# ============================================================================
# NOTE: Do NOT use EXPOSE or ENV PORT in Cloud Run
# Cloud Run automatically injects the PORT environment variable at runtime
# The application reads it via os.Getenv("PORT") in main.go
# Using EXPOSE or ENV PORT can cause port conflicts and startup failures
# ============================================================================

# ============================================================================
# HEALTH CHECK
# Cloud Run provides its own health checks via HTTP probes
# No need for HEALTHCHECK directive (not supported in Cloud Run anyway)
# Application health endpoint: GET /health
# ============================================================================

# ============================================================================
# APPLICATION ENTRYPOINT
# Universal entrypoint that handles both modes:
# - Server mode (default): Runs /app/server
# - Migration mode: Runs migrations when RUN_MIGRATIONS=true
# All logic is contained in a single script for simplicity
# ============================================================================
ENTRYPOINT ["/app/db_migration.sh"]
