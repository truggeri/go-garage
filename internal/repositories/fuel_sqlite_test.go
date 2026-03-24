package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
)

// setupFuelTestData creates prerequisite user and vehicle for fuel tests
func setupFuelTestData(t *testing.T) (*SQLiteFuelRepository, *models.Vehicle, context.Context, func()) {
	t.Helper()

	db, cleanup := setupTestDB(t)
	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	fuelRepo := NewSQLiteFuelRepository(db)
	ctx := context.Background()

	user := &models.User{
		Username:     "fuelowner",
		Email:        "fuel@example.com",
		PasswordHash: "hashed_password",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "FUELH41JXMN109201",
		Make:   "Toyota",
		Model:  "Corolla",
		Year:   2022,
		Status: models.VehicleStatusActive,
	}
	require.NoError(t, vehicleRepo.Create(ctx, vehicle))

	return fuelRepo, vehicle, ctx, cleanup
}

func validTestFuelRecord(vehicleID string) *models.FuelRecord {
	return &models.FuelRecord{
		VehicleID: vehicleID,
		FillDate:  time.Now().Add(-24 * time.Hour),
		Mileage:   50000,
		Volume:    12.5,
		FuelType:  "gasoline",
	}
}

func TestFuelRepository_Create(t *testing.T) {
	fuelRepo, vehicle, ctx, cleanup := setupFuelTestData(t)
	defer cleanup()

	t.Run("create valid fuel record", func(t *testing.T) {
		record := validTestFuelRecord(vehicle.ID)
		err := fuelRepo.Create(ctx, record)
		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
		assert.False(t, record.CreatedAt.IsZero())
		assert.False(t, record.UpdatedAt.IsZero())
	})

	t.Run("create with all fields", func(t *testing.T) {
		price := 3.459
		octane := 93
		cityPct := 60
		mpg := 32.5

		record := &models.FuelRecord{
			VehicleID:             vehicle.ID,
			FillDate:              time.Now().Add(-48 * time.Hour),
			Mileage:               50300,
			Volume:                14.2,
			FuelType:              "gasoline",
			PartialFill:           true,
			PricePerUnit:          &price,
			OctaneRating:          &octane,
			Location:              "Shell Station",
			Brand:                 "Shell",
			Notes:                 "Premium fuel",
			CityDrivingPercentage: &cityPct,
			VehicleReportedMPG:    &mpg,
		}

		err := fuelRepo.Create(ctx, record)
		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
	})

	t.Run("invalid vehicle ID", func(t *testing.T) {
		record := validTestFuelRecord("non-existent-vehicle")
		err := fuelRepo.Create(ctx, record)
		require.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
		assert.Contains(t, err.Error(), "vehicle does not exist")
	})

	t.Run("validation error empty fuel type", func(t *testing.T) {
		record := validTestFuelRecord(vehicle.ID)
		record.FuelType = ""
		err := fuelRepo.Create(ctx, record)
		require.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
	})
}

