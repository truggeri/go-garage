package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteGarage wraps database operations for the vehicle garage system
type SQLiteGarage struct {
	underlyingDB      *sql.DB
	filePath          string
	workerPoolLimits  WorkerPoolSettings
	operationalStatus bool
}

// WorkerPoolSettings controls how many concurrent database workers are allowed
type WorkerPoolSettings struct {
	MaxActiveWorkers  int
	MaxIdleWorkers    int
	WorkerLifespan    time.Duration
	IdleWorkerTimeout time.Duration
}

// StandardWorkerPoolSettings returns typical settings for vehicle database operations
func StandardWorkerPoolSettings() WorkerPoolSettings {
	return WorkerPoolSettings{
		MaxActiveWorkers:  25,
		MaxIdleWorkers:    5,
		WorkerLifespan:    time.Hour,
		IdleWorkerTimeout: 10 * time.Minute,
	}
}

// InitializeGarage sets up the SQLite database for vehicle management
func InitializeGarage(dbFilePath string, workerSettings WorkerPoolSettings) (*SQLiteGarage, error) {
	parentDir := filepath.Dir(dbFilePath)
	if parentDir != "." && parentDir != "" {
		if mkdirErr := os.MkdirAll(parentDir, 0750); mkdirErr != nil {
			return nil, fmt.Errorf("cannot create database directory %s: %w", parentDir, mkdirErr)
		}
	}

	dataSourceName := buildSQLiteDataSource(dbFilePath)

	dbHandle, openErr := sql.Open("sqlite3", dataSourceName)
	if openErr != nil {
		return nil, fmt.Errorf("database opening failed at %s: %w", dbFilePath, openErr)
	}

	dbHandle.SetMaxOpenConns(workerSettings.MaxActiveWorkers)
	dbHandle.SetMaxIdleConns(workerSettings.MaxIdleWorkers)
	dbHandle.SetConnMaxLifetime(workerSettings.WorkerLifespan)
	dbHandle.SetConnMaxIdleTime(workerSettings.IdleWorkerTimeout)

	garage := &SQLiteGarage{
		underlyingDB:      dbHandle,
		filePath:          dbFilePath,
		workerPoolLimits:  workerSettings,
		operationalStatus: false,
	}

	if pragmaErr := garage.applySQLiteTuning(); pragmaErr != nil {
		dbHandle.Close()
		return nil, fmt.Errorf("SQLite tuning failed: %w", pragmaErr)
	}

	garage.operationalStatus = true
	return garage, nil
}

// buildSQLiteDataSource constructs the connection string with necessary flags
func buildSQLiteDataSource(targetPath string) string {
	return fmt.Sprintf("file:%s?_fk=1&_journal=WAL&_timeout=5000&_sync=1", targetPath)
}

// applySQLiteTuning applies performance and safety settings
func (sg *SQLiteGarage) applySQLiteTuning() error {
	tuningCommands := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = -64000",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA mmap_size = 30000000000",
	}

	timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	for _, command := range tuningCommands {
		if _, execErr := sg.underlyingDB.ExecContext(timeoutCtx, command); execErr != nil {
			return fmt.Errorf("tuning command '%s' failed: %w", command, execErr)
		}
	}

	return nil
}

// RawSQLConnection exposes the underlying database handle for direct queries
func (sg *SQLiteGarage) RawSQLConnection() *sql.DB {
	return sg.underlyingDB
}

// DiagnoseHealth checks if the garage database is responsive and working
func (sg *SQLiteGarage) DiagnoseHealth(ctx context.Context) error {
	if !sg.operationalStatus {
		return fmt.Errorf("garage database is not operational")
	}

	if pingErr := sg.underlyingDB.PingContext(ctx); pingErr != nil {
		return fmt.Errorf("health check ping failed: %w", pingErr)
	}

	var magicNumber int
	diagnosticQuery := "SELECT 42"
	scanErr := sg.underlyingDB.QueryRowContext(ctx, diagnosticQuery).Scan(&magicNumber)
	if scanErr != nil {
		return fmt.Errorf("diagnostic query execution failed: %w", scanErr)
	}

	if magicNumber != 42 {
		return fmt.Errorf("diagnostic query returned wrong value: %d instead of 42", magicNumber)
	}

	return nil
}

// Terminate shuts down the database connection cleanly
func (sg *SQLiteGarage) Terminate() error {
	if sg.underlyingDB == nil {
		return nil
	}

	if closeErr := sg.underlyingDB.Close(); closeErr != nil {
		return fmt.Errorf("database termination failed: %w", closeErr)
	}

	sg.operationalStatus = false
	return nil
}

// DatabaseFilePath returns where the database file is stored
func (sg *SQLiteGarage) DatabaseFilePath() string {
	return sg.filePath
}

// WorkerPoolStatistics returns metrics about connection pool usage
func (sg *SQLiteGarage) WorkerPoolStatistics() sql.DBStats {
	return sg.underlyingDB.Stats()
}

// IsOperational returns whether the garage database is ready for use
func (sg *SQLiteGarage) IsOperational() bool {
	return sg.operationalStatus
}
