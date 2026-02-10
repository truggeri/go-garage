package database

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// BackupDatabase creates a copy of the SQLite database file from sourcePath to destPath.
// This function handles SQLite WAL mode by copying the main database file, WAL file, and
// SHM file if they exist. It ensures data consistency by performing a checkpoint operation
// before copying.
func BackupDatabase(sourcePath, destPath string) error {
	// Verify source database exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source database does not exist: %s", sourcePath)
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// For SQLite in WAL mode, we need to checkpoint before backing up
	// This ensures all WAL data is written to the main database file
	if err := checkpointDatabase(sourcePath); err != nil {
		return fmt.Errorf("failed to checkpoint database: %w", err)
	}

	// Copy the main database file
	if err := copyFile(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to copy database file: %w", err)
	}

	// Copy WAL and SHM files if they exist (though checkpoint should have removed them)
	walSource := sourcePath + "-wal"
	if _, err := os.Stat(walSource); err == nil {
		walDest := destPath + "-wal"
		if err := copyFile(walSource, walDest); err != nil {
			// Non-fatal, log but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to copy WAL file: %v\n", err)
		}
	}

	shmSource := sourcePath + "-shm"
	if _, err := os.Stat(shmSource); err == nil {
		shmDest := destPath + "-shm"
		if err := copyFile(shmSource, shmDest); err != nil {
			// Non-fatal, log but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to copy SHM file: %v\n", err)
		}
	}

	return nil
}

// CreateTimestampedBackup creates a backup of the database with a timestamp in the filename.
// Returns the full path to the created backup file.
func CreateTimestampedBackup(dbPath, backupDir string) (string, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Get the base filename without extension
	dbFilename := filepath.Base(dbPath)
	ext := filepath.Ext(dbFilename)
	nameWithoutExt := dbFilename[:len(dbFilename)-len(ext)]

	// Create timestamped filename
	timestamp := time.Now().Format("20060102-150405")
	backupFilename := fmt.Sprintf("%s-backup-%s%s", nameWithoutExt, timestamp, ext)
	backupPath := filepath.Join(backupDir, backupFilename)

	// Perform the backup
	if err := BackupDatabase(dbPath, backupPath); err != nil {
		return "", fmt.Errorf("backup failed: %w", err)
	}

	return backupPath, nil
}

// checkpointDatabase performs a checkpoint operation on the SQLite database
// to ensure all WAL data is written to the main database file
func checkpointDatabase(dbPath string) error {
	// Open a connection to the database
	db, err := InitializeGarage(dbPath, StandardWorkerPoolSettings())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Terminate()

	// Execute checkpoint
	_, err = db.RawSQLConnection().Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	if err != nil {
		return fmt.Errorf("checkpoint execution failed: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the file contents
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Ensure all data is written to disk
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}
