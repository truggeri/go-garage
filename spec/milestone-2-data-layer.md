# Milestone 2: Vehicle Data Model and Database Layer

## Objective
Implement the core data models and database layer for vehicle management, including database schema, repositories, and basic CRUD operations.

## Prerequisites
- Milestone 1 completed (project setup and infrastructure)

## Goals

### 1. Database Setup
- [x] Configure database connection with connection pooling (PR #24)
- [x] Implement health check for database connectivity (PR #24)
- [x] Setup database migrations system (PR #24)
- [x] Create initial migration files (PR #24)

### 2. Data Models

#### Vehicle Model
Define the Vehicle struct with the following fields:
- [x] ID (UUID) (PR #26)
- [x] User ID (foreign key) (PR #26)
- [x] VIN (Vehicle Identification Number) (PR #26)
- [x] Make (PR #26)
- [x] Model (PR #26)
- [x] Year (PR #26)
- [x] Color (PR #26)
- [x] License plate (PR #26)
- [x] Purchase date (PR #26)
- [x] Purchase price (PR #26)
- [x] Purchase mileage (PR #26)
- [x] Current mileage (PR #26)
- [x] Status (active, sold, scrapped) (PR #26)
- [x] Notes (PR #26)
- [x] Created at timestamp (PR #26)
- [x] Updated at timestamp (PR #26)

#### Maintenance Record Model
Define the MaintenanceRecord struct:
- [x] ID (PR #26)
- [x] Vehicle ID (foreign key) (PR #26)
- [x] Service type (oil change, tire rotation, inspection, etc.) (PR #26)
- [x] Service date (PR #26)
- [x] Mileage at service (PR #26)
- [x] Cost (PR #26)
- [x] Service provider (PR #26)
- [x] Notes/description (PR #26)
- [x] Created at timestamp (PR #26)
- [x] Updated at timestamp (PR #26)

#### User Model
Define the User struct:
- [x] ID (PR #26)
- [x] Username (unique) (PR #26)
- [x] Email (unique) (PR #26)
- [x] Password hash (PR #26)
- [x] First name (PR #26)
- [x] Last name (PR #26)
- [x] Created at timestamp (PR #26)
- [x] Updated at timestamp (PR #26)
- [x] Last login timestamp (PR #26)

### 3. Database Schema
- [x] Create users table migration (PR #24)
- [x] Create vehicles table migration (PR #24)
- [x] Create maintenance_records table migration (PR #24)
- [x] Add indexes for foreign keys (PR #24)
- [x] Add indexes for frequently queried fields (VIN, email, username) (PR #24)
- [x] Add unique constraints where needed (PR #24)

### 4. Repository Interface Design
Create repository interfaces for:
- [x] VehicleRepository interface (PR #26)
  - Create(vehicle)
  - FindByID(id)
  - FindByUserID(userID)
  - FindByVIN(vin)
  - Update(vehicle)
  - Delete(id)
  - List(filters, pagination)

- [x] MaintenanceRepository interface (PR #26)
  - Create(record)
  - FindByID(id)
  - FindByVehicleID(vehicleID)
  - Update(record)
  - Delete(id)
  - List(filters, pagination)

- [x] UserRepository interface (PR #26)
  - Create(user)
  - FindByID(id)
  - FindByEmail(email)
  - FindByUsername(username)
  - Update(user)
  - Delete(id)

### 5. Repository Implementation
- [x] Implement SQLite repository for vehicles (PR #26)
- [x] Implement SQLite repository for maintenance records (PR #26)
- [x] Implement SQLite repository for users (PR #26)
- [x] Use prepared statements for all queries (PR #26)
- [x] Implement transaction support (PR #26 - via SQLite connection)
- [x] Add context support for cancellation (PR #26)

### 6. Data Validation
- [x] Implement validation for Vehicle model (PR #26)
  - VIN format validation (17 characters)
  - Year range validation (1900-present)
  - Required fields validation
  - Price and mileage validation (non-negative)

- [x] Implement validation for MaintenanceRecord (PR #26)
  - Date validation (not in future)
  - Cost validation (non-negative)
  - Required fields validation

- [x] Implement validation for User (PR #26)
  - Email format validation
  - Username format validation (alphanumeric, minimum length)
  - Password strength requirements

### 7. Error Handling
- [x] Define custom error types (NotFoundError, ValidationError, etc.) (PR #26)
- [x] Implement error wrapping with context (PR #26)
- [x] Create error helper functions (PR #26)
- [x] Document expected errors for each repository method (PR #26 - inline code documentation)

### 8. Testing

#### Unit Tests
- [x] Test all repository methods (PR #26)
- [x] Test data validation functions (PR #26)
- [x] Test error handling paths (PR #26)
- [x] Mock database for unit tests (PR #26 - uses test utilities)

#### Integration Tests
- [x] Test with test database (PR #26)
- [x] Test CRUD operations end-to-end (PR #26)
- [ ] Test concurrent access scenarios
- [ ] Test transaction rollback scenarios
- [x] Test database constraints (unique, foreign keys) (PR #26)

### 9. Database Utilities
- [ ] Create database seeding script for development
- [ ] Create sample data for testing
- [ ] Implement database backup helper
- [ ] Create migration rollback procedures

### 10. Documentation
- [ ] Document database schema with ERD
- [ ] Document repository interfaces
- [ ] Document validation rules
- [x] Add inline code documentation (PR #26)
- [ ] Create database setup guide

## Deliverables

1. **Database Schema**: Complete schema with all tables, indexes, and constraints ✅
2. **Repository Layer**: Full implementation of all repository interfaces ✅
3. **Data Models**: Validated Go structs for all domain entities ✅
4. **Migrations**: Up and down migrations for schema changes ✅
5. **Tests**: Comprehensive unit and integration tests (>80% coverage) ✅ (77.4% repository, 100% models)
6. **Documentation**: Complete API documentation for repositories ⏳ (inline docs done, dedicated docs pending)

## Success Criteria

- [x] All migrations run successfully (PR #24)
- [x] All repository tests pass (PR #26)
- [ ] Code coverage >80% for repository layer (77.4% repository, 100% models - needs improvement)
- [x] Database constraints enforce data integrity (PR #24, #26)
- [ ] Queries are efficient (use EXPLAIN ANALYZE)
- [x] No SQL injection vulnerabilities (prepared statements used throughout, PR #26)
- [ ] Concurrent access is handled correctly (needs dedicated tests)
- [x] All validation rules work as expected (PR #26)

## Dependencies
- Milestone 1: Project Setup and Core Infrastructure

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Database migration failures | High | Test migrations thoroughly, implement rollback procedures |
| Performance issues with queries | Medium | Add proper indexes, use query analysis tools |
| Data validation complexity | Medium | Use existing validation libraries (validator/v10) |
| Concurrent access issues | Medium | Use database transactions, test concurrent scenarios |
| Schema changes during development | Low | Use versioned migrations, maintain backward compatibility |

## Notes
- Use database/sql standard library for maximum flexibility
- Keep repository implementations simple and focused
- Consider using GORM if ORM features are needed later
- Ensure all database operations are cancellable via context
- Document any database-specific features used (PostgreSQL vs SQLite)
