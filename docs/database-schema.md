# Database Schema Documentation

## Overview

Go-Garage uses SQLite as its database. The schema is managed through versioned migrations located in the `migrations/` directory. This document provides a comprehensive overview of the database schema, entity relationships, and design decisions.

## Entity Relationship Diagram (ERD)

```
┌─────────────────────────────────────────┐
│                 users                    │
├─────────────────────────────────────────┤
│ id           TEXT    PK                  │
│ username     TEXT    UNIQUE NOT NULL     │
│ email        TEXT    UNIQUE NOT NULL     │
│ password_hash TEXT   NOT NULL            │
│ first_name   TEXT                        │
│ last_name    TEXT                        │
│ created_at   DATETIME NOT NULL           │
│ updated_at   DATETIME NOT NULL           │
│ last_login_at DATETIME                   │
├─────────────────────────────────────────┤
│ Indexes:                                 │
│   idx_users_email (email)               │
│   idx_users_username (username)         │
└───────────────────┬─────────────────────┘
                    │
                    │ 1:N
                    ▼
┌─────────────────────────────────────────┐
│               vehicles                   │
├─────────────────────────────────────────┤
│ id              TEXT    PK               │
│ user_id         TEXT    FK NOT NULL      │
│ vin             TEXT    UNIQUE NOT NULL  │
│ make            TEXT    NOT NULL         │
│ model           TEXT    NOT NULL         │
│ year            INTEGER NOT NULL         │
│ color           TEXT                     │
│ license_plate   TEXT                     │
│ purchase_date   DATE                     │
│ purchase_price  REAL                     │
│ purchase_mileage INTEGER                 │
│ current_mileage INTEGER                  │
│ status          TEXT    NOT NULL         │
│ notes           TEXT                     │
│ created_at      DATETIME NOT NULL        │
│ updated_at      DATETIME NOT NULL        │
├─────────────────────────────────────────┤
│ Constraints:                             │
│   CHECK(status IN ('active','sold',     │
│         'scrapped'))                     │
│   FOREIGN KEY (user_id) REFERENCES       │
│         users(id) ON DELETE CASCADE      │
│ Indexes:                                 │
│   idx_vehicles_user_id (user_id)        │
│   idx_vehicles_vin (vin)                │
│   idx_vehicles_status (status)          │
└───────────────────┬─────────────────────┘
                    │
                    │ 1:N
                    ▼
┌─────────────────────────────────────────┐
│          maintenance_records             │
├─────────────────────────────────────────┤
│ id                TEXT    PK             │
│ vehicle_id        TEXT    FK NOT NULL    │
│ service_type      TEXT    NOT NULL       │
│ service_date      DATE    NOT NULL       │
│ mileage_at_service INTEGER               │
│ cost              REAL                   │
│ service_provider  TEXT                   │
│ notes             TEXT                   │
│ created_at        DATETIME NOT NULL      │
│ updated_at        DATETIME NOT NULL      │
├─────────────────────────────────────────┤
│ Constraints:                             │
│   FOREIGN KEY (vehicle_id) REFERENCES    │
│         vehicles(id) ON DELETE CASCADE   │
│ Indexes:                                 │
│   idx_maintenance_vehicle_id (vehicle_id)│
│   idx_maintenance_service_date           │
│         (service_date)                   │
└─────────────────────────────────────────┘
```

## Table Descriptions

### users

The `users` table stores user account information and authentication credentials.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT | PRIMARY KEY | UUID identifier for the user |
| `username` | TEXT | UNIQUE, NOT NULL | Unique username for login |
| `email` | TEXT | UNIQUE, NOT NULL | Unique email address |
| `password_hash` | TEXT | NOT NULL | Bcrypt-hashed password |
| `first_name` | TEXT | | User's first name |
| `last_name` | TEXT | | User's last name |
| `created_at` | DATETIME | NOT NULL | Timestamp of account creation |
| `updated_at` | DATETIME | NOT NULL | Timestamp of last update |
| `last_login_at` | DATETIME | | Timestamp of last successful login |

**Indexes:**
- `idx_users_email` - Speeds up email lookups for authentication
- `idx_users_username` - Speeds up username lookups for authentication

### vehicles

The `vehicles` table stores information about vehicles owned by users.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT | PRIMARY KEY | UUID identifier for the vehicle |
| `user_id` | TEXT | FK, NOT NULL | Reference to the owning user |
| `vin` | TEXT | UNIQUE, NOT NULL | 17-character Vehicle Identification Number |
| `make` | TEXT | NOT NULL | Vehicle manufacturer (e.g., "Toyota") |
| `model` | TEXT | NOT NULL | Vehicle model (e.g., "Camry") |
| `year` | INTEGER | NOT NULL | Model year (1900 - current year + 1) |
| `color` | TEXT | | Vehicle color |
| `license_plate` | TEXT | | License plate number |
| `purchase_date` | DATE | | Date vehicle was purchased |
| `purchase_price` | REAL | | Purchase price in dollars |
| `purchase_mileage` | INTEGER | | Mileage at time of purchase |
| `current_mileage` | INTEGER | | Current odometer reading |
| `status` | TEXT | NOT NULL, CHECK | Vehicle status: 'active', 'sold', or 'scrapped' |
| `notes` | TEXT | | Additional notes about the vehicle |
| `created_at` | DATETIME | NOT NULL | Timestamp of record creation |
| `updated_at` | DATETIME | NOT NULL | Timestamp of last update |

