# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY database/ ./database/

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -trimpath \
    -o server \
    ./cmd/server

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 65532 -S appgroup && \
    adduser -u 65532 -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/server /app/server

# Set ownership
RUN chown -R appuser:appgroup /app

USER appuser

# Default port (Cloud Run will override via PORT env var)
ENV PORT=8080

EXPOSE 8080

CMD ["/app/server"]
