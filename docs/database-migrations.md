# Database Migrations Guide

## Overview

Go-Garage uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema migrations. This guide explains how the migration system works and how to rollback migrations when needed.

## Migration System

### How It Works

The application uses a migration-based approach to evolve the database schema:

1. **Migration Files**: Located in the `migrations/` directory
2. **Naming Convention**: `{version}_{description}.{up|down}.sql`
   - `version`: Sequential number (e.g., `000001`, `000002`)
   - `description`: Short description of the migration
   - `up.sql`: Forward migration (applies changes)
   - `down.sql`: Reverse migration (reverts changes)

3. **Automatic Execution**: Migrations run automatically on application startup via `database.BootstrapSchema()`

### Current Migrations

The following migrations are currently in place:

1. **000001_create_users_table**: Creates the `users` table with indexes
2. **000002_create_vehicles_table**: Creates the `vehicles` table with foreign key to users
3. **000003_create_maintenance_records_table**: Creates the `maintenance_records` table with foreign key to vehicles
4. **000004_add_display_name_to_vehicles**: Adds `display_name` column to vehicles table
5. **000005_add_custom_service_type_to_maintenance**: Adds `custom_service_type` column to maintenance_records table for the "Other" service type option
6. **000006_create_vehicle_metrics_table**: Creates the `vehicle_metrics` table for aggregated vehicle statistics
7. **000007_create_fuel_records_table**: Creates the `fuel_records` table with foreign key to vehicles, supporting fuel fill-up tracking

### Schema Evolution

The `SchemaEvolver` type in `internal/database/schema.go` provides methods for managing migrations:

- `EvolveToLatest()`: Apply all pending migrations
- `RollbackAll()`: Revert all migrations
- `RollbackOne()`: Revert the most recent migration
- `EvolveToSpecificVersion(version)`: Migrate to a specific version
- `CurrentSchemaVersion()`: Get current version and dirty state

## Rolling Back Migrations

### When to Rollback

You should consider rolling back migrations in these situations:

1. **Development Issues**: A new migration causes problems during development
2. **Testing**: Need to reset database to a previous state for testing
3. **Deployment Rollback**: Application rollback requires schema rollback
4. **Migration Errors**: A migration partially applied and left the database in a "dirty" state

### Method 1: Using Code (Programmatic Rollback)

You can use the `SchemaEvolver` API to rollback migrations programmatically:

```go
package main

import (
    "context"
    "fmt"
    "github.com/truggeri/go-garage/internal/database"
)

func main() {
    // Initialize database
    db, err := database.InitializeGarage("./data/go-garage.db", database.StandardWorkerPoolSettings())
    if err != nil {
        panic(err)
    }
    defer db.Terminate()

    // Create schema evolver
    evolver := database.NewSchemaEvolver("./migrations", db.RawSQLConnection())

    // Option 1: Rollback one migration
    if err := evolver.RollbackOne(); err != nil {
        fmt.Printf("Failed to rollback: %v\n", err)
    }

    // Option 2: Rollback all migrations
    if err := evolver.RollbackAll(); err != nil {
        fmt.Printf("Failed to rollback all: %v\n", err)
    }

    // Option 3: Migrate to specific version
    if err := evolver.EvolveToSpecificVersion(1); err != nil {
        fmt.Printf("Failed to migrate to version 1: %v\n", err)
    }

    // Check current version
    version, isDirty, err := evolver.CurrentSchemaVersion()
    if err != nil {
        fmt.Printf("Failed to get version: %v\n", err)
    }
    fmt.Printf("Current version: %d, Dirty: %v\n", version, isDirty)
}
```

### Method 2: Using migrate CLI Tool

If you have the `migrate` CLI tool installed, you can use it directly:

```bash
# Install migrate CLI (if not already installed)
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Check current version
migrate -database "sqlite3://./data/go-garage.db" -path ./migrations version

# Rollback one migration
migrate -database "sqlite3://./data/go-garage.db" -path ./migrations down 1

# Rollback all migrations
migrate -database "sqlite3://./data/go-garage.db" -path ./migrations down

# Migrate to specific version
migrate -database "sqlite3://./data/go-garage.db" -path ./migrations goto 2

# Apply all pending migrations
migrate -database "sqlite3://./data/go-garage.db" -path ./migrations up
```

### Method 3: Manual SQL Rollback

In emergency situations, you can manually execute the down migration SQL:

```bash
# 1. Connect to the database
sqlite3 ./data/go-garage.db

# 2. Check current schema_migrations table
SELECT * FROM schema_migrations;

# 3. Manually execute the down migration
-- For example, to rollback migration 3:
.read migrations/000003_create_maintenance_records_table.down.sql

# 4. Update the schema_migrations table
DELETE FROM schema_migrations WHERE version = 3;

# 5. Exit sqlite
.quit
```

**⚠️ Warning**: Manual rollback should only be used as a last resort. Always prefer using the `SchemaEvolver` API or migrate CLI.

## Common Rollback Scenarios

### Scenario 1: Rollback Last Migration During Development

