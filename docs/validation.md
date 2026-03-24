# Data Validation Rules

## Overview

Go-Garage validates all data before persisting to the database. This document describes the validation rules for each model and how to handle validation errors.

## Validation Functions

Validation functions are located in `internal/models/validation.go`:

```go
ValidateUser(u *User) error
ValidatePassword(password string) error
ValidateVehicle(v *Vehicle) error
ValidateMaintenanceRecord(m *MaintenanceRecord) error
ValidateFuelRecord(f *FuelRecord) error
```

## User Validation

### ValidateUser

Validates a User model before creation or update.

```go
err := models.ValidateUser(user)
if err != nil {
    var valErr *models.ValidationError
    if errors.As(err, &valErr) {
        fmt.Printf("Field: %s, Error: %s\n", valErr.Field, valErr.Message)
    }
}
```

### Field Rules

| Field | Rule | Error Message |
|-------|------|---------------|
| `username` | Required | "username is required" |
| `username` | Minimum 3 characters | "username must be at least 3 characters long" |
| `username` | Maximum 30 characters | "username must be at most 30 characters long" |
| `username` | Alphanumeric with underscores and hyphens only | "username can only contain alphanumeric characters, underscores, and hyphens" |
| `email` | Required | "email is required" |
| `email` | Valid email format | "invalid email format" |

### Username Format

Valid usernames match the pattern: `^[a-zA-Z0-9_\-]+$`

**Valid Examples:**

- `johndoe`
- `john_doe`
- `john-doe`
- `JohnDoe123`

**Invalid Examples:**

- `john doe` (contains space)
- `john@doe` (contains @)
- `jo` (too short)

### Email Format

Valid emails match a standard email pattern.

**Valid Examples:**

- `john@example.com`
- `john.doe@example.co.uk`
- `john+tag@example.com`

**Invalid Examples:**

- `john@` (incomplete)
- `@example.com` (no local part)
- `john` (no domain)

---

## Password Validation

### ValidatePassword

Validates a plain text password before hashing.

```go
err := models.ValidatePassword(plainPassword)
if err != nil {
    // Handle validation error
}
```

### Password Rules

| Rule | Error Message |
|------|---------------|
| Required | "password is required" |
| Minimum 8 characters | "password must be at least 8 characters long" |
| At least one uppercase letter | "password must contain at least one uppercase letter, one lowercase letter, and one digit" |
| At least one lowercase letter | (same as above) |
| At least one digit | (same as above) |

### Password Requirements Summary

1. **Length**: Minimum 8 characters
2. **Uppercase**: At least one A-Z
3. **Lowercase**: At least one a-z
4. **Digit**: At least one 0-9

**Valid Examples:**

- `Password1` ✓
- `MySecure123` ✓
- `Abc12345` ✓

**Invalid Examples:**

- `password` ✗ (no uppercase, no digit)
- `PASSWORD1` ✗ (no lowercase)
- `Password` ✗ (no digit)
- `Pass1` ✗ (too short)

---

## Vehicle Validation

### ValidateVehicle

Validates a Vehicle model before creation or update.

```go
err := models.ValidateVehicle(vehicle)
if err != nil {
    var valErr *models.ValidationError
    if errors.As(err, &valErr) {
        fmt.Printf("Field: %s, Error: %s\n", valErr.Field, valErr.Message)
    }
}
```

### Field Rules

| Field | Rule | Error Message |
|-------|------|---------------|
| `user_id` | Required | "user ID is required" |
| `vin` | Required | "VIN is required" |
| `vin` | Exactly 17 characters | "VIN must be exactly 17 characters" |
| `vin` | Valid characters only (no I, O, Q) | "VIN contains invalid characters (cannot include I, O, or Q)" |
| `make` | Required | "make is required" |
| `model` | Required | "model is required" |
| `year` | Required (non-zero) | "year is required" |
| `year` | Between 1900 and current year + 1 | "year must be between 1900 and current year + 1" |
| `status` | Required | "status is required" |
| `status` | Must be valid enum value | "status must be 'active', 'sold', or 'scrapped'" |
| `purchase_price` | Non-negative if provided | "purchase price cannot be negative" |
| `purchase_mileage` | Non-negative if provided | "purchase mileage cannot be negative" |
| `current_mileage` | Non-negative if provided | "current mileage cannot be negative" |
| `current_mileage` | >= purchase_mileage if both provided | "current mileage cannot be less than purchase mileage" |

