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

```
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

## Testing Strategy

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test database interactions and API endpoints
- **End-to-End Tests**: Test complete user workflows
- **Coverage Goal**: Minimum 80% code coverage
