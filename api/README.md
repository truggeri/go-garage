# Go-Garage API Specification

This directory contains the OpenAPI specification for the Go-Garage API.

## Files

- `openapi.yaml` - OpenAPI 3.0.3 specification for the entire API

## API Overview

The Go-Garage API provides endpoints for:

- **Authentication** - User registration, login, token refresh, and logout
- **Vehicles** - CRUD operations for vehicle management
- **Maintenance Records** - Track and manage vehicle maintenance history
- **Fuel Records** - Track and manage fuel consumption
- **User Profile** - Manage user account information

All endpoints except authentication require a valid JWT token in the `Authorization` header.

## Base URLs

- **Development**: http://localhost:8080
- **Production**: https://api.go-garage.example.com (update with your actual production URL)

## Authentication

1. Register a new account: `POST /api/v1/auth/register`
2. Login: `POST /api/v1/auth/login` (returns access and refresh tokens)
3. Include the access token in all subsequent requests:
   ```
   Authorization: Bearer <access_token>
   ```
4. Refresh the access token when it expires: `POST /api/v1/auth/refresh`

## Response Format

All successful responses follow this format:

```json
{
  "success": true,
  "data": { ... },
  "message": "Optional message"
}
```

Error responses follow this format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error message",
    "details": [
      {
        "field": "field_name",
        "message": "Field error message"
      }
    ]
  }
}
```

## HTTP Status Codes

- `200 OK` - Request succeeded
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid input or validation error
- `401 Unauthorized` - Authentication required or failed
- `403 Forbidden` - User does not have permission to access the resource
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists (e.g., duplicate username/email)
- `500 Internal Server Error` - Server error

## Pagination

List endpoints support pagination with the following query parameters:

- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)

Paginated responses include pagination metadata:

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```
