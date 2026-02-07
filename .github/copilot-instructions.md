# Copilot Instructions for Go-Garage

## Project Overview

Go-Garage is a vehicle management web application written in Go that helps users track and manage their vehicles, maintenance records, and fuel consumption. The project follows Go best practices and emphasizes simplicity, scalability, and reliability.

## Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: Standard library `net/http` with `gorilla/mux` for routing
- **Database**: SQLite with `database/sql` and appropriate drivers
- **Frontend**: Go `html/template` with htmx for interactivity
- **Authentication**: JWT tokens
- **Testing**: `testify` testing toolkit
- **CI/CD**: GitHub Actions
- **Containerization**: Docker and Docker Compose

## Code Style and Conventions

### Go Best Practices
- Follow standard Go idioms and conventions
- Use `gofmt` for code formatting
- Run `go vet` and `golangci-lint` before committing
- Keep functions small and focused on a single responsibility
- Use meaningful variable and function names
- Write clear error messages with context

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
├── spec/                # Project specifications
└── tests/               # Integration and E2E tests
```

### Naming Conventions
- Use `PascalCase` for exported types, functions, and methods
- Use `camelCase` for unexported types, functions, and methods
- Use descriptive names: `vehicleRepository` not `vr`
- Interface names should describe behavior: `VehicleRepository`, `MaintenanceService`
- Test functions should follow pattern: `TestFunctionName_Scenario_ExpectedBehavior`

### Error Handling
- Always handle errors explicitly - never ignore them
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Return errors as the last return value
- Use custom error types for domain-specific errors
- Log errors at appropriate levels before returning

### Testing
- Write unit tests for all business logic
- Use table-driven tests where appropriate
- Aim for minimum 80% code coverage
- Test error cases and edge conditions
- Use `testify` assertions for clear test failures
- Mock external dependencies using interfaces

### Database
- Always use parameterized queries to prevent SQL injection
- Use transactions for operations that modify multiple tables
- Close database connections and rows properly (use `defer`)
- Handle database errors gracefully
- Use repository pattern for data access

### Security
- Validate all user input at entry points
- Escape output in HTML templates to prevent XSS
- Use JWT tokens for authentication
- Never commit secrets or API keys
- Use HTTPS in production
- Implement CSRF protection for forms
- Use prepared statements for all SQL queries

### API Design
- Follow RESTful conventions
- Use appropriate HTTP methods (GET, POST, PUT, DELETE)
- Return proper HTTP status codes
- Use JSON for request/response bodies
- Implement proper error responses with meaningful messages
- Version the API if necessary (e.g., `/api/v1/`)

### Documentation
- Add package comments for all packages
- Document exported functions, types, and methods
- Use GoDoc format for documentation
- Keep README.md up to date with setup instructions
- Document environment variables and configuration options
- Add inline comments for complex logic only

### Configuration
- Use environment variables for configuration
- Support `.env` files for local development
- Provide sensible defaults
- Validate configuration on startup
- Document all configuration options

### Logging
- Use structured logging (preferably `log/slog`)
- Log at appropriate levels (debug, info, warn, error)
- Include context in log messages
- Don't log sensitive information (passwords, tokens, PII)
- Use request IDs for tracing requests through the system

### Code Organization
- Keep related functionality in the same package
- Use dependency injection for better testability
- Avoid circular dependencies
- Keep the `internal/` package private
- Export only what's necessary from packages

### Build and Development
- Use `make` commands for common tasks (build, test, run, clean)
- Ensure code builds without warnings
- Run tests before committing
- Use Docker for consistent development environments
- Document setup steps in README.md

## Domain Models

### Core Entities
- **Vehicle**: Tracks vehicle information (VIN, make, model, year, owner)
- **Maintenance Record**: Tracks service history (date, type, mileage, cost)
- **Fuel Record**: Tracks fuel consumption (date, mileage, price, volume, location)
- **User**: User authentication and profile information

### Relationships
- One User can own multiple Vehicles
- One Vehicle can have multiple Maintenance Records
- One Vehicle can have multiple Fuel Records

## Development Workflow

1. **Before Making Changes**
   - Review relevant specification documents in `/spec/`
   - Understand the milestone goals and acceptance criteria
   - Check existing tests to understand expected behavior

2. **Writing Code**
   - Follow the project structure
   - Write tests alongside code (TDD when possible)
   - Run linters and formatters
   - Keep changes focused and minimal

3. **Testing**
   - Run unit tests: `go test ./...`
   - Run linters: `golangci-lint run`
   - Format code: `gofmt -s -w .`
   - Check test coverage: `go test -cover ./...`

4. **Committing**
   - Write clear commit messages
   - Reference issue numbers when applicable
   - Ensure CI/CD pipeline passes

## Common Tasks

### Adding a New API Endpoint
1. Define the route in the router
2. Create a handler in `internal/handlers/`
3. Implement business logic in `internal/services/`
4. Add repository methods if database access is needed
5. Write tests for handler, service, and repository
6. Update API documentation

### Adding a New Database Table
1. Design the schema and relationships
2. Create a migration in `migrations/`
3. Define the model in `internal/models/`
4. Create repository interface and implementation
5. Write tests for repository methods
6. Update data schema documentation

## References

- [Project Specifications](/spec/README.md)
- [Architecture Document](/spec/architecture.md)
- [API Documentation](/spec/restful-api.md)
- [Data Schema](/spec/data-schema.md)
- [Current Milestone](/spec/milestone-1-project-setup.md)

## Questions?

Refer to the specification documents in the `/spec/` directory for detailed requirements and design decisions.
