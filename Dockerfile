# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies - update index and install
RUN apk update && apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /build

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-linkmode external -extldflags '-static' -w -s" -o app ./cmd/server

# Runtime stage - use alpine for minimal image with libc support
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Create data directory for SQLite database
RUN mkdir -p /app/data && chown -R appuser:appuser /app

# Copy binary, migrations, and web assets from builder
COPY --from=builder /build/app /app/app
COPY --from=builder /build/migrations /app/migrations
COPY --from=builder /build/web/static /app/web/static
COPY --from=builder /build/web/templates /app/web/templates

# Set working directory
WORKDIR /app

# Switch to non-root user
USER appuser

# Expose application port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/app"]
