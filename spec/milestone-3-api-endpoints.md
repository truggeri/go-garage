# Milestone 3: RESTful API Endpoints

## Objective

Implement a complete RESTful API for managing vehicles and maintenance records, with proper authentication, authorization, and error handling.

## Prerequisites

- Milestone 1: Project Setup and Core Infrastructure
- Milestone 2: Vehicle Data Model and Database Layer

## Goals

### 1. Service Layer Implementation

#### Vehicle Service

- [ ] Create VehicleService interface
- [ ] Implement CreateVehicle(ctx, vehicle)
- [ ] Implement GetVehicle(ctx, id)
- [ ] Implement GetUserVehicles(ctx, userID)
- [ ] Implement UpdateVehicle(ctx, id, updates)
- [ ] Implement ArchiveVehicle(ctx, id)
- [ ] Implement ListVehicles(ctx, filters, pagination)
- [ ] Implement ownership verification

#### Maintenance Service

- [ ] Create MaintenanceService interface
- [ ] Implement CreateMaintenance(ctx, record)
- [ ] Implement GetMaintenance(ctx, id)
- [ ] Implement GetVehicleMaintenance(ctx, vehicleID)
- [ ] Implement UpdateMaintenance(ctx, id, updates)
- [ ] Implement DeleteMaintenance(ctx, id)

#### User Service

- [ ] Create UserService interface
- [ ] Implement CreateUser(ctx, user)
- [ ] Implement GetUser(ctx, id)
- [ ] Implement UpdateUser(ctx, id, updates)
- [ ] Implement password hashing (bcrypt)

### 2. Authentication & Authorization

#### JWT Implementation

- [ ] Create JWT token generation
- [ ] Create JWT token validation
- [ ] Implement token refresh mechanism
- [ ] Set appropriate token expiration
- [ ] Store claims (user ID, username, roles)

#### Middleware

- [ ] Create authentication middleware
- [ ] Validate JWT tokens
- [ ] Extract user context from token
- [ ] Handle authentication errors
- [ ] Create authorization middleware
- [ ] Verify resource ownership
- [ ] Implement role-based access control (if needed)

#### Authentication Endpoints

- [ ] POST /api/v1/auth/register - User registration
- [ ] POST /api/v1/auth/login - User login (returns JWT)
- [ ] POST /api/v1/auth/refresh - Refresh JWT token
- [ ] POST /api/v1/auth/logout - Logout (invalidate token)

### 3. Vehicle API Endpoints

- [ ] GET /api/v1/vehicles
  - List all vehicles for authenticated user
  - Support pagination (page, limit)
  - Support filtering (make, model, year, status)
  - Support sorting

- [ ] POST /api/v1/vehicles
  - Create new vehicle
  - Validate input
  - Return created vehicle with ID

- [ ] GET /api/v1/vehicles/{id}
  - Get specific vehicle details
  - Verify ownership
  - Return 404 if not found

- [ ] PUT /api/v1/vehicles/{id}
  - Update vehicle details
  - Verify ownership
  - Validate updates
  - Return updated vehicle

- [ ] DELETE /api/v1/vehicles/{id}
  - Delete vehicle
  - Verify ownership
  - Cascade delete maintenance records (or restrict)

- [ ] GET /api/v1/vehicles/{id}/stats
  - Get vehicle statistics (total maintenance cost, etc.)

### 4. Maintenance API Endpoints

- [x] GET /api/v1/vehicles/{vehicleId}/maintenance
  - List maintenance records for a vehicle
  - Support pagination
  - Support date range filtering
  - Support sorting by date

- [x] POST /api/v1/vehicles/{vehicleId}/maintenance
  - Create maintenance record
  - Verify vehicle ownership
  - Validate input

- [x] GET /api/v1/maintenance/{id}
  - Get specific maintenance record
  - Verify ownership through vehicle

- [x] PUT /api/v1/maintenance/{id}
  - Update maintenance record
  - Verify ownership
  - Validate updates