**Constraints:**
- Foreign key to `users(id)` with `ON DELETE CASCADE`
- Check constraint: `status IN ('active', 'sold', 'scrapped')`

**Indexes:**
- `idx_vehicles_user_id` - Speeds up queries for a user's vehicles
- `idx_vehicles_vin` - Speeds up VIN lookups
- `idx_vehicles_status` - Speeds up filtering by status

### maintenance_records

The `maintenance_records` table stores service and maintenance history for vehicles.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT | PRIMARY KEY | UUID identifier for the record |
| `vehicle_id` | TEXT | FK, NOT NULL | Reference to the vehicle |
| `service_type` | TEXT | NOT NULL | Type of service (e.g., "Oil Change", "Tire Rotation") |
| `service_date` | DATE | NOT NULL | Date service was performed |
| `mileage_at_service` | INTEGER | | Odometer reading at service time |
| `cost` | REAL | | Service cost in dollars |
| `service_provider` | TEXT | | Name of service provider/shop |
| `notes` | TEXT | | Additional notes about the service |
| `created_at` | DATETIME | NOT NULL | Timestamp of record creation |
| `updated_at` | DATETIME | NOT NULL | Timestamp of last update |

**Constraints:**
- Foreign key to `vehicles(id)` with `ON DELETE CASCADE`

**Indexes:**
- `idx_maintenance_vehicle_id` - Speeds up queries for a vehicle's maintenance history
- `idx_maintenance_service_date` - Speeds up date-based queries and sorting

## Relationships

### User → Vehicles (One-to-Many)

A user can own multiple vehicles. When a user is deleted, all their vehicles are automatically deleted due to `ON DELETE CASCADE`.

```sql
-- Example: Get all vehicles for a user
SELECT * FROM vehicles WHERE user_id = ?;
```

### Vehicle → Maintenance Records (One-to-Many)

A vehicle can have multiple maintenance records. When a vehicle is deleted, all its maintenance records are automatically deleted due to `ON DELETE CASCADE`.

```sql
-- Example: Get all maintenance for a vehicle
SELECT * FROM maintenance_records WHERE vehicle_id = ? ORDER BY service_date DESC;
```

## Data Types

### UUID (TEXT)

All primary keys use UUIDs stored as TEXT. UUIDs are generated using `github.com/google/uuid` in Go code.

```go
id := uuid.New().String()
```

### Timestamps (DATETIME)

All timestamps use SQLite's DATETIME type with default `CURRENT_TIMESTAMP`. In Go, these map to `time.Time`.

### Money Values (REAL)

Monetary values (prices, costs) are stored as REAL (float64). For applications requiring precise decimal arithmetic, consider using INTEGER with cents.

## Migrations

The schema is managed through versioned migrations:

| Version | Name | Description |
|---------|------|-------------|
| 000001 | create_users_table | Creates users table with indexes |
| 000002 | create_vehicles_table | Creates vehicles table with FK to users |
| 000003 | create_maintenance_records_table | Creates maintenance_records table with FK to vehicles |

For migration management details, see [Database Migrations Guide](./database-migrations.md).

## Design Decisions

### SQLite Choice

SQLite was chosen for:
- **Simplicity**: No separate database server required
- **Portability**: Single file database, easy to backup
- **Performance**: Excellent read performance for small-to-medium datasets
- **Zero Configuration**: Works out of the box

### UUID Primary Keys

UUIDs are used instead of auto-increment integers for:
- **Distributed Generation**: IDs can be generated in application code
- **No Collisions**: Safe for distributed systems
- **Security**: IDs are not sequential/predictable

### Cascade Deletes

`ON DELETE CASCADE` is used to:
- **Maintain Integrity**: Automatically clean up child records
- **Simplify Code**: No need for manual cleanup logic
- **Prevent Orphans**: No orphaned vehicles or maintenance records

### Index Strategy

Indexes are added for:
- **Foreign Keys**: `user_id`, `vehicle_id` for join performance
- **Unique Lookups**: `email`, `username`, `vin` for authentication and search
- **Filtering**: `status`, `service_date` for common filter operations

## Future Considerations

### Planned Tables

1. **fuel_records** - Track fuel purchases and consumption
2. **reminders** - Service reminders and scheduled maintenance

### Potential Improvements

1. **Full-Text Search**: Add FTS5 for searching notes and descriptions
2. **Audit Log**: Track changes to vehicles and records
3. **Attachments**: Store service receipts and documents

## See Also

- [Database Migrations Guide](./database-migrations.md)
- [Database Setup Guide](./database-setup.md)
- [Repository Interfaces](./repositories.md)
- [Data Validation Rules](./validation.md)
