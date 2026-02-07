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

```bash
# Start the application
docker compose up -d

# View logs
docker compose logs -f

# Stop the application
docker compose down
```

The application will be available at http://localhost:8080

### Local Development

#### Prerequisites

- Go 1.24 or later
- SQLite3
- Make (optional, for convenience commands)

#### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/truggeri/go-garage.git
   cd go-garage
   ```

2. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Build the application:
   ```bash
   make build
   # or
   go build -o bin/go-garage ./cmd/server
   ```

5. Run the application:
   ```bash
   make run
   # or
   ./bin/go-garage
   ```

## Configuration

Configuration is managed through environment variables. See `.env.example` for available options.

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `development` | Application environment (development/production) |
| `APP_PORT` | `8080` | Port the server listens on |
| `SERVER_HOST` | `0.0.0.0` | Host the server binds to |
| `LOG_LEVEL` | `info` | Logging level (debug/info/warn/error) |
| `LOG_FORMAT` | `json` | Log output format (json/text) |
| `DB_PATH` | `./data/go-garage.db` | Path to SQLite database file |
| `JWT_SECRET` | _(required in production)_ | JWT secret key |

## Development

### Available Make Commands

```bash
make build    # Build the application
make test     # Run tests with coverage
make run      # Run the application
make clean    # Clean build artifacts
make fmt      # Format code
make lint     # Run linters
make vet      # Run go vet
```

### Project Structure

```
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
├── web/                 # Frontend assets
│   ├── static/          # CSS, JS, images
│   └── templates/       # HTML templates
├── migrations/          # Database migrations
├── docs/                # Documentation
├── spec/                # Project specifications
└── tests/               # Integration and E2E tests
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Run go vet
make vet
```

## Docker

For detailed Docker setup and usage instructions, see [docs/DOCKER.md](docs/DOCKER.md).

### Building Docker Image

```bash
docker build -t go-garage .
```

### Running with Docker Compose

```bash
# Start services
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down
```

## API Endpoints

- `GET /health` - Health check endpoint

*(More endpoints will be added as the application develops)*

## Documentation

- [Docker Setup Guide](docs/DOCKER.md)
- [Logging Documentation](docs/LOGGING.md)
- [Project Specifications](spec/)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linters
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with:
- [gorilla/mux](https://github.com/gorilla/mux) - HTTP router
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver
- [golang-migrate](https://github.com/golang-migrate/migrate) - Database migrations
- [godotenv](https://github.com/joho/godotenv) - Environment variable management
- [testify](https://github.com/stretchr/testify) - Testing toolkit
