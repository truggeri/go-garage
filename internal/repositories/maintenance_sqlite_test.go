package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
)

func TestMaintenanceRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	maintenanceRepo := NewSQLiteMaintenanceRepository(db)
	ctx := context.Background()

	// Create user and vehicle
	user := &models.User{
		Username:     "maintenanceowner",
		Email:        "maintenance@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "GHGBH41JXMN109201",
		Make:   "Honda",
		Model:  "Civic",
		Year:   2020,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	t.Run("create valid maintenance record", func(t *testing.T) {
		record := &models.MaintenanceRecord{
			VehicleID:   vehicle.ID,
			ServiceType: "oil_change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}

		err := maintenanceRepo.Create(ctx, record)
		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
		assert.False(t, record.CreatedAt.IsZero())
		assert.False(t, record.UpdatedAt.IsZero())
	})

	t.Run("create with all fields", func(t *testing.T) {
		cost := 150.50
		mileage := 15000

		record := &models.MaintenanceRecord{
			VehicleID:        vehicle.ID,
			ServiceType:      "tire_rotation",
			ServiceDate:      time.Now().Add(-48 * time.Hour),
			Cost:             &cost,
			MileageAtService: &mileage,
			ServiceProvider:  "Quick Lube",
			Notes:            "All tires rotated",
		}

		err := maintenanceRepo.Create(ctx, record)
		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
	})

	t.Run("invalid vehicle ID", func(t *testing.T) {
		record := &models.MaintenanceRecord{
			VehicleID:   "non-existent-vehicle",
			ServiceType: "oil_change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}

		err := maintenanceRepo.Create(ctx, record)
		require.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
		assert.Contains(t, err.Error(), "vehicle does not exist")
	})

	t.Run("validation error", func(t *testing.T) {
		record := &models.MaintenanceRecord{
			VehicleID:   vehicle.ID,
			ServiceType: "", // Invalid: empty service type
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}

		err := maintenanceRepo.Create(ctx, record)
		require.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
	})
}

func TestMaintenanceRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	maintenanceRepo := NewSQLiteMaintenanceRepository(db)
	ctx := context.Background()

	// Setup
	user := &models.User{
		Username:     "findmaintowner",
		Email:        "findmaint@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "HHGBH41JXMN109202",
		Make:   "Toyota",
		Model:  "Camry",
		Year:   2019,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	t.Run("find existing record", func(t *testing.T) {
		cost := 200.0
		mileage := 25000

		record := &models.MaintenanceRecord{
			VehicleID:        vehicle.ID,
			ServiceType:      "brakes",
			ServiceDate:      time.Now().Add(-72 * time.Hour),
			Cost:             &cost,
			MileageAtService: &mileage,
			ServiceProvider:  "Brake Masters",
			Notes:            "Front brakes replaced",
		}

		err := maintenanceRepo.Create(ctx, record)
		require.NoError(t, err)

		found, err := maintenanceRepo.FindByID(ctx, record.ID)
		require.NoError(t, err)
		assert.Equal(t, record.ID, found.ID)
		assert.Equal(t, record.ServiceType, found.ServiceType)
		assert.Equal(t, record.ServiceProvider, found.ServiceProvider)
		assert.Equal(t, record.Notes, found.Notes)
		assert.NotNil(t, found.Cost)
		assert.Equal(t, 200.0, *found.Cost)
	})

	t.Run("record not found", func(t *testing.T) {
		_, err := maintenanceRepo.FindByID(ctx, "non-existent-id")
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestMaintenanceRepository_FindByVehicleID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	maintenanceRepo := NewSQLiteMaintenanceRepository(db)
	ctx := context.Background()

	// Setup
	user := &models.User{
		Username:     "vehiclemaintowner",
		Email:        "vehiclemaint@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle1 := &models.Vehicle{
		UserID: user.ID,
		VIN:    "1HGBH41JXMN109203",
		Make:   "Ford",
		Model:  "Focus",
		Year:   2018,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle1)
	require.NoError(t, err)

	vehicle2 := &models.Vehicle{
		UserID: user.ID,
		VIN:    "2HGBH41JXMN109204",
		Make:   "Ford",
		Model:  "Fusion",
		Year:   2019,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle2)
	require.NoError(t, err)

	t.Run("find records for vehicle with records", func(t *testing.T) {
		// Create maintenance records for vehicle1
		record1 := &models.MaintenanceRecord{
			VehicleID:   vehicle1.ID,
			ServiceType: "oil_change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}
		err := maintenanceRepo.Create(ctx, record1)
		require.NoError(t, err)

		record2 := &models.MaintenanceRecord{
			VehicleID:   vehicle1.ID,
			ServiceType: "tire_rotation",
			ServiceDate: time.Now().Add(-48 * time.Hour),
		}
		err = maintenanceRepo.Create(ctx, record2)
		require.NoError(t, err)

		// Create record for vehicle2
		record3 := &models.MaintenanceRecord{
			VehicleID:   vehicle2.ID,
			ServiceType: "glass",
			ServiceDate: time.Now().Add(-72 * time.Hour),
		}
		err = maintenanceRepo.Create(ctx, record3)
		require.NoError(t, err)

		// Find records for vehicle1
		records, err := maintenanceRepo.FindByVehicleID(ctx, vehicle1.ID)
		require.NoError(t, err)
		assert.Len(t, records, 2)
	})

	t.Run("find records for vehicle with no records", func(t *testing.T) {
		vehicle3 := &models.Vehicle{
			UserID: user.ID,
			VIN:    "3HGBH41JXMN109205",
			Make:   "Honda",
			Model:  "Accord",
			Year:   2020,
			Status: models.VehicleStatusActive,
		}
		err := vehicleRepo.Create(ctx, vehicle3)
		require.NoError(t, err)

		records, err := maintenanceRepo.FindByVehicleID(ctx, vehicle3.ID)
		require.NoError(t, err)
		assert.Empty(t, records)
	})
}

func TestMaintenanceRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	maintenanceRepo := NewSQLiteMaintenanceRepository(db)
	ctx := context.Background()

	// Setup
	user := &models.User{
		Username:     "updatemaintowner",
		Email:        "updatemaint@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "LHGBH41JXMN109206",
		Make:   "Toyota",
		Model:  "Corolla",
		Year:   2019,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	t.Run("update existing record", func(t *testing.T) {
		cost := 100.0
		record := &models.MaintenanceRecord{
			VehicleID:   vehicle.ID,
			ServiceType: "oil_change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
			Cost:        &cost,
		}

		err := maintenanceRepo.Create(ctx, record)
		require.NoError(t, err)

		oldUpdatedAt := record.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		// Update record
		newCost := 120.0
		record.Cost = &newCost
		record.ServiceProvider = "Updated Provider"
		record.Notes = "Updated notes"

		err = maintenanceRepo.Update(ctx, record)
		require.NoError(t, err)
		assert.True(t, record.UpdatedAt.After(oldUpdatedAt))

		// Verify update
		found, err := maintenanceRepo.FindByID(ctx, record.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Provider", found.ServiceProvider)
		assert.Equal(t, "Updated notes", found.Notes)
		assert.NotNil(t, found.Cost)
		assert.Equal(t, 120.0, *found.Cost)
	})

	t.Run("update non-existent record", func(t *testing.T) {
		record := &models.MaintenanceRecord{
			ID:          "non-existent",
			VehicleID:   vehicle.ID,
			ServiceType: "oil_change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}

		err := maintenanceRepo.Update(ctx, record)
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestMaintenanceRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	maintenanceRepo := NewSQLiteMaintenanceRepository(db)
	ctx := context.Background()

	// Setup
	user := &models.User{
		Username:     "deletemaintowner",
		Email:        "deletemaint@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "MHGBH41JXMN109207",
		Make:   "Nissan",
		Model:  "Altima",
		Year:   2020,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	t.Run("delete existing record", func(t *testing.T) {
		record := &models.MaintenanceRecord{
			VehicleID:   vehicle.ID,
			ServiceType: "oil_change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}

		err := maintenanceRepo.Create(ctx, record)
		require.NoError(t, err)

		err = maintenanceRepo.Delete(ctx, record.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = maintenanceRepo.FindByID(ctx, record.ID)
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})

	t.Run("delete non-existent record", func(t *testing.T) {
		err := maintenanceRepo.Delete(ctx, "non-existent-id")
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestMaintenanceRepository_Update_ValidationError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	maintenanceRepo := NewSQLiteMaintenanceRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "updatemaintvalowner",
		Email:        "updatemaintval@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "ZHGBH41JXMN109501",
		Make:   "Honda",
		Model:  "Civic",
		Year:   2020,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	t.Run("update with validation error", func(t *testing.T) {
		record := &models.MaintenanceRecord{
			VehicleID:   vehicle.ID,
			ServiceType: "oil_change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}

		err := maintenanceRepo.Create(ctx, record)
		require.NoError(t, err)

		record.ServiceType = ""
		err = maintenanceRepo.Update(ctx, record)
		require.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
	})
}

func TestMaintenanceRepository_Count(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	maintenanceRepo := NewSQLiteMaintenanceRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "countmaintowner",
		Email:        "countmaint@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "ZHGBH41JXMN109502",
		Make:   "Honda",
		Model:  "Civic",
		Year:   2020,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	records := []*models.MaintenanceRecord{
		{
			VehicleID:   vehicle.ID,
			ServiceType: "oil_change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		},
		{
			VehicleID:   vehicle.ID,
			ServiceType: "oil_change",
			ServiceDate: time.Now().Add(-48 * time.Hour),
		},
		{
			VehicleID:   vehicle.ID,
			ServiceType: "tire_rotation",
			ServiceDate: time.Now().Add(-72 * time.Hour),
		},
	}

	for _, r := range records {
		err := maintenanceRepo.Create(ctx, r)
		require.NoError(t, err)
	}

	t.Run("count all records", func(t *testing.T) {
		count, err := maintenanceRepo.Count(ctx, MaintenanceFilters{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3)
	})

	t.Run("count with vehicle ID filter", func(t *testing.T) {
		count, err := maintenanceRepo.Count(ctx, MaintenanceFilters{VehicleID: &vehicle.ID})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3)
	})

	t.Run("count with service type filter", func(t *testing.T) {
		serviceType := "oil_change"
		count, err := maintenanceRepo.Count(ctx, MaintenanceFilters{ServiceType: &serviceType})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 2)
	})
}

func TestMaintenanceRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	maintenanceRepo := NewSQLiteMaintenanceRepository(db)
	ctx := context.Background()

	// Setup
	user := &models.User{
		Username:     "listmaintowner",
		Email:        "listmaint@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "NHGBH41JXMN109208",
		Make:   "Mazda",
		Model:  "CX-5",
		Year:   2021,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	// Create test records
	records := []*models.MaintenanceRecord{
		{
			VehicleID:   vehicle.ID,
			ServiceType: "oil_change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		},
		{
			VehicleID:   vehicle.ID,
			ServiceType: "oil_change",
			ServiceDate: time.Now().Add(-48 * time.Hour),
		},
		{
			VehicleID:   vehicle.ID,
			ServiceType: "tire_rotation",
			ServiceDate: time.Now().Add(-72 * time.Hour),
		},
	}

	for _, r := range records {
		err := maintenanceRepo.Create(ctx, r)
		require.NoError(t, err)
	}

	t.Run("list all records", func(t *testing.T) {
		result, err := maintenanceRepo.List(ctx, MaintenanceFilters{}, PaginationParams{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3)
	})

	t.Run("filter by service type", func(t *testing.T) {
		serviceType := "oil_change"
		result, err := maintenanceRepo.List(ctx, MaintenanceFilters{ServiceType: &serviceType}, PaginationParams{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
		for _, r := range result {
			assert.Equal(t, "oil_change", r.ServiceType)
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		result, err := maintenanceRepo.List(ctx, MaintenanceFilters{}, PaginationParams{Limit: 2})
		require.NoError(t, err)
		assert.LessOrEqual(t, len(result), 2)
	})

	t.Run("filter by vehicle ID", func(t *testing.T) {
		result, err := maintenanceRepo.List(ctx, MaintenanceFilters{VehicleID: &vehicle.ID}, PaginationParams{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3)
		for _, r := range result {
			assert.Equal(t, vehicle.ID, r.VehicleID)
		}
	})

	t.Run("with offset pagination", func(t *testing.T) {
		all, err := maintenanceRepo.List(ctx, MaintenanceFilters{}, PaginationParams{})
		require.NoError(t, err)

		result, err := maintenanceRepo.List(ctx, MaintenanceFilters{}, PaginationParams{Limit: 10, Offset: len(all)})
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestMaintenanceRepository_CascadeDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	maintenanceRepo := NewSQLiteMaintenanceRepository(db)
	ctx := context.Background()

	// Setup
	user := &models.User{
		Username:     "cascadeowner",
		Email:        "cascade@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "4HGBH41JXMN109209",
		Make:   "Subaru",
		Model:  "Outback",
		Year:   2020,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	// Create maintenance records
	record := &models.MaintenanceRecord{
		VehicleID:   vehicle.ID,
		ServiceType: "oil_change",
		ServiceDate: time.Now().Add(-24 * time.Hour),
	}
	err = maintenanceRepo.Create(ctx, record)
	require.NoError(t, err)

	// Delete vehicle should cascade delete maintenance records
	err = vehicleRepo.Delete(ctx, vehicle.ID)
	require.NoError(t, err)

	// Verify maintenance record is also deleted
	_, err = maintenanceRepo.FindByID(ctx, record.ID)
	require.Error(t, err)
	assert.IsType(t, &models.NotFoundError{}, err)
}