```bash
# Using migrate CLI
migrate -database "sqlite3://./data/go-garage.db" -path ./migrations down 1

# Or create a small Go script
go run - <<EOF
package main
import (
    "github.com/truggeri/go-garage/internal/database"
)
func main() {
    db, _ := database.InitializeGarage("./data/go-garage.db", database.StandardWorkerPoolSettings())
    defer db.Terminate()
    evolver := database.NewSchemaEvolver("./migrations", db.RawSQLConnection())
    evolver.RollbackOne()
}
EOF
```

### Scenario 2: Reset Database to Clean State

```bash
# Remove database file
rm -f ./data/go-garage.db*

# Restart application to recreate with latest schema
go run ./cmd/server
```

### Scenario 3: Fix Dirty Migration State

If a migration fails midway and leaves the database "dirty":

```bash
# 1. Check the dirty state
migrate -database "sqlite3://./data/go-garage.db" -path ./migrations version

# 2. Force to a specific version (use with caution)
migrate -database "sqlite3://./data/go-garage.db" -path ./migrations force 2

# 3. Try to apply migrations again
migrate -database "sqlite3://./data/go-garage.db" -path ./migrations up
```

### Scenario 4: Rollback in Production

1. **Backup Database First** (Critical!)

   ```bash
   make db-backup
   # or
   go run ./cmd/dbutil backup
   ```

2. **Stop the Application**

   ```bash
   # Stop the running server
   pkill -f go-garage
   ```

3. **Rollback the Migration**

   ```bash
   # Using migrate CLI
   migrate -database "sqlite3://./data/go-garage.db" -path ./migrations down 1
   ```

4. **Verify the Rollback**

   ```bash
   migrate -database "sqlite3://./data/go-garage.db" -path ./migrations version
   ```

5. **Restart with Previous Application Version**

   ```bash
   # Deploy and start previous version
   git checkout <previous-version>
   go run ./cmd/server
   ```

## Best Practices

1. **Always Backup Before Rollback**: Use `make db-backup` or `go run ./cmd/dbutil backup`
2. **Test Migrations**: Test both up and down migrations in development
3. **Write Reversible Migrations**: Ensure every up migration has a corresponding down migration
4. **Avoid Data Loss**: Down migrations that drop tables will lose data - consider archival strategies
5. **Version Control**: Keep migration files in version control
6. **Sequential Versions**: Never skip version numbers
7. **Production Rollbacks**: Plan and test rollback procedures before deploying

## Migration File Examples

### Creating a New Migration

```bash
# Create new migration files
touch migrations/000004_add_fuel_records_table.up.sql
touch migrations/000004_add_fuel_records_table.down.sql
```

**Up Migration** (`000004_add_fuel_records_table.up.sql`):

```sql
CREATE TABLE IF NOT EXISTS fuel_records (
    id TEXT PRIMARY KEY,
    vehicle_id TEXT NOT NULL,
    date DATE NOT NULL,
    odometer INTEGER,
    gallons REAL NOT NULL,
    price_per_gallon REAL,
    total_cost REAL,
    fuel_type TEXT,
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX idx_fuel_vehicle_id ON fuel_records(vehicle_id);
CREATE INDEX idx_fuel_date ON fuel_records(date);
```

**Down Migration** (`000004_add_fuel_records_table.down.sql`):

```sql
DROP INDEX IF EXISTS idx_fuel_date;
DROP INDEX IF EXISTS idx_fuel_vehicle_id;
DROP TABLE IF EXISTS fuel_records;
```

## Troubleshooting

### Problem: "Dirty database version"

**Cause**: A migration was interrupted or failed midway.

**Solution**:

1. Check which migration is dirty: `migrate version`
2. Manually fix the database state or use `migrate force <version>`
3. Re-run migrations

### Problem: "File does not exist"

**Cause**: Migration files are not in the expected location.

**Solution**: Ensure migrations are in the `./migrations` directory and the path is correct.

### Problem: "No change"

**Cause**: All migrations are already applied or rolled back.

**Solution**: This is not an error, just a status message.

## References

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [SQLite Foreign Keys](https://www.sqlite.org/foreignkeys.html)
- [Go-Garage Schema Evolution Code](../internal/database/schema.go)

## Helper Function: RollbackLastMigration

For convenience, you can use this helper function in your code:

```go
// RollbackLastMigration rolls back the most recent migration
func RollbackLastMigration(ctx context.Context, db *database.SQLiteGarage, migrationsPath string) error {
    evolver := database.NewSchemaEvolver(migrationsPath, db.RawSQLConnection())
    
    // Get current version
    version, isDirty, err := evolver.CurrentSchemaVersion()
    if err != nil {
        return fmt.Errorf("failed to get current version: %w", err)
    }
    
    if isDirty {
        return fmt.Errorf("database is in dirty state, cannot rollback automatically")
    }
    
    if version == 0 {
        return fmt.Errorf("no migrations to rollback")
    }
    
    // Rollback one step
    if err := evolver.RollbackOne(); err != nil {
        return fmt.Errorf("rollback failed: %w", err)
    }
    
    return nil
}
```

## Summary

- Migrations are located in `migrations/` directory
- Use `SchemaEvolver` API for programmatic control
- Always backup before rollback in production
- Test both up and down migrations
- Keep migrations reversible and well-documented
