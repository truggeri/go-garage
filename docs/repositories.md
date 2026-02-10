# Repository Interfaces Documentation

## Overview

Go-Garage uses the Repository pattern to abstract database operations. This document describes the repository interfaces, their methods, and expected behaviors.

## Package Structure

```
internal/repositories/
├── user.go                  # UserRepository interface
├── user_sqlite.go           # SQLite implementation
├── user_sqlite_test.go      # Unit tests
├── vehicle.go               # VehicleRepository interface
├── vehicle_sqlite.go        # SQLite implementation
├── vehicle_sqlite_test.go   # Unit tests
├── maintenance.go           # MaintenanceRepository interface
├── maintenance_sqlite.go    # SQLite implementation
├── maintenance_sqlite_test.go # Unit tests
└── testutils_test.go        # Shared test utilities
```

## Common Types

### PaginationParams

Used for paginating list results.

```go
type PaginationParams struct {
    Limit  int  // Maximum number of results to return
    Offset int  // Number of results to skip
}
```

**Example Usage:**
```go
// Get first 10 vehicles
pagination := PaginationParams{Limit: 10, Offset: 0}

// Get next 10 vehicles
pagination := PaginationParams{Limit: 10, Offset: 10}
```

## UserRepository

Interface for user data access operations.

### Interface Definition

```go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    FindByID(ctx context.Context, id string) (*models.User, error)
    FindByEmail(ctx context.Context, email string) (*models.User, error)
    FindByUsername(ctx context.Context, username string) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id string) error
    UpdateLastLogin(ctx context.Context, id string) error
}
```

### Method Details

#### Create

Inserts a new user into the database.

```go
Create(ctx context.Context, user *models.User) error
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `user`: User object to create. ID will be generated if empty.

**Returns:**
- `error`: `nil` on success, `*models.ValidationError` if validation fails, `*models.DuplicateError` if username or email exists

**Side Effects:**
- Sets `user.ID` if empty (generates UUID)
- Sets `user.CreatedAt` and `user.UpdatedAt` to current time

**Example:**
```go
user := &models.User{
    Username:     "johndoe",
    Email:        "john@example.com",
    PasswordHash: hashedPassword,
    FirstName:    "John",
    LastName:     "Doe",
}
err := repo.Create(ctx, user)
if err != nil {
    var dupErr *models.DuplicateError
    if errors.As(err, &dupErr) {
        // Handle duplicate username/email
    }
    return err
}
fmt.Println("Created user:", user.ID)
```

#### FindByID

Retrieves a user by their ID.

```go
FindByID(ctx context.Context, id string) (*models.User, error)
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `id`: UUID of the user

**Returns:**
- `*models.User`: User object if found
- `error`: `*models.NotFoundError` if user doesn't exist

**Example:**
```go
user, err := repo.FindByID(ctx, "550e8400-e29b-41d4-a716-446655440000")
if err != nil {
    var notFound *models.NotFoundError
    if errors.As(err, &notFound) {
        // User not found
    }
    return err
}
```

#### FindByEmail

Retrieves a user by their email address.

```go
FindByEmail(ctx context.Context, email string) (*models.User, error)
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `email`: Email address to search for

**Returns:**
- `*models.User`: User object if found
- `error`: `*models.NotFoundError` if user doesn't exist

**Use Case:** Authentication by email

#### FindByUsername

Retrieves a user by their username.

```go
FindByUsername(ctx context.Context, username string) (*models.User, error)
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `username`: Username to search for

**Returns:**
- `*models.User`: User object if found
- `error`: `*models.NotFoundError` if user doesn't exist

**Use Case:** Authentication by username

#### Update

Modifies an existing user's information.

```go
Update(ctx context.Context, user *models.User) error
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `user`: User object with updated fields. ID must be set.

**Returns:**
- `error`: `*models.NotFoundError` if user doesn't exist, `*models.ValidationError` if validation fails, `*models.DuplicateError` if email/username conflicts

**Side Effects:**
- Updates `user.UpdatedAt` to current time

**Example:**
```go
user.FirstName = "Jane"
user.LastName = "Smith"
err := repo.Update(ctx, user)
```

#### Delete

Removes a user from the database.

```go
Delete(ctx context.Context, id string) error
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `id`: UUID of the user to delete

**Returns:**
- `error`: `*models.NotFoundError` if user doesn't exist

**Side Effects:**
- Cascading delete of all user's vehicles and their maintenance records

#### UpdateLastLogin

Updates the last login timestamp for a user.

