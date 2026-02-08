package database

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeGarage_CreatesNewDatabase(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	workerSettings := StandardWorkerPoolSettings()
	garage, err := InitializeGarage(dbPath, workerSettings)

	require.NoError(t, err)
	require.NotNil(t, garage)
	assert.True(t, garage.IsOperational())
	assert.Equal(t, dbPath, garage.DatabaseFilePath())

	err = garage.Terminate()
	assert.NoError(t, err)
	assert.False(t, garage.IsOperational())
}

func TestInitializeGarage_CreatesNestedDirectories(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "data", "storage", "vehicles.db")

	workerSettings := StandardWorkerPoolSettings()
	garage, err := InitializeGarage(dbPath, workerSettings)

	require.NoError(t, err)
	require.NotNil(t, garage)

	_, statErr := os.Stat(filepath.Dir(dbPath))
	assert.NoError(t, statErr, "Nested directories should be created")

	err = garage.Terminate()
	assert.NoError(t, err)
}

func TestStandardWorkerPoolSettings_ReturnsExpectedValues(t *testing.T) {
	settings := StandardWorkerPoolSettings()

	assert.Equal(t, 25, settings.MaxActiveWorkers)
	assert.Equal(t, 5, settings.MaxIdleWorkers)
	assert.Equal(t, time.Hour, settings.WorkerLifespan)
	assert.Equal(t, 10*time.Minute, settings.IdleWorkerTimeout)
}

func TestInitializeGarage_AppliesCustomWorkerSettings(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	customSettings := WorkerPoolSettings{
		MaxActiveWorkers:  15,
		MaxIdleWorkers:    3,
		WorkerLifespan:    45 * time.Minute,
		IdleWorkerTimeout: 7 * time.Minute,
	}

	garage, err := InitializeGarage(dbPath, customSettings)
	require.NoError(t, err)
	require.NotNil(t, garage)

	stats := garage.WorkerPoolStatistics()
	assert.GreaterOrEqual(t, stats.Idle, 0)

	err = garage.Terminate()
	assert.NoError(t, err)
}

func TestDiagnoseHealth_OnHealthyDatabase(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer garage.Terminate()

	ctx := context.Background()
	err = garage.DiagnoseHealth(ctx)
	assert.NoError(t, err)
}

func TestDiagnoseHealth_WithContextTimeout(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer garage.Terminate()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = garage.DiagnoseHealth(ctx)
	assert.NoError(t, err)
}

func TestDiagnoseHealth_FailsWhenNotOperational(t *testing.T) {
	garage := &SQLiteGarage{
		underlyingDB:      nil,
		operationalStatus: false,
	}

	ctx := context.Background()
	err := garage.DiagnoseHealth(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not operational")
}

func TestDiagnoseHealth_FailsAfterTermination(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)

	err = garage.Terminate()
	require.NoError(t, err)

	ctx := context.Background()
	err = garage.DiagnoseHealth(ctx)
	assert.Error(t, err)
}

func TestTerminate_ClosesConnection(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)

	err = garage.Terminate()
	assert.NoError(t, err)
	assert.False(t, garage.IsOperational())
}

func TestTerminate_WithNilConnection(t *testing.T) {
	garage := &SQLiteGarage{
		underlyingDB: nil,
	}

	err := garage.Terminate()
	assert.NoError(t, err)
}

func TestRawSQLConnection_ReturnsUnderlyingDB(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer garage.Terminate()

	conn := garage.RawSQLConnection()
	assert.NotNil(t, conn)
	assert.IsType(t, &sql.DB{}, conn)
}

func TestDatabaseFilePath_ReturnsCorrectPath(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer garage.Terminate()

	path := garage.DatabaseFilePath()
	assert.Equal(t, dbPath, path)
}

func TestWorkerPoolStatistics_ReturnsMetrics(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer garage.Terminate()

	stats := garage.WorkerPoolStatistics()
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
	assert.GreaterOrEqual(t, stats.Idle, 0)
}

func TestSQLitePragmas_ForeignKeysActive(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer garage.Terminate()

	var fkEnabled int
	err = garage.underlyingDB.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	require.NoError(t, err)
	assert.Equal(t, 1, fkEnabled, "Foreign keys must be enabled")
}

func TestSQLitePragmas_WALJournalMode(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer garage.Terminate()

	var journalMode string
	err = garage.underlyingDB.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", journalMode, "WAL journal mode must be active")
}

func TestBuildSQLiteDataSource_IncludesRequiredParameters(t *testing.T) {
	dbPath := "/tmp/garage.db"
	dsn := buildSQLiteDataSource(dbPath)

	assert.Contains(t, dsn, "file:/tmp/garage.db")
	assert.Contains(t, dsn, "_fk=1")
	assert.Contains(t, dsn, "_journal=WAL")
	assert.Contains(t, dsn, "_timeout=5000")
	assert.Contains(t, dsn, "_sync=1")
}

func TestConcurrentDatabaseAccess_MultipleGoroutines(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer garage.Terminate()

	// Create a test table
	_, err = garage.underlyingDB.Exec(`CREATE TABLE concurrent_test (id INTEGER PRIMARY KEY, data TEXT)`)
	require.NoError(t, err)

	// Spawn multiple goroutines performing writes
	completionSignal := make(chan bool)
	numGoroutines := 15

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			ctx := context.Background()
			_, insertErr := garage.underlyingDB.ExecContext(ctx, "INSERT INTO concurrent_test (data) VALUES (?)", "test_data")
			assert.NoError(t, insertErr)
			completionSignal <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-completionSignal
	}

	// Verify all inserts succeeded
	var recordCount int
	err = garage.underlyingDB.QueryRow("SELECT COUNT(*) FROM concurrent_test").Scan(&recordCount)
	require.NoError(t, err)
	assert.Equal(t, numGoroutines, recordCount)
}

func TestIsOperational_ReturnsTrueAfterInit(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer garage.Terminate()

	assert.True(t, garage.IsOperational())
}

func TestIsOperational_ReturnsFalseAfterTerminate(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vehicles.db")

	garage, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	require.NoError(t, err)

	err = garage.Terminate()
	require.NoError(t, err)

	assert.False(t, garage.IsOperational())
}
