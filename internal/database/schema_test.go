package database

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGarageForMigrationTests(t *testing.T) (*SQLiteGarage, string) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "migrations_test.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)

	// migrations directory relative to test file location
	migrationsDir := filepath.Join("..", "..", "migrations")

	return garage, migrationsDir
}

func TestNewSchemaEvolver_CreatesInstance(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	evolver := NewSchemaEvolver(migrationsDir, garage.underlyingDB)
	assert.NotNil(t, evolver)
	assert.Equal(t, migrationsDir, evolver.migrationFilesDir)
	assert.Equal(t, garage.underlyingDB, evolver.garageDB)
}

func TestEvolveToLatest_AppliesAllMigrations(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	evolver := NewSchemaEvolver(migrationsDir, garage.underlyingDB)
	err := evolver.EvolveToLatest()
	assert.NoError(t, err)

	// Verify expected tables were created
	expectedTables := []string{"users", "vehicles", "maintenance_records"}
	for _, tableName := range expectedTables {
		var foundName string
		queryErr := garage.underlyingDB.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName,
		).Scan(&foundName)
		assert.NoError(t, queryErr, "Table %s should exist", tableName)
		assert.Equal(t, tableName, foundName)
	}
}

func TestEvolveToLatest_IdempotentOperation(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	evolver := NewSchemaEvolver(migrationsDir, garage.underlyingDB)

	// First evolution
	err := evolver.EvolveToLatest()
	assert.NoError(t, err)

	// Second evolution should not error
	err = evolver.EvolveToLatest()
	assert.NoError(t, err)
}

func TestCurrentSchemaVersion_BeforeMigrations(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	evolver := NewSchemaEvolver(migrationsDir, garage.underlyingDB)

	version, isDirty, err := evolver.CurrentSchemaVersion()
	assert.NoError(t, err)
	assert.False(t, isDirty)
	assert.Equal(t, uint(0), version)
}

func TestCurrentSchemaVersion_AfterMigrations(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	evolver := NewSchemaEvolver(migrationsDir, garage.underlyingDB)

	err := evolver.EvolveToLatest()
	require.NoError(t, err)

	version, isDirty, err := evolver.CurrentSchemaVersion()
	assert.NoError(t, err)
	assert.False(t, isDirty)
	assert.Greater(t, version, uint(0))
}

func TestRollbackAll_RemovesAllTables(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	evolver := NewSchemaEvolver(migrationsDir, garage.underlyingDB)

	// Apply migrations first
	err := evolver.EvolveToLatest()
	require.NoError(t, err)

	// Verify table exists
	var tableName string
	err = garage.underlyingDB.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='users'",
	).Scan(&tableName)
	require.NoError(t, err)

	// Rollback all
	err = evolver.RollbackAll()
	assert.NoError(t, err)

	// Verify table no longer exists
	err = garage.underlyingDB.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='users'",
	).Scan(&tableName)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestRollbackOne_RevertsOneMigration(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	evolver := NewSchemaEvolver(migrationsDir, garage.underlyingDB)

	// Apply all migrations
	err := evolver.EvolveToLatest()
	require.NoError(t, err)

	versionBefore, _, err := evolver.CurrentSchemaVersion()
	require.NoError(t, err)

	// Rollback one step
	err = evolver.RollbackOne()
	assert.NoError(t, err)

	versionAfter, _, err := evolver.CurrentSchemaVersion()
	assert.NoError(t, err)
	assert.Less(t, versionAfter, versionBefore)
}

func TestEvolveToSpecificVersion_MigratesToTargetVersion(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	evolver := NewSchemaEvolver(migrationsDir, garage.underlyingDB)

	// Migrate to version 2
	targetVer := uint(2)
	err := evolver.EvolveToSpecificVersion(targetVer)
	assert.NoError(t, err)

	version, isDirty, err := evolver.CurrentSchemaVersion()
	assert.NoError(t, err)
	assert.False(t, isDirty)
	assert.Equal(t, targetVer, version)

	// Verify only first two tables exist
	var tableName string
	err = garage.underlyingDB.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='users'",
	).Scan(&tableName)
	assert.NoError(t, err)

	err = garage.underlyingDB.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='vehicles'",
	).Scan(&tableName)
	assert.NoError(t, err)
}

func TestBootstrapSchema_AppliesAllMigrations(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	ctx := context.Background()
	err := BootstrapSchema(ctx, garage, migrationsDir)
	assert.NoError(t, err)

	// Verify tables were created
	expectedTables := []string{"users", "vehicles", "maintenance_records"}
	for _, tableName := range expectedTables {
		var foundName string
		queryErr := garage.underlyingDB.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName,
		).Scan(&foundName)
		assert.NoError(t, queryErr, "Table %s should exist", tableName)
	}
}

