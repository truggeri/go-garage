# Database Setup Guide

## Overview

Go-Garage uses SQLite as its database. This guide explains how to set up, configure, and manage the database for development and production environments.

## Prerequisites

- Go 1.24 or later
- SQLite3 (typically pre-installed on most systems)
- Make (for using Makefile commands)

## Quick Start

### Development Setup

1. **Clone and build the project:**
   ```bash
   git clone https://github.com/truggeri/go-garage.git
   cd go-garage
   go mod download
   ```

2. **Create data directory:**
   ```bash
   mkdir -p data
   ```

3. **Start the server (creates database automatically):**
   ```bash
   make run
   # or
   go run ./cmd/server
   ```

The database will be created at `./data/go-garage.db` with all migrations applied.

### Docker Setup

With Docker, the database is created inside the container:

```bash
docker-compose up -d
```

Data is persisted in a Docker volume (`garage-data`) at `/app/data/go-garage.db`.

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_PATH` | `./data/go-garage.db` | Path to SQLite database file |
| `DB_MAX_OPEN_CONNS` | `25` | Maximum number of open connections |
| `DB_MAX_IDLE_CONNS` | `5` | Maximum number of idle connections |
| `DB_CONN_MAX_LIFETIME` | `5m` | Maximum connection lifetime |

### Example .env File

```bash
# Database configuration
DB_PATH=./data/go-garage.db
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

## Database Initialization

### Automatic Initialization

When the server starts, it automatically:

1. Creates the database file if it doesn't exist
2. Applies all pending migrations
3. Verifies database connectivity

```go
// From cmd/server/main.go
db, err := database.InitializeGarage(cfg.Database.Path, database.StandardWorkerPoolSettings())
if err != nil {
    log.Fatal("Failed to initialize database:", err)
}
defer db.Terminate()

// Apply migrations
if err := database.BootstrapSchema(migrationsPath, db.RawSQLConnection()); err != nil {
    log.Fatal("Failed to apply migrations:", err)
}
```

### Manual Initialization

To manually initialize the database:

```bash
# Create database directory
mkdir -p data

# Start server briefly to create database
go run ./cmd/server &
sleep 2
kill $!

# Or use sqlite3 directly
sqlite3 data/go-garage.db ".databases"
```

## Schema Management

### Viewing Current Schema

```bash
# Connect to database
sqlite3 data/go-garage.db

# List all tables
.tables

# Show schema for a specific table
.schema users
.schema vehicles
.schema maintenance_records

# Show full schema
.schema

# Exit
.quit
```

### Checking Migration Version

```bash
# Using sqlite3
sqlite3 data/go-garage.db "SELECT * FROM schema_migrations;"

# Output shows current version and dirty state
# version|dirty
# 3|0
```

### Applying Migrations

Migrations are applied automatically on startup. To manually apply:

```bash
# Using migrate CLI (if installed)
migrate -database "sqlite3://./data/go-garage.db" -path ./migrations up

# Or restart the server
make run
```

For detailed migration management, see [Database Migrations Guide](./database-migrations.md).

## Seeding Data

### Using the Seed Command

Populate the database with sample data for development:

```bash
# Using make
make seed

# Or directly
go run ./cmd/seed

# With custom database path
DB_PATH=./data/test.db go run ./cmd/seed
```

The seed command:
- Creates sample users
- Creates sample vehicles
- Creates sample maintenance records
- Creates sample fuel records
- Checks for existing data to avoid duplicates

### Manual Data Insertion

```bash
sqlite3 data/go-garage.db <<EOF
INSERT INTO users (id, username, email, password_hash, first_name, last_name, created_at, updated_at)
VALUES (
    'test-user-001',
    'testuser',
    'test@example.com',
    '\$2a\$10\$hashedpassword',
    'Test',
    'User',
    datetime('now'),
    datetime('now')
);
EOF
```

## Backup and Restore

### Creating Backups

```bash
# Using make
make db-backup

# Or using the utility command
go run ./cmd/dbutil backup

# Or using sqlite3
sqlite3 data/go-garage.db ".backup data/go-garage-backup.db"
```

Backups are created in the same directory with a timestamp suffix.

### Restoring from Backup

```bash
# Stop the server first
pkill -f go-garage

# Replace the database file
cp data/go-garage-backup-2024-01-15.db data/go-garage.db

# Restart the server
make run
```

## Database Maintenance

### Optimizing the Database

SQLite databases can be optimized with VACUUM:

```bash
sqlite3 data/go-garage.db "VACUUM;"
```

### Analyzing Query Performance

```bash
sqlite3 data/go-garage.db

# Enable query explain
.mode column
.headers on

# Analyze a query
EXPLAIN QUERY PLAN SELECT * FROM vehicles WHERE user_id = 'some-id';

# Check index usage
EXPLAIN QUERY PLAN SELECT * FROM users WHERE email = 'test@example.com';
```

### Database Integrity Check

```bash
sqlite3 data/go-garage.db "PRAGMA integrity_check;"
```

## Health Checks

### HTTP Health Endpoint

The application exposes a `/health` endpoint that includes database status:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
    "status": "healthy",
    "database": {
        "status": "connected",
        "latency_ms": 1
    }
}
```

### Programmatic Health Check

```go
health, err := db.DiagnoseHealth(ctx)
if err != nil {
    log.Error("Database health check failed:", err)
}
```

## Connection Pooling

### Default Settings

```go
// Standard settings
settings := database.StandardWorkerPoolSettings()
// MaxOpenConns: 25
// MaxIdleConns: 5
// ConnMaxLifetime: 5 minutes
```

### Custom Settings

```go
settings := database.WorkerPoolSettings{
    MaxOpenConns:    50,
    MaxIdleConns:    10,
    ConnMaxLifetime: 10 * time.Minute,
}
db, err := database.InitializeGarage(dbPath, settings)
```

## Common Tasks

### Resetting the Database

```bash
# Stop the server
pkill -f go-garage

# Delete the database file
rm -f data/go-garage.db*

# Restart (creates fresh database with migrations)
make run

# Optionally seed data
make seed
```

### Using a Different Database Path

```bash
# Set environment variable
export DB_PATH=/path/to/custom.db

# Or in .env file
echo "DB_PATH=/path/to/custom.db" >> .env

# Or via command line
DB_PATH=/path/to/custom.db go run ./cmd/server
```

### Testing with In-Memory Database

For testing, you can use an in-memory database:

```go
db, err := database.InitializeGarage(":memory:", database.StandardWorkerPoolSettings())
```

Note: In-memory databases are ephemeral and lost when the connection closes.

## Troubleshooting

### Problem: "database is locked"

**Cause:** Multiple processes or connections trying to write simultaneously.

**Solutions:**
1. Ensure only one instance of the server is running
2. Check for orphaned SQLite processes
3. Enable WAL mode (Write-Ahead Logging):
   ```bash
   sqlite3 data/go-garage.db "PRAGMA journal_mode=WAL;"
   ```

### Problem: "no such table"

**Cause:** Migrations haven't been applied.

**Solutions:**
1. Check if the database file exists
2. Restart the server to apply migrations
3. Verify migration files exist in `migrations/` directory
4. Check schema_migrations table for current version

### Problem: "disk I/O error"

**Cause:** File system issues or permissions.

**Solutions:**
1. Check disk space: `df -h`
2. Check file permissions: `ls -la data/`
3. Ensure the data directory exists and is writable
4. Check for disk errors

### Problem: "constraint failed"

**Cause:** Attempting to insert duplicate data or violating foreign key constraints.

**Solutions:**
1. Check for unique constraint violations (email, username, VIN)
2. Ensure referenced records exist (user_id, vehicle_id)
3. Use proper error handling to catch and report constraint violations

## SQLite-Specific Features

### WAL Mode

Go-Garage enables WAL mode for better concurrent read/write performance:

```go
// Enabled in database initialization
_, err := db.Exec("PRAGMA journal_mode=WAL;")
```

### Foreign Key Enforcement

Foreign keys are enabled by default:

```go
// Enabled in database initialization
_, err := db.Exec("PRAGMA foreign_keys = ON;")
```

### Connection String Options

The SQLite connection string supports various options:

```go
// Example with options
dsn := "file:./data/go-garage.db?cache=shared&mode=rwc"
```

Common options:
- `cache=shared` - Enable shared cache
- `mode=rwc` - Read-write-create mode
- `_journal_mode=WAL` - Enable WAL mode via DSN

## See Also

- [Database Schema](./database-schema.md) - Schema documentation and ERD
- [Database Migrations](./database-migrations.md) - Migration management guide
- [Repository Interfaces](./repositories.md) - Data access layer documentation
- [Docker Guide](./DOCKER.md) - Container-based deployment
- [Development Guide](./DEVELOPMENT.md) - General development setup