### VIN Format

Vehicle Identification Numbers (VINs) must:

1. Be exactly 17 characters
2. Contain only alphanumeric characters (A-Z, 0-9)
3. Not contain letters I, O, or Q (to avoid confusion with 1 and 0)

**VIN Pattern**: `^[A-HJ-NPR-Z0-9]{17}$`

**Valid Examples:**

- `1HGCM82633A004352`
- `WVWZZZ3CZWE123456`
- `5YJSA1S20EFP12345`

**Invalid Examples:**

- `1HGCM82633A00435` (16 characters)
- `1HGCM82633A0043521` (18 characters)
- `1HGCM826I3A004352` (contains I)
- `1HGCM8263OA004352` (contains O)
- `1HGCM8263QA004352` (contains Q)

**VIN Normalization:**

- Spaces are automatically removed
- Letters are converted to uppercase

```go
// These are equivalent after normalization:
"1hgcm82633a004352"   → "1HGCM82633A004352"
"1HGC M826 33A0 0435 2" → "1HGCM82633A004352"
```

### Year Validation

The year must be:

- At least 1900 (minimum supported year)
- At most current year + 1 (allows for next model year vehicles)

**Example (in 2024):**

- Valid: 1900, 1950, 2020, 2024, 2025
- Invalid: 1899, 2026, 2030

### Vehicle Status

Valid status values:

- `active` - Vehicle is currently in use
- `sold` - Vehicle has been sold
- `scrapped` - Vehicle has been disposed of

```go
const (
    VehicleStatusActive   VehicleStatus = "active"
    VehicleStatusSold     VehicleStatus = "sold"
    VehicleStatusScrapped VehicleStatus = "scrapped"
)
```

### Mileage Validation

- Purchase mileage must be non-negative (if provided)
- Current mileage must be non-negative (if provided)
- Current mileage must be >= purchase mileage (if both provided)

This ensures mileage values are logical and prevents data entry errors.

---

## Maintenance Record Validation

### ValidateMaintenanceRecord

Validates a MaintenanceRecord model before creation or update.

```go
err := models.ValidateMaintenanceRecord(record)
if err != nil {
    var valErr *models.ValidationError
    if errors.As(err, &valErr) {
        fmt.Printf("Field: %s, Error: %s\n", valErr.Field, valErr.Message)
    }
}
```

### Field Rules

| Field | Rule | Error Message |
|-------|------|---------------|
| `vehicle_id` | Required | "vehicle ID is required" |
| `service_type` | Required | "service type is required" |
| `service_type` | Must be a valid enum value | "invalid service type" |
| `custom_service_type` | Required when `service_type` is "other" | "custom service type is required when service type is Other" |
| `service_date` | Required (non-zero) | "service date is required" |
| `service_date` | Not in the future | "service date cannot be in the future" |
| `cost` | Non-negative if provided | "cost cannot be negative" |
| `mileage_at_service` | Non-negative if provided | "mileage at service cannot be negative" |

### Service Types

The `service_type` field must be one of the following predefined enum values:

| Enum Value | Display Name |
|-----------|-------------|
| `oil_change` | Oil Change |
| `tire_rotation` | Tire Rotation |
| `air_filter` | Air Filter |
| `cabin_air_filter` | Cabin Air Filter |
| `fuel_additive` | Fuel Additive |
| `battery` | Battery |
| `brakes` | Brakes |
| `brake_fluid` | Brake Fluid |
| `radiator_fluid` | Radiator Fluid |
| `tires` | Tires |
| `glass` | Glass |
| `body_work` | Body Work |
| `interior` | Interior |
| `other` | Other |

When `service_type` is `"other"`, the `custom_service_type` field must be provided with a description of the service.

### Service Date

The service date:

- Cannot be empty (zero value)
- Cannot be in the future

This prevents scheduling future maintenance as completed work.

---

## Fuel Record Validation

### ValidateFuelRecord

Validates a FuelRecord model before creation or update.

```go
err := models.ValidateFuelRecord(record)
if err != nil {
    var valErr *models.ValidationError
    if errors.As(err, &valErr) {
        fmt.Printf("Field: %s, Error: %s\n", valErr.Field, valErr.Message)
    }
}
```