func TestBootstrapSchema_FailsWithNilGarage(t *testing.T) {
	ctx := context.Background()
	err := BootstrapSchema(ctx, nil, "./migrations")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid garage")
}

func TestBootstrapSchema_FailsWithNilUnderlyingDB(t *testing.T) {
	garage := &SQLiteGarage{
		underlyingDB: nil,
	}

	ctx := context.Background()
	err := BootstrapSchema(ctx, garage, "./migrations")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid garage")
}

func TestMigrations_EnforceForeignKeys(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	evolver := NewSchemaEvolver(migrationsDir, garage.underlyingDB)
	err := evolver.EvolveToLatest()
	require.NoError(t, err)

	// Insert a user
	_, err = garage.underlyingDB.Exec(`
		INSERT INTO users (id, username, email, password_hash) 
		VALUES ('usr_001', 'garageuser', 'user@garage.com', 'hashed_pw')
	`)
	require.NoError(t, err)

	// Attempt to insert vehicle with non-existent user_id - should fail
	_, err = garage.underlyingDB.Exec(`
		INSERT INTO vehicles (id, user_id, vin, make, model, year) 
		VALUES ('veh_001', 'nonexistent_user', 'ABC123XYZ', 'Honda', 'Accord', 2021)
	`)
	assert.Error(t, err, "Foreign key constraint must be enforced")
}

func TestMigrations_EnforceUniqueConstraints(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	evolver := NewSchemaEvolver(migrationsDir, garage.underlyingDB)
	err := evolver.EvolveToLatest()
	require.NoError(t, err)

	// Insert first user
	_, err = garage.underlyingDB.Exec(`
		INSERT INTO users (id, username, email, password_hash) 
		VALUES ('usr_001', 'garageuser', 'user@garage.com', 'hashed_pw')
	`)
	require.NoError(t, err)

	// Attempt duplicate username - should fail
	_, err = garage.underlyingDB.Exec(`
		INSERT INTO users (id, username, email, password_hash) 
		VALUES ('usr_002', 'garageuser', 'different@garage.com', 'other_pw')
	`)
	assert.Error(t, err, "Username unique constraint must be enforced")

	// Attempt duplicate email - should fail
	_, err = garage.underlyingDB.Exec(`
		INSERT INTO users (id, username, email, password_hash) 
		VALUES ('usr_003', 'differentuser', 'user@garage.com', 'another_pw')
	`)
	assert.Error(t, err, "Email unique constraint must be enforced")
}

func TestMigrations_CreatesExpectedIndexes(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	evolver := NewSchemaEvolver(migrationsDir, garage.underlyingDB)
	err := evolver.EvolveToLatest()
	require.NoError(t, err)

	// Check for expected indexes
	requiredIndexes := []string{
		"idx_users_email",
		"idx_users_username",
		"idx_vehicles_user_id",
		"idx_vehicles_vin",
		"idx_vehicles_status",
		"idx_maintenance_vehicle_id",
		"idx_maintenance_service_date",
	}

	for _, indexName := range requiredIndexes {
		var foundName string
		queryErr := garage.underlyingDB.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='index' AND name=?", indexName,
		).Scan(&foundName)
		assert.NoError(t, queryErr, "Index %s must exist", indexName)
	}
}

func TestMigrations_CascadeDeleteBehavior(t *testing.T) {
	garage, migrationsDir := setupGarageForMigrationTests(t)
	defer garage.Terminate()

	evolver := NewSchemaEvolver(migrationsDir, garage.underlyingDB)
	err := evolver.EvolveToLatest()
	require.NoError(t, err)

	// Insert user and associated vehicle
	_, err = garage.underlyingDB.Exec(`
		INSERT INTO users (id, username, email, password_hash) 
		VALUES ('usr_001', 'garageuser', 'user@garage.com', 'hashed_pw')
	`)
	require.NoError(t, err)

	_, err = garage.underlyingDB.Exec(`
		INSERT INTO vehicles (id, user_id, vin, make, model, year) 
		VALUES ('veh_001', 'usr_001', 'ABC123XYZ', 'Honda', 'Accord', 2021)
	`)
	require.NoError(t, err)

	// Delete the user
	_, err = garage.underlyingDB.Exec("DELETE FROM users WHERE id = 'usr_001'")
	require.NoError(t, err)

	// Verify vehicle was cascade deleted
	var vehicleCount int
	err = garage.underlyingDB.QueryRow("SELECT COUNT(*) FROM vehicles WHERE id = 'veh_001'").Scan(&vehicleCount)
	require.NoError(t, err)
	assert.Equal(t, 0, vehicleCount, "Vehicle should be cascade deleted")
}
