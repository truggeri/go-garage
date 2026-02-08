package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// SchemaEvolver manages database schema changes over time
type SchemaEvolver struct {
	garageDB          *sql.DB
	migrationFilesDir string
}

// NewSchemaEvolver creates an evolver for managing schema migrations
func NewSchemaEvolver(migrationsDirectory string, dbConnection *sql.DB) *SchemaEvolver {
	return &SchemaEvolver{
		garageDB:          dbConnection,
		migrationFilesDir: migrationsDirectory,
	}
}

// EvolveToLatest applies all pending schema changes
func (se *SchemaEvolver) EvolveToLatest() error {
	migrator, buildErr := se.constructMigrator()
	if buildErr != nil {
		return fmt.Errorf("migrator construction failed: %w", buildErr)
	}

	// Note: Don't defer migrator.Close() as it closes the underlying DB connection
	// which we don't want since the connection is managed externally

	upErr := migrator.Up()
	if upErr != nil && !errors.Is(upErr, migrate.ErrNoChange) {
		return fmt.Errorf("schema evolution failed: %w", upErr)
	}

	return nil
}

// RollbackAll reverts all schema changes
func (se *SchemaEvolver) RollbackAll() error {
	migrator, buildErr := se.constructMigrator()
	if buildErr != nil {
		return fmt.Errorf("migrator construction failed: %w", buildErr)
	}

	downErr := migrator.Down()
	if downErr != nil && !errors.Is(downErr, migrate.ErrNoChange) {
		return fmt.Errorf("schema rollback failed: %w", downErr)
	}

	return nil
}

// RollbackOne reverts the most recent schema change
func (se *SchemaEvolver) RollbackOne() error {
	migrator, buildErr := se.constructMigrator()
	if buildErr != nil {
		return fmt.Errorf("migrator construction failed: %w", buildErr)
	}

	stepErr := migrator.Steps(-1)
	if stepErr != nil && !errors.Is(stepErr, migrate.ErrNoChange) {
		return fmt.Errorf("single rollback failed: %w", stepErr)
	}

	return nil
}

// EvolveToSpecificVersion migrates to a particular schema version
func (se *SchemaEvolver) EvolveToSpecificVersion(targetVersion uint) error {
	migrator, buildErr := se.constructMigrator()
	if buildErr != nil {
		return fmt.Errorf("migrator construction failed: %w", buildErr)
	}

	migrateErr := migrator.Migrate(targetVersion)
	if migrateErr != nil && !errors.Is(migrateErr, migrate.ErrNoChange) {
		return fmt.Errorf("evolution to version %d failed: %w", targetVersion, migrateErr)
	}

	return nil
}

// CurrentSchemaVersion returns the current version and dirty state
func (se *SchemaEvolver) CurrentSchemaVersion() (uint, bool, error) {
	migrator, buildErr := se.constructMigrator()
	if buildErr != nil {
		return 0, false, fmt.Errorf("migrator construction failed: %w", buildErr)
	}

	version, isDirty, versionErr := migrator.Version()
	if versionErr != nil && !errors.Is(versionErr, migrate.ErrNilVersion) {
		return 0, false, fmt.Errorf("version check failed: %w", versionErr)
	}

	return version, isDirty, nil
}

// constructMigrator builds a migration engine instance
func (se *SchemaEvolver) constructMigrator() (*migrate.Migrate, error) {
	sqliteDriver, driverErr := sqlite3.WithInstance(se.garageDB, &sqlite3.Config{})
	if driverErr != nil {
		return nil, fmt.Errorf("SQLite driver creation failed: %w", driverErr)
	}

	absolutePath, pathErr := filepath.Abs(se.migrationFilesDir)
	if pathErr != nil {
		return nil, fmt.Errorf("absolute path resolution failed for %s: %w", se.migrationFilesDir, pathErr)
	}

	fileSourceURL := fmt.Sprintf("file://%s", absolutePath)

	migrationEngine, engineErr := migrate.NewWithDatabaseInstance(fileSourceURL, "sqlite3", sqliteDriver)
	if engineErr != nil {
		return nil, fmt.Errorf("migration engine creation failed: %w", engineErr)
	}

	return migrationEngine, nil
}

// BootstrapSchema is a helper that runs migrations during application startup
func BootstrapSchema(ctx context.Context, garage *SQLiteGarage, migrationsFolder string) error {
	if garage == nil || garage.underlyingDB == nil {
		return fmt.Errorf("invalid garage provided for schema bootstrap")
	}

	evolver := NewSchemaEvolver(migrationsFolder, garage.underlyingDB)

	if evolveErr := evolver.EvolveToLatest(); evolveErr != nil {
		return fmt.Errorf("schema bootstrap failed: %w", evolveErr)
	}

	return nil
}
