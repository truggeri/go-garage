# Go-Garage Application Architecture

## Overview

Go-Garage is a vehicle management web application built with Go, designed to help users track and manage their vehicles, maintenance records, and related information.

## Architecture Principles

- **Simplicity**: Focus on clean, maintainable code following Go best practices
- **Scalability**: Design for growth with modular components
- **Reliability**: Implement robust error handling and data validation
- **Performance**: Leverage Go's concurrency features for efficient operations

## System Architecture

### High-Level Components

```text
┌─────────────────────────────────────────────────────────┐
│                     Web Browser                         │
└─────────────────────┬───────────────────────────────────┘
                      │
                      │ HTTP/HTTPS
                      │
┌─────────────────────▼───────────────────────────────────┐
│                  Go Web Server                          │
│                                                         │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Router    │  │  Middleware  │  │   Handlers   │  │
│  │  (HTTP Mux) │  │  (Auth, Log) │  │  (API/View)  │  │
│  └─────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────┬───────────────────────────────────┘
                      │
                      │
┌─────────────────────▼───────────────────────────────────┐
│                 Business Logic Layer                    │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Vehicle    │  │ Maintenance  │  │     User     │  │
│  │   Service    │  │   Service    │  │   Service    │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────┬───────────────────────────────────┘
                      │
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   Data Layer                            │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  Repository  │  │   Database   │  │    Cache     │  │
│  │  Interface   │  │   (SQLite)   │  │  (Optional)  │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Technology Stack

### Backend

- **Language**: Go (1.24+)
- **Web Framework**: Standard library `net/http` with gorilla/mux for routing
- **Database**: SQLite
- **ORM/Database Driver**: database/sql with appropriate drivers
- **Authentication**: JWT tokens
- **Configuration**: Environment variables and config files

### Frontend

- **Template Engine**: Go html/template
- **CSS Framework**: To be determined
- **JavaScript**: [htmx](https://htmx.org)

### Infrastructure

- **Containerization**: Docker
- **Deployment**: Docker Compose
- **CI/CD**: GitHub Actions

## Core Domain Models

### Vehicle

- Identification (VIN, make, model, year)
- Owner information
- Purchase details
- Current status

### Maintenance Record

- Service date and type
- Mileage at service
- Cost and provider
- Notes and attachments

### Fuel Record

- Date and time
- Mileage at fill up
- Price
- Volume (gallons)
- Percentage city driving
- Octane rating
- Location
- Brand
- Notes and attachments

### User

- Authentication credentials
- Profile information
- Owned vehicles

## API Design

See [the open api spec](./openapi.yaml) for detailed API endpoint documentation.

## Security Considerations

- **Authentication**: JWT-based authentication for API access
- **Authorization**: Role-based access control (RBAC)
- **Data Validation**: Input validation at all entry points
- **SQL Injection Prevention**: Use parameterized queries
- **XSS Prevention**: Escape output in templates
- **CSRF Protection**: Implement CSRF tokens for forms

## Data Storage

See [data-schema.md](./data-schema.md) for detailed database schema documentation.

## Layer Responsibilities

Each layer has specific responsibilities. Code must not cross these boundaries.

### Handlers (`internal/handlers/`)

Handlers are the HTTP boundary. They are responsible for:

- Parsing HTTP requests (path params, query params, form values, JSON body)
- Converting raw string inputs into typed values (e.g. `strconv.Atoi`)
- Calling the appropriate service method
- Mapping service results/errors to HTTP status codes and response bodies
- Rendering templates (page handlers) or encoding JSON (API handlers)

Handlers must **not** contain:

- Business logic (calculations, sorting, aggregations, statistics)
- Direct database access or repository calls
- Domain validation rules (use model validators instead)
- Data transformation beyond request/response mapping

### Services (`internal/services/`)

Services contain all business logic. They are responsible for:

- Orchestrating operations across one or more repositories
- Enforcing business rules and invariants
- Computing derived data (statistics, aggregations, sorting)
- Coordinating multi-step operations and transactions
- Ownership verification and authorization logic

Services must **not** contain:

- HTTP-specific concerns (status codes, headers, request parsing)
- SQL queries or direct database access
- Presentation/formatting logic

### Models (`internal/models/`)

Models define domain types and validation. They are responsible for:

- Defining domain structs (Vehicle, User, MaintenanceRecord, etc.)
- All validation rules (`ValidateVehicle`, `ValidateUser`, etc.)
- Domain constants and enums (VehicleStatus, etc.)
- Custom domain error types

Models must **not** contain:

- Business logic or service orchestration
- HTTP or database concerns
- Presentation/formatting logic

### Repositories (`internal/repositories/`)

Repositories handle data persistence. They are responsible for:

- Defining repository interfaces
- Implementing CRUD operations against the database
- Translating between database rows and domain models
- Query building with filters and pagination

Repositories must **not** contain:

- Business logic or validation
- HTTP concerns
- Cross-repository orchestration (use services for that)

## File Organization

### File Size Limits

- **Target**: 100–150 lines per file (excluding tests)
- **Hard limit**: No file should exceed 200 lines (excluding tests)
- Test files may be longer due to table-driven tests

### Splitting Guidelines

When a file approaches or exceeds 200 lines, split by responsibility:

- **Handlers**: One file per resource action group (e.g. `page_vehicle_list.go`, `page_vehicle_detail.go`, `page_vehicle_create.go`)
- **Helpers**: Extract request parsing, response building, and form parsing into `*_helpers.go` or `*_form.go` files
- **Data structs**: Page data structs can live alongside their handler or in a dedicated `*_types.go` file
- **Services**: Split large services into focused files (e.g. `vehicle.go` for interface + core, `vehicle_stats.go` for statistics)
- **Repositories**: Split by operation category if needed (e.g. `vehicle_sqlite_read.go`, `vehicle_sqlite_write.go`)

### Naming Conventions for Files

| Pattern | Purpose | Example |
|---------|---------|---------|
| `page_<resource>_<action>.go` | Page handler for one action | `page_vehicle_list.go` |
| `page_<resource>_form.go` | Form parsing helpers | `page_vehicle_form.go` |
| `<resource>.go` | API handler | `vehicle.go` |
| `<resource>_helpers.go` | Request/response mapping helpers | `vehicle_helpers.go` |
| `<resource>_response.go` | JSON response builders | `vehicle_response.go` |

## Testing Strategy

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test database interactions and API endpoints
- **End-to-End Tests**: Test complete user workflows
- **Coverage Goal**: Minimum 80% code coverage