- [x] DELETE /api/v1/maintenance/{id}
  - Delete maintenance record
  - Verify ownership

### 5. User API Endpoints

- [x] GET /api/v1/users/me
  - Get current user profile
  - Return user info without password

- [x] PUT /api/v1/users/me
  - Update current user profile
  - Allow username, email, name updates
  - Re-authenticate for sensitive changes

- [x] PUT /api/v1/users/me/password
  - Change user password
  - Require current password
  - Validate new password strength

- [x] DELETE /api/v1/users/me
  - Delete user account
  - Require password confirmation
  - Cascade delete vehicles and maintenance

### 6. Request/Response Handling

#### Request Parsing

- [ ] Implement JSON request body parsing
- [ ] Validate content-type headers
- [ ] Implement request size limits
- [ ] Handle URL parameters and query strings
- [ ] Implement request validation middleware

#### Response Formatting

- [ ] Standardize JSON response format

```json
{
  "success": true,
  "data": {...},
  "message": "Optional message",
  "errors": []
}
```

- [ ] Set appropriate HTTP status codes
- [ ] Implement pagination metadata

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5
  }
}
```

### 7. Error Handling

- [ ] Create standardized error responses
- [ ] Map internal errors to HTTP status codes
- [ ] Implement error logging
- [ ] Don't expose internal errors to clients
- [ ] Return validation errors with field details

Error response format:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "details": [
      {"field": "email", "message": "Invalid email format"}
    ]
  }
}
```

### 8. API Middleware Stack

- [ ] Request logging
- [ ] CORS headers
- [ ] Request ID generation
- [ ] Request timeout
- [ ] Panic recovery
- [ ] Security headers (X-Content-Type-Options, etc.)

### 9. Input Validation

- [ ] Validate all input fields
- [ ] Implement custom validators
- [ ] Validate ID formats (UUID/integer)
- [ ] Validate pagination parameters
- [ ] Sanitize user input

### 10. API Testing

#### Unit Tests

- [ ] Test all service methods
- [ ] Test authentication logic
- [ ] Test authorization logic
- [ ] Mock repositories

#### Integration Tests

- [ ] Test all API endpoints
- [ ] Test authentication flows
- [ ] Test authorization (access control)
- [ ] Test error scenarios
- [ ] Test pagination
- [ ] Test filtering and sorting

#### API Documentation

- [ ] Create OpenAPI specification
- [ ] Document all endpoints
- [ ] Document request/response schemas
- [ ] Document authentication requirements
- [ ] Provide example requests/responses

## Deliverables

1. **Service Layer**: Complete business logic implementation
2. **API Endpoints**: All RESTful endpoints implemented
3. **Authentication**: JWT-based authentication system
4. **Authorization**: Resource ownership verification
5. **API Documentation**: Complete OpenAPI specification
6. **Tests**: Comprehensive unit and integration tests
7. **Postman Collection**: API testing collection (optional)

## Success Criteria

- [ ] All API endpoints return correct status codes
- [ ] Authentication is required for protected endpoints
- [ ] Users can only access their own resources
- [ ] Input validation prevents invalid data
- [ ] Error responses are consistent and informative
- [ ] API tests achieve >80% coverage
- [ ] API documentation is complete and accurate

## Dependencies

- Milestone 1: Project Setup and Core Infrastructure
- Milestone 2: Vehicle Data Model and Database Layer

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Security vulnerabilities | High | Security review, use established libraries, input validation |
| Performance issues | Medium | Implement caching, optimize queries, load testing |
| API design changes | Medium | Version API (v1), maintain backward compatibility |
| Authentication complexity | Medium | Use proven JWT libraries, follow best practices |
| Authorization bugs | High | Thorough testing, code review, consistent patterns |

## Notes

- Follow RESTful conventions
- Use standard HTTP status codes appropriately
- Keep endpoints simple and focused
- Consider API versioning from the start
