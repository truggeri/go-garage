# Milestone 1: Project Setup and Core Infrastructure

## Objective
Establish the foundational project structure, development environment, and core infrastructure components needed to build the Go-Garage application.

## Goals

### 1. Project Initialization
- [x] Create GitHub repository
- [ ] Initialize Go module (`go mod init`)

### 2. Development Environment
- [ ] Create Makefile for common tasks (build, test, run, clean)
- [ ] Setup development Docker container
- [ ] Configure VSCode/GoLand settings
- [ ] Setup pre-commit hooks (gofmt, golint)

### 3. Project Structure
Create the following directory structure:
```
go-garage/
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers
│   ├── models/          # Data models
│   ├── repositories/    # Data access layer
│   ├── services/        # Business logic
│   └── middleware/      # HTTP middleware
├── pkg/                 # Public, reusable packages
├── web/
│   ├── static/          # CSS, JS, images
│   └── templates/       # HTML templates
├── migrations/          # Database migrations
├── spec/                # Project specifications
└── tests/               # Integration and E2E tests
```

### 4. Core Dependencies
Install and configure:
- [ ] gorilla/mux - HTTP router and URL matcher
- [ ] godotenv - Environment variable management
- [ ] golang-migrate - Database migration tool
- [ ] Database driver (mattn/go-sqlite3 for SQLite)
- [ ] testify - Testing toolkit

### 5. Configuration Management
- [ ] Create config package to load environment variables
- [ ] Support for .env files in development
- [ ] Configuration struct with validation
- [ ] Default values for all config options

Configuration items to include:
- Server port and host
- Database connection string
- Log level and format
- JWT secret key
- Environment name (dev, prod)

### 6. Basic HTTP Server
- [ ] Implement main.go in cmd/server/
- [ ] Setup gorilla/mux router
- [ ] Create health check endpoint (`/health`)
- [ ] Implement graceful shutdown
- [ ] Add request logging middleware
- [ ] Add panic recovery middleware

### 7. Logging
- [ ] Implement structured logging (using log/slog or logrus)
- [ ] Configure log levels (debug, info, warn, error)
- [ ] Add request/response logging
- [ ] Log rotation configuration

### 8. Docker Setup
- [ ] Create Dockerfile for the application
- [ ] Create docker-compose.yml for local development
- [ ] Volume configuration for persistence
- [ ] Network configuration

### 9. CI/CD Pipeline
- [ ] Create GitHub Actions workflow
- [ ] Run tests on pull requests
- [ ] Run linters (golangci-lint)
- [ ] Build Docker image
- [ ] Code coverage reporting

### 10. Documentation
- [ ] Update README with setup instructions
- [ ] Document environment variables
- [ ] Add contributing guidelines
- [ ] Create development setup guide

## Deliverables

1. **Runnable Application**: A basic web server that responds to HTTP requests
2. **Health Check Endpoint**: `/health` endpoint returning application status
3. **Configuration System**: Environment-based configuration management
4. **Docker Setup**: Containerized application ready for deployment
5. **CI Pipeline**: Automated testing and linting on code changes
6. **Documentation**: Clear setup and development instructions

## Success Criteria

- [ ] Application starts without errors
- [ ] Health check endpoint returns 200 OK
- [ ] Configuration loads from environment variables
- [ ] Docker container builds and runs successfully
- [ ] CI pipeline passes all checks
- [ ] Code passes all linters (gofmt, go vet, golangci-lint)
- [ ] Development documentation is complete and accurate

## Dependencies
None

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Dependency version conflicts | Medium | Use Go modules with specific versions, maintain go.sum |
| Complex setup process | Low | Provide clear documentation and Makefile commands |
| CI/CD configuration issues | Medium | Test locally with act or similar tools before committing |

## Notes
- Focus on simplicity and maintainability
- Follow Go best practices and idioms
- Keep the initial setup minimal but extensible
- Ensure all team members can run the project locally
