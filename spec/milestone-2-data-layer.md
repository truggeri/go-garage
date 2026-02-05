# Milestone 2: Vehicle Data Model and Database Layer

## Objective
Implement the core data models and database layer for vehicle management, including database schema, repositories, and basic CRUD operations.

## Prerequisites
- Milestone 1 completed (project setup and infrastructure)

## Goals

### 1. Database Setup
- [ ] Configure database connection with connection pooling
- [ ] Implement health check for database connectivity
- [ ] Setup database migrations system
- [ ] Create initial migration files

### 2. Data Models

#### Vehicle Model
Define the Vehicle struct with the following fields:
- [ ] ID (UUID)
- [ ] User ID (foreign key)
- [ ] VIN (Vehicle Identification Number)
- [ ] Make
- [ ] Model
- [ ] Year
- [ ] Color
- [ ] License plate
- [ ] Purchase date
- [ ] Purchase price
- [ ] Purchase mileage
- [ ] Current mileage
- [ ] Status (active, sold, scrapped)
- [ ] Notes
- [ ] Created at timestamp
- [ ] Updated at timestamp

#### Maintenance Record Model
Define the MaintenanceRecord struct:
- [ ] ID
- [ ] Vehicle ID (foreign key)
- [ ] Service type (oil change, tire rotation, inspection, etc.)
- [ ] Service date
- [ ] Mileage at service
- [ ] Cost
- [ ] Service provider
- [ ] Notes/description
- [ ] Created at timestamp
- [ ] Updated at timestamp

#### User Model
Define the User struct:
- [ ] ID
- [ ] Username (unique)
- [ ] Email (unique)
- [ ] Password hash
- [ ] First name
- [ ] Last name
- [ ] Created at timestamp
- [ ] Updated at timestamp
- [ ] Last login timestamp

### 3. Database Schema
- [ ] Create users table migration
- [ ] Create vehicles table migration
- [ ] Create maintenance_records table migration
- [ ] Add indexes for foreign keys
- [ ] Add indexes for frequently queried fields (VIN, email, username)
- [ ] Add unique constraints where needed

### 4. Repository Interface Design
Create repository interfaces for:
- [ ] VehicleRepository interface
  - Create(vehicle)
  - FindByID(id)
  - FindByUserID(userID)
  - FindByVIN(vin)
  - Update(vehicle)
  - Delete(id)
  - List(filters, pagination)

- [ ] MaintenanceRepository interface
  - Create(record)
  - FindByID(id)
  - FindByVehicleID(vehicleID)
  - Update(record)
  - Delete(id)
  - List(filters, pagination)

- [ ] UserRepository interface
  - Create(user)
  - FindByID(id)
  - FindByEmail(email)
  - FindByUsername(username)
  - Update(user)
  - Delete(id)

### 5. Repository Implementation
- [ ] Implement SQLite repository for vehicles
- [ ] Implement SQLite repository for maintenance records
- [ ] Implement SQLite repository for users
- [ ] Use prepared statements for all queries
- [ ] Implement transaction support
- [ ] Add context support for cancellation

### 6. Data Validation
- [ ] Implement validation for Vehicle model
  - VIN format validation (17 characters)
  - Year range validation (1900-present)
  - Required fields validation
  - Price and mileage validation (non-negative)

- [ ] Implement validation for MaintenanceRecord
  - Date validation (not in future)
  - Cost validation (non-negative)
  - Required fields validation

- [ ] Implement validation for User
  - Email format validation
  - Username format validation (alphanumeric, minimum length)
  - Password strength requirements

### 7. Error Handling
- [ ] Define custom error types (NotFoundError, ValidationError, etc.)
- [ ] Implement error wrapping with context
- [ ] Create error helper functions
- [ ] Document expected errors for each repository method

### 8. Testing

#### Unit Tests
- [ ] Test all repository methods
- [ ] Test data validation functions
- [ ] Test error handling paths
- [ ] Mock database for unit tests

#### Integration Tests
- [ ] Test with test database
- [ ] Test CRUD operations end-to-end
- [ ] Test concurrent access scenarios
- [ ] Test transaction rollback scenarios
- [ ] Test database constraints (unique, foreign keys)

### 9. Database Utilities
- [ ] Create database seeding script for development
- [ ] Create sample data for testing
- [ ] Implement database backup helper
- [ ] Create migration rollback procedures

### 10. Documentation
- [ ] Document database schema with ERD
- [ ] Document repository interfaces
- [ ] Document validation rules
- [ ] Add inline code documentation
- [ ] Create database setup guide

## Deliverables

1. **Database Schema**: Complete schema with all tables, indexes, and constraints
2. **Repository Layer**: Full implementation of all repository interfaces
3. **Data Models**: Validated Go structs for all domain entities
4. **Migrations**: Up and down migrations for schema changes
5. **Tests**: Comprehensive unit and integration tests (>80% coverage)
6. **Documentation**: Complete API documentation for repositories

## Success Criteria

- [ ] All migrations run successfully
- [ ] All repository tests pass
- [ ] Code coverage >80% for repository layer
- [ ] Database constraints enforce data integrity
- [ ] Queries are efficient (use EXPLAIN ANALYZE)
- [ ] No SQL injection vulnerabilities
- [ ] Concurrent access is handled correctly
- [ ] All validation rules work as expected

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