```go
UpdateLastLogin(ctx context.Context, id string) error
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `id`: UUID of the user

**Returns:**
- `error`: `*models.NotFoundError` if user doesn't exist

**Use Case:** Called after successful authentication

---

## VehicleRepository

Interface for vehicle data access operations.

### Filter Types

```go
type VehicleFilters struct {
    Status *models.VehicleStatus  // Filter by status (active, sold, scrapped)
    Make   *string                // Filter by make
    Model  *string                // Filter by model
    Year   *int                   // Filter by year
}
```

### Interface Definition

```go
type VehicleRepository interface {
    Create(ctx context.Context, vehicle *models.Vehicle) error
    FindByID(ctx context.Context, id string) (*models.Vehicle, error)
    FindByUserID(ctx context.Context, userID string) ([]*models.Vehicle, error)
    FindByVIN(ctx context.Context, vin string) (*models.Vehicle, error)
    Update(ctx context.Context, vehicle *models.Vehicle) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filters VehicleFilters, pagination PaginationParams) ([]*models.Vehicle, error)
}
```

### Method Details

#### Create

Inserts a new vehicle into the database.

```go
Create(ctx context.Context, vehicle *models.Vehicle) error
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `vehicle`: Vehicle object to create. ID will be generated if empty.

**Returns:**
- `error`: `*models.ValidationError` if validation fails, `*models.DuplicateError` if VIN exists

**Validation:**
- VIN must be exactly 17 characters (excluding I, O, Q)
- Year must be between 1900 and current year + 1
- UserID must reference an existing user
- Make, Model, and Status are required

**Example:**
```go
vehicle := &models.Vehicle{
    UserID: userID,
    VIN:    "1HGCM82633A004352",
    Make:   "Honda",
    Model:  "Accord",
    Year:   2023,
    Status: models.VehicleStatusActive,
}
err := repo.Create(ctx, vehicle)
```

#### FindByID

Retrieves a vehicle by its ID.

```go
FindByID(ctx context.Context, id string) (*models.Vehicle, error)
```

**Returns:**
- `*models.Vehicle`: Vehicle object if found
- `error`: `*models.NotFoundError` if vehicle doesn't exist

#### FindByUserID

Retrieves all vehicles for a specific user.

```go
FindByUserID(ctx context.Context, userID string) ([]*models.Vehicle, error)
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `userID`: UUID of the user

**Returns:**
- `[]*models.Vehicle`: Slice of vehicles (empty if none found)
- `error`: Database errors only (not NotFoundError for empty results)

**Example:**
```go
vehicles, err := repo.FindByUserID(ctx, userID)
if err != nil {
    return err
}
for _, v := range vehicles {
    fmt.Printf("%s %s %d\n", v.Make, v.Model, v.Year)
}
```

#### FindByVIN

Retrieves a vehicle by its VIN.

```go
FindByVIN(ctx context.Context, vin string) (*models.Vehicle, error)
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `vin`: Vehicle Identification Number

**Returns:**
- `*models.Vehicle`: Vehicle object if found
- `error`: `*models.NotFoundError` if vehicle doesn't exist

**Use Case:** VIN lookup for registration, validation

#### Update

Modifies an existing vehicle's information.

```go
Update(ctx context.Context, vehicle *models.Vehicle) error
```

**Returns:**
- `error`: `*models.NotFoundError` if vehicle doesn't exist, `*models.ValidationError` if validation fails, `*models.DuplicateError` if VIN conflicts

#### Delete

Removes a vehicle from the database.

```go
Delete(ctx context.Context, id string) error
```

**Returns:**
- `error`: `*models.NotFoundError` if vehicle doesn't exist

**Side Effects:**
- Cascading delete of all vehicle's maintenance records

#### List

Retrieves vehicles with optional filters and pagination.

