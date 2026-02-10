package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupDatabase(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test database
	srcDBPath := filepath.Join(tmpDir, "test.db")
	db, err := InitializeGarage(srcDBPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer db.Terminate()

	// Create some test data by creating a simple table
	_, err = db.RawSQLConnection().Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT)")
	require.NoError(t, err)
	_, err = db.RawSQLConnection().Exec("INSERT INTO test (value) VALUES ('test data')")
	require.NoError(t, err)

	// Close the database to ensure all data is flushed
	err = db.Terminate()
	require.NoError(t, err)

	// Create backup directory
	backupDir := filepath.Join(tmpDir, "backups")
	destDBPath := filepath.Join(backupDir, "test-backup.db")

	// Perform backup
	err = BackupDatabase(srcDBPath, destDBPath)
	require.NoError(t, err)

	// Verify backup file exists
	_, err = os.Stat(destDBPath)
	assert.NoError(t, err, "Backup file should exist")

	// Verify backup is not empty
	fileInfo, err := os.Stat(destDBPath)
	require.NoError(t, err)
	assert.Greater(t, fileInfo.Size(), int64(0), "Backup file should not be empty")

	// Open backup database and verify data
	backupDB, err := InitializeGarage(destDBPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer backupDB.Terminate()

	var value string
	err = backupDB.RawSQLConnection().QueryRow("SELECT value FROM test WHERE id = 1").Scan(&value)
	require.NoError(t, err)
	assert.Equal(t, "test data", value, "Backup should contain the same data as source")
}

func TestBackupDatabase_NonExistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	srcDBPath := filepath.Join(tmpDir, "nonexistent.db")
	destDBPath := filepath.Join(tmpDir, "backup.db")

	err := BackupDatabase(srcDBPath, destDBPath)
	assert.Error(t, err, "Should fail when source database doesn't exist")
	assert.Contains(t, err.Error(), "does not exist")
}

func TestBackupDatabase_CreateDestinationDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test database
	srcDBPath := filepath.Join(tmpDir, "test.db")
	db, err := InitializeGarage(srcDBPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer db.Terminate()

	// Create destination in a non-existent directory
	destDBPath := filepath.Join(tmpDir, "deep", "nested", "path", "backup.db")

	// Perform backup (should create nested directories)
	err = BackupDatabase(srcDBPath, destDBPath)
	require.NoError(t, err)

	// Verify backup file exists
	_, err = os.Stat(destDBPath)
	assert.NoError(t, err, "Backup file should exist in nested directory")
}

func TestCreateTimestampedBackup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test database
	srcDBPath := filepath.Join(tmpDir, "test.db")
	db, err := InitializeGarage(srcDBPath, StandardWorkerPoolSettings())
	require.NoError(t, err)

	// Create some test data
	_, err = db.RawSQLConnection().Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT)")
	require.NoError(t, err)
	_, err = db.RawSQLConnection().Exec("INSERT INTO test (value) VALUES ('test data')")
	require.NoError(t, err)

	err = db.Terminate()
	require.NoError(t, err)

	// Create backup directory
	backupDir := filepath.Join(tmpDir, "backups")

	// Create timestamped backup
	backupPath, err := CreateTimestampedBackup(srcDBPath, backupDir)
	require.NoError(t, err)

	// Verify backup path is returned
	assert.NotEmpty(t, backupPath, "Backup path should be returned")

	// Verify backup file exists
	_, err = os.Stat(backupPath)
	assert.NoError(t, err, "Backup file should exist")

	// Verify filename contains timestamp
	filename := filepath.Base(backupPath)
	assert.Contains(t, filename, "backup-", "Filename should contain 'backup-'")
	assert.Contains(t, filename, "test", "Filename should contain original database name")

	// Verify backup is not empty
	fileInfo, err := os.Stat(backupPath)
	require.NoError(t, err)
	assert.Greater(t, fileInfo.Size(), int64(0), "Backup file should not be empty")

	// Open backup database and verify data
	backupDB, err := InitializeGarage(backupPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer backupDB.Terminate()

	var value string
	err = backupDB.RawSQLConnection().QueryRow("SELECT value FROM test WHERE id = 1").Scan(&value)
	require.NoError(t, err)
	assert.Equal(t, "test data", value, "Backup should contain the same data as source")
}

func TestCreateTimestampedBackup_MultipleBackups(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test database
	srcDBPath := filepath.Join(tmpDir, "test.db")
	db, err := InitializeGarage(srcDBPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer db.Terminate()

	// Create backup directory
	backupDir := filepath.Join(tmpDir, "backups")

	// Create first backup
	backup1Path, err := CreateTimestampedBackup(srcDBPath, backupDir)
	require.NoError(t, err)

	// Wait a bit to ensure different timestamp
	time.Sleep(1 * time.Second)

	// Create second backup
	backup2Path, err := CreateTimestampedBackup(srcDBPath, backupDir)
	require.NoError(t, err)

	// Verify both backups exist and have different names
	assert.NotEqual(t, backup1Path, backup2Path, "Backups should have different filenames")

	_, err = os.Stat(backup1Path)
	assert.NoError(t, err, "First backup should exist")

	_, err = os.Stat(backup2Path)
	assert.NoError(t, err, "Second backup should exist")
}

func TestCreateTimestampedBackup_NonExistentBackupDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test database
	srcDBPath := filepath.Join(tmpDir, "test.db")
	db, err := InitializeGarage(srcDBPath, StandardWorkerPoolSettings())
	require.NoError(t, err)
	defer db.Terminate()

	// Use a non-existent backup directory
	backupDir := filepath.Join(tmpDir, "nonexistent", "backups")

	// Create timestamped backup (should create the directory)
	backupPath, err := CreateTimestampedBackup(srcDBPath, backupDir)
	require.NoError(t, err)

	// Verify backup exists
	_, err = os.Stat(backupPath)
	assert.NoError(t, err, "Backup should exist even when directory didn't exist initially")
}
