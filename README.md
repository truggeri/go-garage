# Go-Garage

A vehicle management application written in Go that helps users track and manage their vehicles, maintenance records, and fuel consumption.

## Features

- Vehicle tracking and management
- Maintenance record history
- Fuel consumption tracking
- RESTful API
- SQLite database for data persistence
- Structured logging
- Docker support

## Quick Start

### Using Docker (Recommended)

```shell
# Start the application
docker compose up -d

# View logs
docker compose logs -f

# Stop the application
docker compose down
```

The application will be available at <http://localhost:8080>

### Local Development

#### Prerequisites

- Go 1.24 or later
- SQLite3
- Make (optional, for convenience commands)

#### Setup

1. Clone the repository:

   ```shell
   git clone https://github.com/truggeri/go-garage.git
   cd go-garage
   ```

2. Copy the example environment file:

   ```shell
   cp .env.example .env
   ```

3. Install dependencies:

   ```shell
   go mod download
   ```

4. Build the application:

   ```shell
   make build
   # or
   go build -o bin/go-garage ./cmd/server
   ```

5. Run the application:

   ```shell
   make run
   # or
   ./bin/go-garage
   ```

## Configuration

Configuration is managed through environment variables. See `.env.example` for available options.

| Variable | Default | Description |
| ---------- | --------- | ------------- |
| `ENVIRONMENT` | `development` | Application environment (development/production) |
| `APP_PORT` | `8080` | Port the server listens on |
| `SERVER_HOST` | `0.0.0.0` | Host the server binds to |
| `LOG_LEVEL` | `info` | Logging level (debug/info/warn/error) |
| `LOG_FORMAT` | `json` | Log output format (json/text) |
| `DB_PATH` | `./data/go-garage.db` | Path to SQLite database file |
| `JWT_SECRET` | _(required in production)_ | JWT secret key |

## Development

### Available Make Commands

```shell
make build    # Build the application
make test     # Run tests with coverage
make run      # Run the application
make clean    # Clean build artifacts
make fmt      # Format code
make lint     # Run linters
make vet      # Run go vet
```

### Project Structure

```text
go-garage/
├── cmd/server/          # Main application entry point
├── internal/            # Private application code
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers
│   ├── models/          # Data models
│   ├── repositories/    # Data access layer
│   ├── services/        # Business logic
│   └── middleware/      # HTTP middleware
├── pkg/                 # Public, reusable packages
├── api/                 # API specification (OpenAPI)
├── web/                 # Frontend assets
│   ├── static/          # CSS, JS, images
│   └── templates/       # HTML templates
├── migrations/          # Database migrations
├── docs/                # Documentation
├── spec/                # Project specifications
└── tests/               # Integration and E2E tests
```

### Running Tests

```shell
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Code Quality

```shell
# Format code
make fmt

# Run linters
make lint

# Run go vet
make vet
```

### Continuous Integration

The project uses GitHub Actions for automated testing and quality checks. Every pull request and push to main triggers:

- **Code Quality Checks**
  - Go module tidiness verification
  - Code formatting validation (gofmt)
  - Static analysis (go vet)
  - Comprehensive linting (golangci-lint)
  - Test suite execution with race detection
  - Code coverage reporting to Codecov

- **Docker Validation**
  - Multi-stage Docker image build
  - Container health check verification
  - Image structure validation

All CI checks must pass before merging pull requests. View the workflow configuration in `.github/workflows/go-garage-quality.yml`.

## Docker

For detailed Docker setup and usage instructions, see [docs/DOCKER.md](docs/DOCKER.md).

### Building Docker Image

```shell
docker build -t go-garage .
```

### Running with Docker Compose

```shell
# Start services
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down
```

## API Endpoints

The Go-Garage API provides comprehensive endpoints for vehicle management. For complete API documentation, see the [OpenAPI specification](spec/openapi.yaml).

### Quick Reference

#### Authentication

- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login and receive JWT tokens
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout

#### Vehicles

- `GET /api/v1/vehicles` - List all user's vehicles
- `POST /api/v1/vehicles` - Create a new vehicle
- `GET /api/v1/vehicles/{id}` - Get vehicle details
- `PUT /api/v1/vehicles/{id}` - Update vehicle
- `DELETE /api/v1/vehicles/{id}` - Delete vehicle
- `GET /api/v1/vehicles/{id}/stats` - Get vehicle statistics

#### Maintenance Records

- `GET /api/v1/vehicles/{vehicleId}/maintenance` - List maintenance records
- `POST /api/v1/vehicles/{vehicleId}/maintenance` - Create maintenance record
- `GET /api/v1/maintenance/{id}` - Get maintenance record
- `PUT /api/v1/maintenance/{id}` - Update maintenance record
- `DELETE /api/v1/maintenance/{id}` - Delete maintenance record

#### Fuel Records

- `GET /api/v1/vehicles/{vehicleId}/fuel` - List fuel records
- `POST /api/v1/vehicles/{vehicleId}/fuel` - Create fuel record
- `GET /api/v1/fuel/{id}` - Get fuel record
- `PUT /api/v1/fuel/{id}` - Update fuel record
- `DELETE /api/v1/fuel/{id}` - Delete fuel record

#### User Profile

- `GET /api/v1/users/me` - Get current user profile
- `PUT /api/v1/users/me` - Update user profile
- `DELETE /api/v1/users/me` - Delete user account
- `PUT /api/v1/users/me/password` - Change password

#### Health

- `GET /health` - Health check endpoint

For detailed request/response schemas, authentication requirements, and examples, refer to the [API documentation](spec/openapi.yaml).

## Documentation

- [API Specification](spec/openapi.yaml) - Complete OpenAPI specification
- [Development Setup Guide](docs/DEVELOPMENT.md) - Comprehensive guide for setting up your development environment
- [Docker Setup Guide](docs/DOCKER.md) - Docker installation and usage instructions
- [Logging Documentation](docs/LOGGING.md) - Logging configuration and best practices
- [Contributing Guidelines](CONTRIBUTING.md) - How to contribute to the project
- [Project Specifications](spec/) - Detailed project specifications and milestones

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on:

- Setting up your development environment
- Coding standards and best practices
- Testing requirements
- Pull request process
- Reporting issues

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with:

- [gorilla/mux](https://github.com/gorilla/mux) - HTTP router
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver
- [golang-migrate](https://github.com/golang-migrate/migrate) - Database migrations
- [godotenv](https://github.com/joho/godotenv) - Environment variable management
- [testify](https://github.com/stretchr/testify) - Testing toolkit