### Field Rules

| Field | Rule | Error Message |
|-------|------|---------------|
| `vehicle_id` | Required | "vehicle ID is required" |
| `fill_date` | Required (non-zero) | "fill date is required" |
| `fill_date` | Not in the future | "fill date cannot be in the future" |
| `mileage` | Must be > 0 | "mileage must be greater than zero" |
| `volume` | Must be > 0 | "volume must be greater than zero" |
| `fuel_type` | Required | "fuel type is required" |
| `fuel_type` | Must be valid enum value | "invalid fuel type" |
| `price_per_unit` | Non-negative if provided | "price per unit cannot be negative" |
| `octane_rating` | Greater than zero if provided | "octane rating must be greater than zero" |
| `city_driving_percentage` | 0-100 if provided | "city driving percentage must be between 0 and 100" |
| `vehicle_reported_mpg` | Greater than zero if provided | "vehicle reported MPG must be greater than zero" |

### Fuel Types

The `fuel_type` field must be one of the following:

| Enum Value | Display Name |
|-----------|-------------|
| `gasoline` | Gasoline |
| `diesel` | Diesel |
| `e85` | E85 |

### Fill Date

The fill date:

- Cannot be empty (zero value)
- Cannot be in the future

---

## Error Types

### ValidationError

All validation functions return `*ValidationError` on failure:

```go
type ValidationError struct {
    Field   string  // The field that failed validation
    Message string  // Human-readable error message
}

func (e *ValidationError) Error() string {
    if e.Field != "" {
        return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
    }
    return fmt.Sprintf("validation error: %s", e.Message)
}
```

### Creating Validation Errors

```go
// Create a validation error
err := models.NewValidationError("email", "invalid email format")
```

### Handling Validation Errors

```go
err := models.ValidateVehicle(vehicle)
if err != nil {
    var valErr *models.ValidationError
    if errors.As(err, &valErr) {
        // Handle specific validation error
        switch valErr.Field {
        case "vin":
            // Handle VIN error
        case "year":
            // Handle year error
        default:
            // Handle other errors
        }
    }
    return err
}
```

---

## Best Practices

### 1. Validate Early

Validate data as early as possible, preferably in the handler layer:

```go
func CreateVehicleHandler(w http.ResponseWriter, r *http.Request) {
    vehicle := parseVehicle(r)
    
    // Validate before calling service
    if err := models.ValidateVehicle(vehicle); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Proceed with creation
    err := vehicleService.Create(ctx, vehicle)
    // ...
}
```

### 2. Provide User-Friendly Messages

The validation error messages are designed to be user-friendly and can be returned directly to API clients:

```go
{
    "error": "validation error on field 'vin': VIN must be exactly 17 characters"
}
```

### 3. Validate All Fields Together

When displaying forms, you may want to collect all validation errors at once. Currently, validation stops at the first error. For multi-field validation, consider creating a wrapper:

```go
func ValidateAllFields(vehicle *models.Vehicle) []error {
    var errors []error
    
    // Check each field individually
    if vehicle.VIN == "" {
        errors = append(errors, models.NewValidationError("vin", "VIN is required"))
    } else if len(vehicle.VIN) != 17 {
        errors = append(errors, models.NewValidationError("vin", "VIN must be exactly 17 characters"))
    }
    
    // ... check other fields
    
    return errors
}
```

### 4. Normalize Before Validation

VIN normalization happens automatically, but for other fields you may want to normalize before validation:

```go
// Normalize email to lowercase
user.Email = strings.ToLower(strings.TrimSpace(user.Email))

// Then validate
err := models.ValidateUser(user)
```

---

## Validation Constants

The following constants are used in validation:

```go
const (
    minUsernameLength = 3    // Minimum username length
    maxUsernameLength = 30   // Maximum username length
    minPasswordLength = 8    // Minimum password length
    vinLength         = 17   // Required VIN length
    minYear           = 1900 // Minimum vehicle year
)
```

---

## See Also

- [Repository Interfaces](./repositories.md) - How validation is used in repositories
- [Database Schema](./database-schema.md) - Database constraints
- [API Documentation](../spec/restful-api.md) - API error responses
