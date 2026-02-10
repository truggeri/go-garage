.PHONY: help build test run clean fmt lint vet install-tools seed db-backup

# Default target when just running 'make'
help:
	@echo "Go-Garage Development Commands"
	@echo "=============================="
	@echo "make build        - Compile the application"
	@echo "make test         - Execute all tests"
	@echo "make run          - Start the application"
	@echo "make clean        - Remove build artifacts"
	@echo "make fmt          - Format code with gofmt"
	@echo "make lint         - Run linting checks"
	@echo "make vet          - Run go vet analysis"
	@echo "make install-tools - Install development dependencies"
	@echo "make seed         - Seed database with sample data"
	@echo "make db-backup    - Create a timestamped database backup"

# Build the application binary
build:
	@echo "Building application..."
	@go build -o bin/go-garage ./cmd/server

# Run all tests with verbose output
test:
	@echo "Running tests..."
	@go test -v -cover ./...

# Start the application
run:
	@echo "Starting application..."
	@go run ./cmd/server

# Clean up build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean

# Format all Go source files
fmt:
	@echo "Formatting code..."
	@gofmt -s -w .

# Run static analysis with golangci-lint
lint:
	@echo "Running linter..."
	@golangci-lint run

# Run go vet for static analysis
vet:
	@echo "Running go vet..."
	@go vet ./...

# Install required development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Seed the database with sample data
seed:
	@echo "Seeding database..."
	@go run ./cmd/seed

# Create a timestamped backup of the database
db-backup:
	@echo "Creating database backup..."
	@go run ./cmd/dbutil backup