func TestFuelRepository_FindByID(t *testing.T) {
	fuelRepo, vehicle, ctx, cleanup := setupFuelTestData(t)
	defer cleanup()

	t.Run("find existing record", func(t *testing.T) {
		price := 3.50
		octane := 87
		record := &models.FuelRecord{
			VehicleID:    vehicle.ID,
			FillDate:     time.Now().Add(-24 * time.Hour),
			Mileage:      50000,
			Volume:       12.5,
			FuelType:     string(models.FuelTypeDiesel),
			PartialFill:  false,
			PricePerUnit: &price,
			OctaneRating: &octane,
			Location:     "Costco",
			Brand:        "Kirkland",
			Notes:        "Cheap fuel",
		}
		require.NoError(t, fuelRepo.Create(ctx, record))

		found, err := fuelRepo.FindByID(ctx, record.ID)
		require.NoError(t, err)
		assert.Equal(t, record.ID, found.ID)
		assert.Equal(t, vehicle.ID, found.VehicleID)
		assert.Equal(t, 50000, found.Mileage)
		assert.Equal(t, 12.5, found.Volume)
		assert.Equal(t, string(models.FuelTypeDiesel), found.FuelType)
		assert.False(t, found.PartialFill)
		assert.InDelta(t, 3.50, *found.PricePerUnit, 0.001)
		assert.Equal(t, 87, *found.OctaneRating)
		assert.Equal(t, "Costco", found.Location)
		assert.Equal(t, "Kirkland", found.Brand)
		assert.Equal(t, "Cheap fuel", found.Notes)
	})

	t.Run("find with nullable fields nil", func(t *testing.T) {
		record := validTestFuelRecord(vehicle.ID)
		require.NoError(t, fuelRepo.Create(ctx, record))

		found, err := fuelRepo.FindByID(ctx, record.ID)
		require.NoError(t, err)
		assert.Nil(t, found.PricePerUnit)
		assert.Nil(t, found.OctaneRating)
		assert.Empty(t, found.Location)
		assert.Empty(t, found.Brand)
		assert.Empty(t, found.Notes)
		assert.Nil(t, found.CityDrivingPercentage)
		assert.Nil(t, found.VehicleReportedMPG)
	})

	t.Run("not found", func(t *testing.T) {
		found, err := fuelRepo.FindByID(ctx, "non-existent-id")
		assert.Nil(t, found)
		assert.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestFuelRepository_FindByVehicleID(t *testing.T) {
	fuelRepo, vehicle, ctx, cleanup := setupFuelTestData(t)
	defer cleanup()

	// Create multiple records
	for i := 0; i < 3; i++ {
		record := validTestFuelRecord(vehicle.ID)
		record.Mileage = 50000 + (i * 300)
		record.FillDate = time.Now().Add(-time.Duration(i+1) * 24 * time.Hour)
		require.NoError(t, fuelRepo.Create(ctx, record))
	}

	t.Run("returns records ordered by fill_date desc", func(t *testing.T) {
		records, err := fuelRepo.FindByVehicleID(ctx, vehicle.ID)
		require.NoError(t, err)
		assert.Len(t, records, 3)
		// Most recent first
		assert.True(t, records[0].FillDate.After(records[1].FillDate))
		assert.True(t, records[1].FillDate.After(records[2].FillDate))
	})

	t.Run("returns empty for unknown vehicle", func(t *testing.T) {
		records, err := fuelRepo.FindByVehicleID(ctx, "unknown-vehicle")
		require.NoError(t, err)
		assert.Empty(t, records)
	})
}

func TestFuelRepository_Update(t *testing.T) {
	fuelRepo, vehicle, ctx, cleanup := setupFuelTestData(t)
	defer cleanup()

	t.Run("update existing record", func(t *testing.T) {
		record := validTestFuelRecord(vehicle.ID)
		require.NoError(t, fuelRepo.Create(ctx, record))

		record.Mileage = 55000
		record.Volume = 15.0
		record.FuelType = "e85"
		record.PartialFill = true

		err := fuelRepo.Update(ctx, record)
		require.NoError(t, err)

		found, err := fuelRepo.FindByID(ctx, record.ID)
		require.NoError(t, err)
		assert.Equal(t, 55000, found.Mileage)
		assert.Equal(t, 15.0, found.Volume)
		assert.Equal(t, "e85", found.FuelType)
		assert.True(t, found.PartialFill)
	})

	t.Run("update non-existent record", func(t *testing.T) {
		record := validTestFuelRecord(vehicle.ID)
		record.ID = "non-existent-id"
		err := fuelRepo.Update(ctx, record)
		assert.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})

	t.Run("update with validation error", func(t *testing.T) {
		record := validTestFuelRecord(vehicle.ID)
		require.NoError(t, fuelRepo.Create(ctx, record))

		record.Mileage = 0
		err := fuelRepo.Update(ctx, record)
		assert.Error(t, err)
		assert.IsType(t, &models.ValidationError{}, err)
	})
}

func TestFuelRepository_Delete(t *testing.T) {
	fuelRepo, vehicle, ctx, cleanup := setupFuelTestData(t)
	defer cleanup()

	t.Run("delete existing record", func(t *testing.T) {
		record := validTestFuelRecord(vehicle.ID)
		require.NoError(t, fuelRepo.Create(ctx, record))

		err := fuelRepo.Delete(ctx, record.ID)
		require.NoError(t, err)

		found, err := fuelRepo.FindByID(ctx, record.ID)
		assert.Nil(t, found)
		assert.Error(t, err)
	})

	t.Run("delete non-existent record", func(t *testing.T) {
		err := fuelRepo.Delete(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.IsType(t, &models.NotFoundError{}, err)
	})
}

func TestFuelRepository_List(t *testing.T) {
	fuelRepo, vehicle, ctx, cleanup := setupFuelTestData(t)
	defer cleanup()

	// Create records with different fuel types
	for i, ft := range []string{string(models.FuelTypeGasoline), string(models.FuelTypeDiesel), string(models.FuelTypeGasoline)} {
		record := validTestFuelRecord(vehicle.ID)
		record.Mileage = 50000 + (i * 300)
		record.FuelType = ft
		record.FillDate = time.Now().Add(-time.Duration(i+1) * 24 * time.Hour)
		require.NoError(t, fuelRepo.Create(ctx, record))
	}

	t.Run("list all", func(t *testing.T) {
		records, err := fuelRepo.List(ctx, FuelFilters{}, PaginationParams{})
		require.NoError(t, err)
		assert.Len(t, records, 3)
	})

	t.Run("filter by vehicle_id", func(t *testing.T) {
		vid := vehicle.ID
		records, err := fuelRepo.List(ctx, FuelFilters{VehicleID: &vid}, PaginationParams{})
		require.NoError(t, err)
		assert.Len(t, records, 3)
	})

	t.Run("filter by fuel_type", func(t *testing.T) {
		ft := string(models.FuelTypeDiesel)
		records, err := fuelRepo.List(ctx, FuelFilters{FuelType: &ft}, PaginationParams{})
		require.NoError(t, err)
		assert.Len(t, records, 1)
		assert.Equal(t, string(models.FuelTypeDiesel), records[0].FuelType)
	})

	t.Run("pagination limit", func(t *testing.T) {
		records, err := fuelRepo.List(ctx, FuelFilters{}, PaginationParams{Limit: 2})
		require.NoError(t, err)
		assert.Len(t, records, 2)
	})

	t.Run("pagination offset", func(t *testing.T) {
		records, err := fuelRepo.List(ctx, FuelFilters{}, PaginationParams{Limit: 10, Offset: 2})
		require.NoError(t, err)
		assert.Len(t, records, 1)
	})
}

func TestFuelRepository_Count(t *testing.T) {
	fuelRepo, vehicle, ctx, cleanup := setupFuelTestData(t)
	defer cleanup()

	for i := 0; i < 3; i++ {
		record := validTestFuelRecord(vehicle.ID)
		record.Mileage = 50000 + (i * 300)
		record.FillDate = time.Now().Add(-time.Duration(i+1) * 24 * time.Hour)
		if i == 0 {
			record.FuelType = string(models.FuelTypeDiesel)
		}
		require.NoError(t, fuelRepo.Create(ctx, record))
	}

	t.Run("count all", func(t *testing.T) {
		count, err := fuelRepo.Count(ctx, FuelFilters{})
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("count with fuel_type filter", func(t *testing.T) {
		ft := string(models.FuelTypeDiesel)
		count, err := fuelRepo.Count(ctx, FuelFilters{FuelType: &ft})
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("count with no matches", func(t *testing.T) {
		ft := "e85"
		count, err := fuelRepo.Count(ctx, FuelFilters{FuelType: &ft})
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
