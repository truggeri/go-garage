package repositories

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database connection
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?_fk=1&_journal=WAL")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

// runMigrations applies the database schema
func runMigrations(db *sql.DB) error {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			first_name TEXT,
			last_name TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_login_at DATETIME
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`)
	if err != nil {
		return err
	}

	// Create vehicles table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS vehicles (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			display_name TEXT,
			vin TEXT NOT NULL UNIQUE,
			make TEXT NOT NULL,
			model TEXT NOT NULL,
			year INTEGER NOT NULL,
			color TEXT,
			license_plate TEXT,
			purchase_date DATE,
			purchase_price REAL,
			purchase_mileage INTEGER,
			current_mileage INTEGER,
			status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'sold', 'scrapped')),
			notes TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_vehicles_user_id ON vehicles(user_id)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_vehicles_vin ON vehicles(vin)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_vehicles_status ON vehicles(status)`)
	if err != nil {
		return err
	}

	// Create maintenance_records table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS maintenance_records (
			id TEXT PRIMARY KEY,
			vehicle_id TEXT NOT NULL,
			service_type TEXT NOT NULL,
			custom_service_type TEXT DEFAULT '',
			service_date DATE NOT NULL,
			mileage_at_service INTEGER,
			cost REAL,
			service_provider TEXT,
			notes TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_maintenance_vehicle_id ON maintenance_records(vehicle_id)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_maintenance_service_date ON maintenance_records(service_date)`)
	if err != nil {
		return err
	}

	// Create vehicle_metrics table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS vehicle_metrics (
			vehicle_id TEXT PRIMARY KEY,
			total_spent REAL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create fuel_records table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS fuel_records (
			id TEXT PRIMARY KEY,
			vehicle_id TEXT NOT NULL,
			fill_date DATE NOT NULL,
			mileage INTEGER NOT NULL,
			volume REAL NOT NULL,
			fuel_type TEXT NOT NULL DEFAULT 'gasoline' CHECK(fuel_type IN ('gasoline', 'diesel', 'e85')),
			partial_fill INTEGER NOT NULL DEFAULT 0,
			price_per_unit REAL,
			octane_rating INTEGER,
			location TEXT,
			brand TEXT,
			notes TEXT,
			city_driving_percentage INTEGER CHECK(city_driving_percentage IS NULL OR (city_driving_percentage >= 0 AND city_driving_percentage <= 100)),
			vehicle_reported_mpg REAL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_fuel_records_vehicle_id ON fuel_records(vehicle_id)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_fuel_records_fill_date ON fuel_records(fill_date)`)
	if err != nil {
		return err
	}

	return nil
}
