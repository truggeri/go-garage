package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
)

func TestFuelRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	fuelRepo := NewSQLiteFuelRepository(db)
	ctx := context.Background()

	// Create user and vehicle
	user := &models.User{
		Username:     "fuelowner",
		Email:        "fuel@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "1FGBH41JXMN200001",
		Make:   "Honda",
		Model:  "Civic",
		Year:   2020,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	t.Run("create valid fuel record", func(t *testing.T) {
		record := &models.FuelRecord{
			VehicleID:   vehicle.ID,
			FillDate:    time.Now().Add(-24 * time.Hour),
			Odometer:    50000,
			CostPerUnit: 3.50,
			Volume:      12.5,
		}

		err := fuelRepo.Create(ctx, record)
		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
		assert.False(t, record.CreatedAt.IsZero())
		assert.False(t, record.UpdatedAt.IsZero())
	})

	t.Run("create with all optional fields", func(t *testing.T) {
		pct := 60
		mpg := 32.5

		record := &models.FuelRecord{
			VehicleID:      vehicle.ID,
			FillDate:       time.Now().Add(-48 * time.Hour),
			Odometer:       51000,
			CostPerUnit:    3.75,
			Volume:         11.8,
			FuelType:       "Regular",
			CityDrivingPct: &pct,
			Location:       "Shell on Main St",
			Brand:          "Shell",
			Notes:          "Filled up on road trip",
			ReportedMPG:    &mpg,
			PartialFuelUp:  false,
		}

		err := fuelRepo.Create(ctx, record)
		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
	})

	t.Run("create partial fuel up", func(t *testing.T) {
		record := &models.FuelRecord{
			VehicleID:     vehicle.ID,
			FillDate:      time.Now().Add(-72 * time.Hour),
			Odometer:      49000,
			CostPerUnit:   3.60,
			Volume:        5.0,
			PartialFuelUp: true,
		}

		err := fuelRepo.Create(ctx, record)
		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
	})

	t.Run("invalid vehicle ID", func(t *testing.T) {
		record := &models.FuelRecord{
			VehicleID:   "non-existent-vehicle",
			FillDate:    time.Now().Add(-24 * time.Hour),
			Odometer:    50000,
			CostPerUnit: 3.50,
			Volume:      12.5,
		}

		err := fuelRepo.Create(ctx, record)
		require.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
		assert.Contains(t, err.Error(), "vehicle does not exist")
	})

	t.Run("validation error - zero volume", func(t *testing.T) {
		record := &models.FuelRecord{
			VehicleID:   vehicle.ID,
			FillDate:    time.Now().Add(-24 * time.Hour),
			Odometer:    50000,
			CostPerUnit: 3.50,
			Volume:      0, // Invalid
		}

		err := fuelRepo.Create(ctx, record)
		require.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
	})
}

func TestFuelRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	fuelRepo := NewSQLiteFuelRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "findfuelowner",
		Email:        "findfuel@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "1FGBH41JXMN200002",
		Make:   "Toyota",
		Model:  "Camry",
		Year:   2019,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	t.Run("find existing record", func(t *testing.T) {
		pct := 40
		mpg := 30.0

		record := &models.FuelRecord{
			VehicleID:      vehicle.ID,
			FillDate:       time.Now().Add(-24 * time.Hour),
			Odometer:       25000,
			CostPerUnit:    3.99,
			Volume:         14.2,
			FuelType:       "Premium",
			CityDrivingPct: &pct,
			Location:       "BP on Oak Ave",
			Brand:          "BP",
			Notes:          "First fill-up",
			ReportedMPG:    &mpg,
			PartialFuelUp:  false,
		}

		err := fuelRepo.Create(ctx, record)
		require.NoError(t, err)

		found, err := fuelRepo.FindByID(ctx, record.ID)
		require.NoError(t, err)
		assert.Equal(t, record.ID, found.ID)
		assert.Equal(t, record.VehicleID, found.VehicleID)
		assert.Equal(t, record.Odometer, found.Odometer)
		assert.Equal(t, record.CostPerUnit, found.CostPerUnit)
		assert.Equal(t, record.Volume, found.Volume)
		assert.Equal(t, "Premium", found.FuelType)
		assert.Equal(t, "BP on Oak Ave", found.Location)
		assert.Equal(t, "BP", found.Brand)
		assert.Equal(t, "First fill-up", found.Notes)
		assert.NotNil(t, found.CityDrivingPct)
		assert.Equal(t, 40, *found.CityDrivingPct)
		assert.NotNil(t, found.ReportedMPG)
		assert.Equal(t, 30.0, *found.ReportedMPG)
		assert.False(t, found.PartialFuelUp)
	})

	t.Run("record not found", func(t *testing.T) {
		_, err := fuelRepo.FindByID(ctx, "non-existent-id")
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestFuelRepository_FindByVehicleID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	fuelRepo := NewSQLiteFuelRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "vehiclefuelowner",
		Email:        "vehiclefuel@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle1 := &models.Vehicle{
		UserID: user.ID,
		VIN:    "1FGBH41JXMN200003",
		Make:   "Ford",
		Model:  "Focus",
		Year:   2018,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle1)
	require.NoError(t, err)

	vehicle2 := &models.Vehicle{
		UserID: user.ID,
		VIN:    "1FGBH41JXMN200004",
		Make:   "Ford",
		Model:  "Fusion",
		Year:   2019,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle2)
	require.NoError(t, err)

	t.Run("find records for vehicle with records", func(t *testing.T) {
		record1 := &models.FuelRecord{
			VehicleID:   vehicle1.ID,
			FillDate:    time.Now().Add(-24 * time.Hour),
			Odometer:    30000,
			CostPerUnit: 3.45,
			Volume:      10.0,
		}
		err := fuelRepo.Create(ctx, record1)
		require.NoError(t, err)

		record2 := &models.FuelRecord{
			VehicleID:   vehicle1.ID,
			FillDate:    time.Now().Add(-48 * time.Hour),
			Odometer:    29700,
			CostPerUnit: 3.40,
			Volume:      11.5,
		}
		err = fuelRepo.Create(ctx, record2)
		require.NoError(t, err)

		// Create record for vehicle2
		record3 := &models.FuelRecord{
			VehicleID:   vehicle2.ID,
			FillDate:    time.Now().Add(-72 * time.Hour),
			Odometer:    20000,
			CostPerUnit: 3.55,
			Volume:      9.0,
		}
		err = fuelRepo.Create(ctx, record3)
		require.NoError(t, err)

		records, err := fuelRepo.FindByVehicleID(ctx, vehicle1.ID)
		require.NoError(t, err)
		assert.Len(t, records, 2)
	})

	t.Run("find records for vehicle with no records", func(t *testing.T) {
		vehicle3 := &models.Vehicle{
			UserID: user.ID,
			VIN:    "1FGBH41JXMN200005",
			Make:   "Honda",
			Model:  "Accord",
			Year:   2020,
			Status: models.VehicleStatusActive,
		}
		err := vehicleRepo.Create(ctx, vehicle3)
		require.NoError(t, err)

		records, err := fuelRepo.FindByVehicleID(ctx, vehicle3.ID)
		require.NoError(t, err)
		assert.Empty(t, records)
	})
}

func TestFuelRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	fuelRepo := NewSQLiteFuelRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "updatefuelowner",
		Email:        "updatefuel@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "1FGBH41JXMN200006",
		Make:   "Toyota",
		Model:  "Corolla",
		Year:   2019,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	t.Run("update existing record", func(t *testing.T) {
		record := &models.FuelRecord{
			VehicleID:   vehicle.ID,
			FillDate:    time.Now().Add(-24 * time.Hour),
			Odometer:    40000,
			CostPerUnit: 3.50,
			Volume:      12.0,
		}

		err := fuelRepo.Create(ctx, record)
		require.NoError(t, err)

		oldUpdatedAt := record.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		record.CostPerUnit = 3.75
		record.Notes = "Updated notes"
		record.Location = "New Location"

		err = fuelRepo.Update(ctx, record)
		require.NoError(t, err)
		assert.True(t, record.UpdatedAt.After(oldUpdatedAt))

		found, err := fuelRepo.FindByID(ctx, record.ID)
		require.NoError(t, err)
		assert.Equal(t, 3.75, found.CostPerUnit)
		assert.Equal(t, "Updated notes", found.Notes)
		assert.Equal(t, "New Location", found.Location)
	})

	t.Run("update non-existent record", func(t *testing.T) {
		record := &models.FuelRecord{
			ID:          "non-existent",
			VehicleID:   vehicle.ID,
			FillDate:    time.Now().Add(-24 * time.Hour),
			Odometer:    40000,
			CostPerUnit: 3.50,
			Volume:      12.0,
		}

		err := fuelRepo.Update(ctx, record)
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})

	t.Run("update with validation error", func(t *testing.T) {
		record := &models.FuelRecord{
			VehicleID:   vehicle.ID,
			FillDate:    time.Now().Add(-24 * time.Hour),
			Odometer:    40000,
			CostPerUnit: 3.50,
			Volume:      10.0,
		}

		err := fuelRepo.Create(ctx, record)
		require.NoError(t, err)

		record.Volume = 0 // Invalid
		err = fuelRepo.Update(ctx, record)
		require.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
	})
}

func TestFuelRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	fuelRepo := NewSQLiteFuelRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "deletefuelowner",
		Email:        "deletefuel@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "1FGBH41JXMN200007",
		Make:   "Nissan",
		Model:  "Altima",
		Year:   2020,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	t.Run("delete existing record", func(t *testing.T) {
		record := &models.FuelRecord{
			VehicleID:   vehicle.ID,
			FillDate:    time.Now().Add(-24 * time.Hour),
			Odometer:    35000,
			CostPerUnit: 3.60,
			Volume:      13.0,
		}

		err := fuelRepo.Create(ctx, record)
		require.NoError(t, err)

		err = fuelRepo.Delete(ctx, record.ID)
		require.NoError(t, err)

		_, err = fuelRepo.FindByID(ctx, record.ID)
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})

	t.Run("delete non-existent record", func(t *testing.T) {
		err := fuelRepo.Delete(ctx, "non-existent-id")
		require.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestFuelRepository_Count(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	fuelRepo := NewSQLiteFuelRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "countfuelowner",
		Email:        "countfuel@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "1FGBH41JXMN200008",
		Make:   "Honda",
		Model:  "Civic",
		Year:   2020,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	records := []*models.FuelRecord{
		{
			VehicleID:   vehicle.ID,
			FillDate:    time.Now().Add(-24 * time.Hour),
			Odometer:    60000,
			CostPerUnit: 3.50,
			Volume:      12.0,
			FuelType:    "Regular",
		},
		{
			VehicleID:   vehicle.ID,
			FillDate:    time.Now().Add(-48 * time.Hour),
			Odometer:    59700,
			CostPerUnit: 3.45,
			Volume:      11.5,
			FuelType:    "Regular",
		},
		{
			VehicleID:   vehicle.ID,
			FillDate:    time.Now().Add(-72 * time.Hour),
			Odometer:    59400,
			CostPerUnit: 4.20,
			Volume:      10.8,
			FuelType:    "Premium",
		},
	}

	for _, r := range records {
		err := fuelRepo.Create(ctx, r)
		require.NoError(t, err)
	}

	t.Run("count all records", func(t *testing.T) {
		count, err := fuelRepo.Count(ctx, FuelFilters{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3)
	})

	t.Run("count with vehicle ID filter", func(t *testing.T) {
		count, err := fuelRepo.Count(ctx, FuelFilters{VehicleID: &vehicle.ID})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3)
	})

	t.Run("count with fuel type filter", func(t *testing.T) {
		fuelType := "Regular"
		count, err := fuelRepo.Count(ctx, FuelFilters{FuelType: &fuelType})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 2)
	})
}

func TestFuelRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	fuelRepo := NewSQLiteFuelRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "listfuelowner",
		Email:        "listfuel@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "1FGBH41JXMN200009",
		Make:   "Mazda",
		Model:  "CX-5",
		Year:   2021,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	// Create test records
	testRecords := []*models.FuelRecord{
		{
			VehicleID:   vehicle.ID,
			FillDate:    time.Now().Add(-24 * time.Hour),
			Odometer:    70000,
			CostPerUnit: 3.50,
			Volume:      12.0,
			FuelType:    "Regular",
		},
		{
			VehicleID:   vehicle.ID,
			FillDate:    time.Now().Add(-48 * time.Hour),
			Odometer:    69700,
			CostPerUnit: 3.45,
			Volume:      11.5,
			FuelType:    "Regular",
		},
		{
			VehicleID:   vehicle.ID,
			FillDate:    time.Now().Add(-72 * time.Hour),
			Odometer:    69400,
			CostPerUnit: 4.20,
			Volume:      10.8,
			FuelType:    "Premium",
		},
	}

	for _, r := range testRecords {
		err := fuelRepo.Create(ctx, r)
		require.NoError(t, err)
	}

	t.Run("list all records", func(t *testing.T) {
		result, err := fuelRepo.List(ctx, FuelFilters{}, PaginationParams{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3)
	})

	t.Run("filter by fuel type", func(t *testing.T) {
		fuelType := "Regular"
		result, err := fuelRepo.List(ctx, FuelFilters{FuelType: &fuelType}, PaginationParams{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
		for _, r := range result {
			assert.Equal(t, "Regular", r.FuelType)
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		result, err := fuelRepo.List(ctx, FuelFilters{}, PaginationParams{Limit: 2})
		require.NoError(t, err)
		assert.LessOrEqual(t, len(result), 2)
	})

	t.Run("filter by vehicle ID", func(t *testing.T) {
		result, err := fuelRepo.List(ctx, FuelFilters{VehicleID: &vehicle.ID}, PaginationParams{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3)
		for _, r := range result {
			assert.Equal(t, vehicle.ID, r.VehicleID)
		}
	})

	t.Run("with offset pagination", func(t *testing.T) {
		all, err := fuelRepo.List(ctx, FuelFilters{}, PaginationParams{})
		require.NoError(t, err)

		result, err := fuelRepo.List(ctx, FuelFilters{}, PaginationParams{Limit: 10, Offset: len(all)})
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestFuelRepository_CascadeDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	fuelRepo := NewSQLiteFuelRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "cascadefuelowner",
		Email:        "cascadefuel@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "1FGBH41JXMN200010",
		Make:   "Subaru",
		Model:  "Outback",
		Year:   2020,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	record := &models.FuelRecord{
		VehicleID:   vehicle.ID,
		FillDate:    time.Now().Add(-24 * time.Hour),
		Odometer:    45000,
		CostPerUnit: 3.55,
		Volume:      11.0,
	}
	err = fuelRepo.Create(ctx, record)
	require.NoError(t, err)

	// Delete vehicle should cascade delete fuel records
	err = vehicleRepo.Delete(ctx, vehicle.ID)
	require.NoError(t, err)

	// Verify fuel record is also deleted
	_, err = fuelRepo.FindByID(ctx, record.ID)
	require.Error(t, err)
	assert.IsType(t, &models.NotFoundError{}, err)
}