```go
List(ctx context.Context, filters VehicleFilters, pagination PaginationParams) ([]*models.Vehicle, error)
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `filters`: Optional filters (nil values are ignored)
- `pagination`: Limit and offset for results

**Returns:**
- `[]*models.Vehicle`: Slice of matching vehicles
- `error`: Database errors only

**Example:**
```go
// Get all active vehicles, page 1
status := models.VehicleStatusActive
filters := VehicleFilters{Status: &status}
pagination := PaginationParams{Limit: 20, Offset: 0}
vehicles, err := repo.List(ctx, filters, pagination)
```

---

## MaintenanceRepository

Interface for maintenance record data access operations.

### Filter Types

```go
type MaintenanceFilters struct {
    ServiceType *string  // Filter by service type
}
```

### Interface Definition

```go
type MaintenanceRepository interface {
    Create(ctx context.Context, record *models.MaintenanceRecord) error
    FindByID(ctx context.Context, id string) (*models.MaintenanceRecord, error)
    FindByVehicleID(ctx context.Context, vehicleID string) ([]*models.MaintenanceRecord, error)
    Update(ctx context.Context, record *models.MaintenanceRecord) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filters MaintenanceFilters, pagination PaginationParams) ([]*models.MaintenanceRecord, error)
}
```

### Method Details

#### Create

Inserts a new maintenance record into the database.

```go
Create(ctx context.Context, record *models.MaintenanceRecord) error
```

**Validation:**
- VehicleID must reference an existing vehicle
- ServiceType is required
- ServiceDate is required and cannot be in the future
- Cost must be non-negative if provided

**Example:**
```go
record := &models.MaintenanceRecord{
    VehicleID:   vehicleID,
    ServiceType: "Oil Change",
    ServiceDate: time.Now(),
    Cost:        ptr(49.99),
}
err := repo.Create(ctx, record)
```

#### FindByID

Retrieves a maintenance record by its ID.

```go
FindByID(ctx context.Context, id string) (*models.MaintenanceRecord, error)
```

**Returns:**
- `*models.MaintenanceRecord`: Record if found
- `error`: `*models.NotFoundError` if record doesn't exist

#### FindByVehicleID

Retrieves all maintenance records for a specific vehicle.

```go
FindByVehicleID(ctx context.Context, vehicleID string) ([]*models.MaintenanceRecord, error)
```

**Returns:**
- `[]*models.MaintenanceRecord`: Slice of records (empty if none found)
- `error`: Database errors only

**Note:** Results are typically ordered by service date descending (most recent first)

#### Update

Modifies an existing maintenance record.

```go
Update(ctx context.Context, record *models.MaintenanceRecord) error
```

**Returns:**
- `error`: `*models.NotFoundError` if record doesn't exist, `*models.ValidationError` if validation fails

#### Delete

Removes a maintenance record from the database.

```go
Delete(ctx context.Context, id string) error
```

**Returns:**
- `error`: `*models.NotFoundError` if record doesn't exist

#### List

Retrieves maintenance records with optional filters and pagination.

```go
List(ctx context.Context, filters MaintenanceFilters, pagination PaginationParams) ([]*models.MaintenanceRecord, error)
```

**Example:**
```go
// Get all oil changes
serviceType := "Oil Change"
filters := MaintenanceFilters{ServiceType: &serviceType}
pagination := PaginationParams{Limit: 100, Offset: 0}
records, err := repo.List(ctx, filters, pagination)
```

---

## Error Handling

All repository methods may return the following error types:

### NotFoundError

Returned when a requested resource doesn't exist.

```go
type NotFoundError struct {
    Resource string  // e.g., "user", "vehicle", "maintenance_record"
    ID       string  // The ID that was not found
}
```

**Handling:**
```go
user, err := repo.FindByID(ctx, id)
if err != nil {
    var notFound *models.NotFoundError
    if errors.As(err, &notFound) {
        return fmt.Errorf("user %s not found", id)
    }
    return err
}
```

### ValidationError

Returned when input data fails validation.

```go
type ValidationError struct {
    Field   string  // Field that failed validation
    Message string  // Human-readable error message
}
```

**Handling:**
```go
err := repo.Create(ctx, user)
if err != nil {
    var valErr *models.ValidationError
    if errors.As(err, &valErr) {
        return fmt.Errorf("invalid %s: %s", valErr.Field, valErr.Message)
    }
    return err
}
```

### DuplicateError

Returned when a unique constraint would be violated.

```go
type DuplicateError struct {
    Resource string  // e.g., "user", "vehicle"
    Field    string  // Field with duplicate value (e.g., "email", "vin")
    Value    string  // The duplicate value
}
```

**Handling:**
```go
err := repo.Create(ctx, user)
if err != nil {
    var dupErr *models.DuplicateError
    if errors.As(err, &dupErr) {
        return fmt.Errorf("%s already registered", dupErr.Field)
    }
    return err
}
```

### DatabaseError

Wraps low-level database errors with operation context.

```go
type DatabaseError struct {
    Err       error   // Underlying error
    Operation string  // Operation that failed (e.g., "create user")
}
```

---

## SQLite Implementation

### Creating Repositories

```go
import (
    "github.com/truggeri/go-garage/internal/database"
    "github.com/truggeri/go-garage/internal/repositories"
)

// Initialize database
db, err := database.InitializeGarage("./data/go-garage.db", database.StandardWorkerPoolSettings())
if err != nil {
    log.Fatal(err)
}
defer db.Terminate()

// Create repositories
userRepo := repositories.NewSQLiteUserRepository(db.RawSQLConnection())
vehicleRepo := repositories.NewSQLiteVehicleRepository(db.RawSQLConnection())
maintenanceRepo := repositories.NewSQLiteMaintenanceRepository(db.RawSQLConnection())
```

### Transaction Support

For operations that need to span multiple repositories, use database transactions:

```go
tx, err := db.RawSQLConnection().BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

// Perform operations...

if err := tx.Commit(); err != nil {
    return err
}
```

### Context Support

All repository methods accept a `context.Context` for:
- **Cancellation**: Cancel long-running queries
- **Timeouts**: Set query timeouts
- **Tracing**: Propagate request IDs for logging

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user, err := repo.FindByID(ctx, id)
```

---

## Testing

### Test Utilities

The `testutils_test.go` file provides helpers for testing:

```go
// Create test database
db := setupTestDB(t)
defer db.Close()

// Create repository with test db
repo := NewSQLiteUserRepository(db)
```

### Mocking Repositories

For unit testing services, create mock implementations:

```go
type MockUserRepository struct {
    users map[string]*models.User
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
    if user, ok := m.users[id]; ok {
        return user, nil
    }
    return nil, models.NewNotFoundError("user", id)
}

// Implement other methods...
```

---

## See Also

- [Database Schema](./database-schema.md)
- [Data Validation Rules](./validation.md)
- [Database Setup Guide](./database-setup.md)
- [Database Migrations](./database-migrations.md)
