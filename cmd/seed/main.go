package main

import (
	"context"
	"fmt"
	"os"

	"github.com/truggeri/go-garage/internal/database"
	"github.com/truggeri/go-garage/internal/database/seeddata"
	"github.com/truggeri/go-garage/internal/repositories"
)

func main() {
	fmt.Println("Go-Garage Database Seeding Tool")
	fmt.Println("================================")
	fmt.Println()

	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/go-garage.db"
	}

	fmt.Printf("Database path: %s\n", dbPath)
	fmt.Println()

	// Initialize database connection
	fmt.Println("Connecting to database...")
	garageDB, err := database.InitializeGarage(dbPath, database.StandardWorkerPoolSettings())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer garageDB.Terminate()

	// Verify database connectivity
	ctx := context.Background()
	if err := garageDB.DiagnoseHealth(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Database health check failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Database connection established")

	// Run migrations
	fmt.Println()
	fmt.Println("Running database migrations...")
	migrationsPath := "./migrations"
	if err := database.BootstrapSchema(ctx, garageDB, migrationsPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Database migrations completed")

	// Initialize repositories
	db := garageDB.RawSQLConnection()
	userRepo := repositories.NewSQLiteUserRepository(db)
	vehicleRepo := repositories.NewSQLiteVehicleRepository(db)
	maintenanceRepo := repositories.NewSQLiteMaintenanceRepository(db)

	// Check if data already exists
	fmt.Println()
	fmt.Println("Checking for existing data...")
	existingDataFound := false

	// Check for existing users
	sampleUsers := seeddata.GetSampleUsers()
	for _, user := range sampleUsers {
		if existingUser, err := userRepo.FindByID(ctx, user.ID); err == nil && existingUser != nil {
			fmt.Printf("⚠ User already exists: %s (%s)\n", user.Username, user.ID)
			existingDataFound = true
		}
	}

	if existingDataFound {
		fmt.Println()
		fmt.Print("Some data already exists. Do you want to continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Seeding cancelled.")
			os.Exit(0)
		}
	} else {
		fmt.Println("✓ No existing seed data found")
	}

	// Seed users
	fmt.Println()
	fmt.Println("Seeding users...")
	userCount := 0
	for _, user := range sampleUsers {
		// Check if user already exists
		if existingUser, err := userRepo.FindByID(ctx, user.ID); err == nil && existingUser != nil {
			fmt.Printf("  - Skipping user %s (already exists)\n", user.Username)
			continue
		}

		if err := userRepo.Create(ctx, user); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Failed to create user %s: %v\n", user.Username, err)
			continue
		}
		fmt.Printf("  ✓ Created user: %s (%s %s)\n", user.Username, user.FirstName, user.LastName)
		userCount++
	}
	fmt.Printf("✓ Seeded %d users\n", userCount)

	// Seed vehicles
	fmt.Println()
	fmt.Println("Seeding vehicles...")
	vehicleCount := 0
	sampleVehicles := seeddata.GetSampleVehicles()
	for _, vehicle := range sampleVehicles {
		// Check if vehicle already exists
		if existingVehicle, err := vehicleRepo.FindByID(ctx, vehicle.ID); err == nil && existingVehicle != nil {
			fmt.Printf("  - Skipping vehicle %s %s (already exists)\n", vehicle.Make, vehicle.Model)
			continue
		}

		if err := vehicleRepo.Create(ctx, vehicle); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Failed to create vehicle %s %s: %v\n", vehicle.Make, vehicle.Model, err)
			continue
		}
		fmt.Printf("  ✓ Created vehicle: %d %s %s (VIN: %s)\n", vehicle.Year, vehicle.Make, vehicle.Model, vehicle.VIN)
		vehicleCount++
	}
	fmt.Printf("✓ Seeded %d vehicles\n", vehicleCount)

	// Seed maintenance records
	fmt.Println()
	fmt.Println("Seeding maintenance records...")
	maintenanceCount := 0
	sampleRecords := seeddata.GetSampleMaintenanceRecords()
	for _, record := range sampleRecords {
		// Check if record already exists
		if existingRecord, err := maintenanceRepo.FindByID(ctx, record.ID); err == nil && existingRecord != nil {
			fmt.Printf("  - Skipping maintenance record (already exists)\n")
			continue
		}

		if err := maintenanceRepo.Create(ctx, record); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Failed to create maintenance record: %v\n", err)
			continue
		}
		fmt.Printf("  ✓ Created maintenance record: %s on %s\n", record.ServiceType, record.ServiceDate.Format("2006-01-02"))
		maintenanceCount++
	}
	fmt.Printf("✓ Seeded %d maintenance records\n", maintenanceCount)

	// Summary
	fmt.Println()
	fmt.Println("================================")
	fmt.Println("Seeding completed successfully!")
	fmt.Printf("  - Users: %d\n", userCount)
	fmt.Printf("  - Vehicles: %d\n", vehicleCount)
	fmt.Printf("  - Maintenance records: %d\n", maintenanceCount)
	fmt.Println("================================")
}
