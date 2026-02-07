# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies - update index and install
RUN apk update && apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /build

# Copy dependency files first for better caching
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with static linking to avoid runtime dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o app ./cmd/server

# Runtime stage - use scratch for minimal image
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder
COPY --from=builder /build/app /app

# Expose application port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app"]
