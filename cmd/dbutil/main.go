package main

import (
	"fmt"
	"os"

	"github.com/truggeri/go-garage/internal/database"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "backup":
		handleBackup()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Go-Garage Database Utilities")
	fmt.Println("=============================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dbutil <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  backup     Create a timestamped backup of the database")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  DB_PATH           Path to the database file (default: ./data/go-garage.db)")
	fmt.Println("  BACKUP_DIR        Directory for backups (default: ./data/backups)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  dbutil backup")
	fmt.Println("  DB_PATH=/path/to/db.db BACKUP_DIR=/path/to/backups dbutil backup")
}

func handleBackup() {
	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/go-garage.db"
	}

	// Get backup directory from environment or use default
	backupDir := os.Getenv("BACKUP_DIR")
	if backupDir == "" {
		backupDir = "./data/backups"
	}

	fmt.Println("Go-Garage Database Backup")
	fmt.Println("=========================")
	fmt.Println()
	fmt.Printf("Database path: %s\n", dbPath)
	fmt.Printf("Backup directory: %s\n", backupDir)
	fmt.Println()

	// Verify database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Database file does not exist: %s\n", dbPath)
		os.Exit(1)
	}

	// Create backup
	fmt.Println("Creating backup...")
	backupPath, err := database.CreateTimestampedBackup(dbPath, backupDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create backup: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Backup created successfully!")
	fmt.Printf("  Backup location: %s\n", backupPath)

	// Get file size for info
	fileInfo, err := os.Stat(backupPath)
	if err == nil {
		fmt.Printf("  Backup size: %.2f MB\n", float64(fileInfo.Size())/(1024*1024))
	}

	fmt.Println()
	fmt.Println("=========================")
}
