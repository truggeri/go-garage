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
- **Language**: Go (1.21+)
- **Web Framework**: Standard library `net/http` with gorilla/mux for routing
- **Database**: SQLite
- **ORM/Database Driver**: database/sql with appropriate drivers
- **Authentication**: JWT tokens
- **Configuration**: Environment variables and config files

### Frontend
- **Template Engine**: Go html/template
- **CSS Framework**: Bootstrap or Tailwind CSS
- **JavaScript**: Vanilla JS or lightweight framework (Alpine.js/htmx)

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

### User
- Authentication credentials
- Profile information
- Owned vehicles

## API Design

See [restful-api.md](./restful-api.md) for detailed API endpoint documentation.

## Security Considerations

- **Authentication**: JWT-based authentication for API access
- **Authorization**: Role-based access control (RBAC)
- **Data Validation**: Input validation at all entry points
- **SQL Injection Prevention**: Use parameterized queries
- **XSS Prevention**: Escape output in templates
- **CSRF Protection**: Implement CSRF tokens for forms

## Data Storage

### Database Schema

**Users Table**
- id (primary key)
- username (unique)
- email (unique)
- password_hash
- created_at, updated_at

**Vehicles Table**
- id (primary key)
- user_id (foreign key)
- vin (unique)
- make, model, year
- purchase_date, purchase_price
- status (active, sold, etc.)
- created_at, updated_at

**Maintenance Records Table**
- id (primary key)
- vehicle_id (foreign key)
- service_type
- service_date
- mileage
- cost
- provider
- notes
- created_at, updated_at

## Testing Strategy

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test database interactions and API endpoints
- **End-to-End Tests**: Test complete user workflows
- **Coverage Goal**: Minimum 80% code coverage
