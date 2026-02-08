# Development Setup Guide

This guide provides detailed instructions for setting up your development environment for Go-Garage.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Initial Setup](#initial-setup)
- [Development Environment](#development-environment)
- [Building and Running](#building-and-running)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [Database Management](#database-management)
- [Docker Development](#docker-development)
- [IDE Setup](#ide-setup)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Software

- **Go**: Version 1.24 or later
  - Download from https://golang.org/dl/
  - Verify installation: `go version`

- **SQLite3**: For local database
  - macOS: `brew install sqlite3`
  - Ubuntu/Debian: `sudo apt-get install sqlite3`
  - Windows: Download from https://www.sqlite.org/download.html

- **Git**: For version control
  - Download from https://git-scm.com/downloads
  - Verify installation: `git --version`

### Recommended Software

- **Make**: For convenience commands
  - macOS: `brew install make`
  - Ubuntu/Debian: `sudo apt-get install build-essential`
  - Windows: Use WSL or Git Bash

- **Docker & Docker Compose**: For containerized development
  - Download Docker Desktop from https://www.docker.com/products/docker-desktop
  - Verify installation: `docker --version && docker compose version`

- **golangci-lint**: For code linting
  - Install: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
  - Verify installation: `golangci-lint --version`

## Initial Setup

### 1. Clone the Repository

```bash
# Clone via HTTPS
git clone https://github.com/truggeri/go-garage.git
cd go-garage

# Or clone via SSH
git clone git@github.com:truggeri/go-garage.git
cd go-garage
```

### 2. Install Go Dependencies

```bash
# Download all dependencies
go mod download

# Verify dependencies
go mod verify
```

### 3. Configure Environment

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your preferred settings
# For development, the defaults are usually fine
```

#### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `development` | Application environment (development/production) |
| `APP_PORT` | `8080` | Port the server listens on |
| `SERVER_HOST` | `0.0.0.0` | Host the server binds to |
| `LOG_LEVEL` | `info` | Logging level (debug/info/warn/error) |
| `LOG_FORMAT` | `json` | Log output format (json/text) |
| `DB_PATH` | `./data/go-garage.db` | Path to SQLite database file |
| `JWT_SECRET` | _(optional in dev)_ | JWT secret key for authentication |

### 4. Set Up Pre-commit Hooks

```bash
# Copy the pre-commit hook
cp .githooks/pre-commit .git/hooks/pre-commit

# Make it executable
chmod +x .git/hooks/pre-commit
```

The pre-commit hook will:
- Automatically format Go files with `gofmt`
- Run `go vet` to catch common issues
- Prevent commits if checks fail

### 5. Install Development Tools (Optional)

```bash
# Install all development tools
make install-tools

# Or install individually
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Development Environment

### Project Structure

Understanding the project layout:

```
go-garage/
├── cmd/server/              # Application entry point
│   └── main.go             # Server initialization
├── internal/               # Private application code
│   ├── config/             # Configuration management
│   ├── handlers/           # HTTP request handlers
│   ├── middleware/         # HTTP middleware (logging, recovery)
│   ├── models/             # Data models
│   ├── repositories/       # Data access layer
│   └── services/           # Business logic
├── pkg/                    # Public packages
│   └── applog/             # Logging utilities
├── web/                    # Frontend assets
│   ├── static/             # CSS, JavaScript, images
│   └── templates/          # HTML templates
├── migrations/             # Database migrations
├── docs/                   # Documentation
├── spec/                   # Project specifications
├── tests/                  # Integration tests
├── .env.example           # Example environment configuration
├── .golangci.yml          # Linter configuration
├── Dockerfile             # Docker image definition
├── docker-compose.yml     # Docker Compose configuration
├── Makefile               # Common development tasks
└── go.mod                 # Go module definition
```

### Key Packages

- **config**: Loads and validates configuration from environment variables
- **handlers**: HTTP request handlers (controllers)
- **middleware**: HTTP middleware for logging, recovery, authentication
- **models**: Data structures and business entities
- **repositories**: Database access layer (CRUD operations)
- **services**: Business logic and orchestration
- **applog**: Structured logging with log/slog

## Building and Running

### Using Make Commands

```bash
# Build the application
make build
# Creates binary at: bin/go-garage

# Run the application
make run
# Starts server on http://localhost:8080

# View all available commands
make help
```

### Using Go Commands Directly

```bash
# Build
go build -o bin/go-garage ./cmd/server

# Run without building
go run ./cmd/server

# Build with race detector (for development)
go build -race -o bin/go-garage ./cmd/server
```

### Verifying the Application

Once running, test the health endpoint:

```bash
# Using curl
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy"}

# Using httpie (if installed)
http GET http://localhost:8080/health
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests with race detector
go test -race ./...

# Run tests for a specific package
go test ./internal/config/...

# Run a specific test
go test -run TestConfig_Load ./internal/config/
```

### Writing Tests

Follow these conventions:

1. **File naming**: Use `_test.go` suffix (e.g., `config_test.go`)

2. **Test naming**: Use descriptive names
   ```go
   func TestFunctionName_Scenario_ExpectedBehavior(t *testing.T)
   ```

3. **Structure**: Use Arrange-Act-Assert pattern
   ```go
   func TestVehicle_Create_ValidInput_Success(t *testing.T) {
       // Arrange
       vehicle := &Vehicle{Make: "Toyota"}
       
       // Act
       err := vehicle.Validate()
       
       // Assert
       assert.NoError(t, err)
   }
   ```

4. **Use testify**: For better assertions
   ```go
   import "github.com/stretchr/testify/assert"
   
   assert.Equal(t, expected, actual)
   assert.NoError(t, err)
   assert.True(t, condition)
   ```

### Test Coverage Goals

- Minimum: 80% coverage
- Target: 90% coverage for critical paths
- Business logic (services): 95%+ coverage

## Code Quality

### Formatting

```bash
# Format all Go files
make fmt

# Or use gofmt directly
gofmt -s -w .

# Check formatting without modifying
gofmt -l .
```

### Static Analysis

```bash
# Run go vet
make vet

# Or directly
go vet ./...
```

### Linting

```bash
# Run all linters
make lint

# Or use golangci-lint directly
golangci-lint run

# Run specific linters
golangci-lint run --disable-all --enable=errcheck
```

### Enabled Linters

The project uses the following linters (configured in `.golangci.yml`):

- **errcheck**: Check for unchecked errors
- **gosimple**: Simplify code
- **govet**: Standard Go analyzer
- **ineffassign**: Detect ineffectual assignments
- **staticcheck**: Go static analysis
- **unused**: Find unused code
- **misspell**: Find misspelled words
- **unparam**: Find unused function parameters
- **unconvert**: Remove unnecessary type conversions
- **goconst**: Find repeated strings that should be constants

### Pre-commit Checks

Before committing, ensure:

```bash
# Format code
make fmt

# Run tests
make test

# Run linters
make lint
make vet

# Build successfully
make build
```

Or use the pre-commit hook that does this automatically.

## Database Management

### Database Location

- Development: `./data/go-garage.db` (configurable via `DB_PATH`)
- Docker: `/app/data/go-garage.db` (persisted in Docker volume)

### Creating the Database

The database is created automatically on first run. The application will:
1. Create the directory if it doesn't exist
2. Initialize the SQLite database file
3. Run any pending migrations

### Manual Database Operations

```bash
# Open SQLite CLI
sqlite3 ./data/go-garage.db

# List tables
.tables

# Describe table structure
.schema vehicles

# Exit SQLite
.quit
```

### Database Migrations

Migrations are located in the `migrations/` directory.

```bash
# Migrations will be applied automatically on startup
# No manual intervention needed for development
```

## Docker Development

### Using Docker Compose (Recommended)

```bash
# Start the application
docker compose up -d

# View logs
docker compose logs -f

# Restart after code changes
docker compose restart

# Rebuild after dependency changes
docker compose up -d --build

# Stop the application
docker compose down

# Stop and remove volumes (fresh start)
docker compose down -v
```

### Using Docker Directly

```bash
# Build the image
docker build -t go-garage .

# Run the container
docker run -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -e LOG_LEVEL=debug \
  go-garage

# Run with custom environment file
docker run -p 8080:8080 \
  --env-file .env \
  -v $(pwd)/data:/app/data \
  go-garage
```

### Docker Development Tips

- Use `docker compose logs -f` to tail logs
- Database persists in `garage-data` volume
- Rebuild image after changing dependencies: `docker compose build`
- Access running container: `docker compose exec web sh`

For more Docker information, see [docs/DOCKER.md](DOCKER.md).

## IDE Setup

### Visual Studio Code

Recommended extensions:

1. **Go** (golang.go)
   - Official Go extension
   - Provides IntelliSense, formatting, debugging

2. **Go Test Explorer** (premparihar.gotestexplorer)
   - Test explorer sidebar

3. **Error Lens** (usernamehw.errorlens)
   - Inline error highlighting

#### VS Code Settings

Create or update `.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "gofmt",
  "go.formatFlags": ["-s"],
  "editor.formatOnSave": true,
  "go.testFlags": ["-v"],
  "go.coverOnSave": true
}
```

### GoLand / IntelliJ IDEA

1. Open the project folder
2. GoLand will automatically detect the Go module
3. Set SDK to Go 1.24+
4. Enable "Go Modules" in Preferences → Go → Go Modules

#### GoLand Settings

- **Code Style**: Use `gofmt` formatting
- **Inspections**: Enable all Go inspections
- **File Watchers**: Add `gofmt` file watcher for auto-formatting

## Troubleshooting

### Common Issues

#### Port Already in Use

```bash
# Error: address already in use
# Solution: Find and kill the process using port 8080

# On macOS/Linux
lsof -ti:8080 | xargs kill -9

# Or change the port in .env
APP_PORT=8081
```

#### Build Fails with CGO Error

```bash
# Error: C compiler not found or CGO_ENABLED required
# SQLite requires CGO

# Solution: Install C compiler
# macOS: Install Xcode Command Line Tools
xcode-select --install

# Ubuntu/Debian
sudo apt-get install build-essential

# Verify CGO is enabled
go env CGO_ENABLED  # Should output: 1
```

#### Database Locked Error

```bash
# Error: database is locked
# Solution: Close any open database connections

# Find processes using the database
lsof ./data/go-garage.db

# Or remove the lock (use with caution)
rm ./data/go-garage.db-shm ./data/go-garage.db-wal
```

#### Module Download Fails

```bash
# Error: cannot download modules
# Solution: Clear module cache

go clean -modcache
go mod download
```

#### Tests Fail with Permission Error

```bash
# Error: permission denied
# Solution: Ensure test data directory is writable

chmod -R 755 ./data
```

### Getting Help

1. **Check Documentation**:
   - [README.md](../README.md)
   - [CONTRIBUTING.md](../CONTRIBUTING.md)
   - [docs/DOCKER.md](DOCKER.md)
   - [docs/LOGGING.md](LOGGING.md)

2. **Review Specifications**:
   - [spec/milestone-1-project-setup.md](../spec/milestone-1-project-setup.md)

3. **Search Issues**: Check GitHub issues for similar problems

4. **Ask for Help**: Open a new issue with:
   - Clear description of the problem
   - Steps to reproduce
   - Environment details (OS, Go version)
   - Error messages or logs

## Development Tips

### Productivity Hints

1. **Use Make Commands**: Save time with `make build`, `make test`, etc.

2. **Watch Mode**: Use `air` or similar for live reloading during development
   ```bash
   go install github.com/cosmtrek/air@latest
   air
   ```

3. **Debug Logging**: Set `LOG_LEVEL=debug` in development

4. **Test Specific Packages**: Target tests to speed up feedback
   ```bash
   go test ./internal/handlers/...
   ```

5. **Use Docker for Clean Environment**: Test in Docker to catch environment-specific issues early

### Best Practices

- **Commit Often**: Make small, atomic commits
- **Run Tests Before Pushing**: Catch issues early
- **Keep Dependencies Updated**: Regularly run `go get -u` and `go mod tidy`
- **Use Feature Branches**: Never commit directly to main
- **Write Tests First**: Practice TDD when possible
- **Document As You Go**: Update docs when changing functionality

## Next Steps

Now that your environment is set up:

1. Review the [CONTRIBUTING.md](../CONTRIBUTING.md) guidelines
2. Check the current milestone in [spec/](../spec/)
3. Look for open issues labeled "good first issue"
4. Make your first contribution!

Happy coding! 🚗
