package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
)

func setupMetricsTestData(t *testing.T) (context.Context, *SQLiteMetricsRepository, *models.Vehicle, func()) {
	t.Helper()
	db, cleanup := setupTestDB(t)
	ctx := context.Background()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	metricsRepo := NewSQLiteMetricsRepository(db)

	user := &models.User{
		Username:     "metricsowner",
		Email:        "metrics@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle := &models.Vehicle{
		UserID: user.ID,
		VIN:    "METRCS12345678901",
		Make:   "Toyota",
		Model:  "Camry",
		Year:   2022,
		Status: models.VehicleStatusActive,
	}
	err = vehicleRepo.Create(ctx, vehicle)
	require.NoError(t, err)

	return ctx, metricsRepo, vehicle, cleanup
}

func TestMetricsRepository_Upsert(t *testing.T) {
	ctx, metricsRepo, vehicle, cleanup := setupMetricsTestData(t)
	defer cleanup()

	t.Run("inserts new metrics record", func(t *testing.T) {
		totalSpent := 250.00
		metrics := &models.VehicleMetrics{
			VehicleID:  vehicle.ID,
			TotalSpent: &totalSpent,
		}

		err := metricsRepo.Upsert(ctx, metrics)
		require.NoError(t, err)

		result, err := metricsRepo.GetByVehicleID(ctx, vehicle.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, vehicle.ID, result.VehicleID)
		require.NotNil(t, result.TotalSpent)
		assert.Equal(t, 250.00, *result.TotalSpent)
	})

	t.Run("updates existing metrics record", func(t *testing.T) {
		newTotal := 500.00
		metrics := &models.VehicleMetrics{
			VehicleID:  vehicle.ID,
			TotalSpent: &newTotal,
		}

		err := metricsRepo.Upsert(ctx, metrics)
		require.NoError(t, err)

		result, err := metricsRepo.GetByVehicleID(ctx, vehicle.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.TotalSpent)
		assert.Equal(t, 500.00, *result.TotalSpent)
	})

	t.Run("upserts with nil total_spent", func(t *testing.T) {
		metrics := &models.VehicleMetrics{
			VehicleID:  vehicle.ID,
			TotalSpent: nil,
		}

		err := metricsRepo.Upsert(ctx, metrics)
		require.NoError(t, err)

		result, err := metricsRepo.GetByVehicleID(ctx, vehicle.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Nil(t, result.TotalSpent)
	})
}

func TestMetricsRepository_GetByVehicleID(t *testing.T) {
	ctx, metricsRepo, _, cleanup := setupMetricsTestData(t)
	defer cleanup()

	t.Run("returns nil for non-existent vehicle", func(t *testing.T) {
		result, err := metricsRepo.GetByVehicleID(ctx, "non-existent-id")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestMetricsRepository_SumTotalSpentByVehicleIDs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	metricsRepo := NewSQLiteMetricsRepository(db)

	user := &models.User{
		Username:     "sumowner",
		Email:        "sum@example.com",
		PasswordHash: "hashed_password",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	vehicle1 := &models.Vehicle{
		UserID: user.ID, VIN: "SUM00000000000001",
		Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
	}
	vehicle2 := &models.Vehicle{
		UserID: user.ID, VIN: "SUM00000000000002",
		Make: "Honda", Model: "Civic", Year: 2021, Status: models.VehicleStatusActive,
	}
	require.NoError(t, vehicleRepo.Create(ctx, vehicle1))
	require.NoError(t, vehicleRepo.Create(ctx, vehicle2))

	t.Run("returns 0 for empty vehicle list", func(t *testing.T) {
		sum, err := metricsRepo.SumTotalSpentByVehicleIDs(ctx, []string{})
		require.NoError(t, err)
		assert.Equal(t, 0.0, sum)
	})

	t.Run("returns 0 when no metrics exist", func(t *testing.T) {
		sum, err := metricsRepo.SumTotalSpentByVehicleIDs(ctx, []string{vehicle1.ID, vehicle2.ID})
		require.NoError(t, err)
		assert.Equal(t, 0.0, sum)
	})

	t.Run("sums total spent across vehicles", func(t *testing.T) {
		spent1 := 100.50
		spent2 := 200.75
		require.NoError(t, metricsRepo.Upsert(ctx, &models.VehicleMetrics{VehicleID: vehicle1.ID, TotalSpent: &spent1}))
		require.NoError(t, metricsRepo.Upsert(ctx, &models.VehicleMetrics{VehicleID: vehicle2.ID, TotalSpent: &spent2}))

		sum, err := metricsRepo.SumTotalSpentByVehicleIDs(ctx, []string{vehicle1.ID, vehicle2.ID})
		require.NoError(t, err)
		assert.Equal(t, 301.25, sum)
	})

	t.Run("handles nil total_spent in sum", func(t *testing.T) {
		// Update vehicle2 to nil
		require.NoError(t, metricsRepo.Upsert(ctx, &models.VehicleMetrics{VehicleID: vehicle2.ID, TotalSpent: nil}))

		sum, err := metricsRepo.SumTotalSpentByVehicleIDs(ctx, []string{vehicle1.ID, vehicle2.ID})
		require.NoError(t, err)
		assert.Equal(t, 100.50, sum)
	})
}

func TestMaintenanceRepository_SumCostByVehicleID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	userRepo := NewSQLiteUserRepository(db)
	vehicleRepo := NewSQLiteVehicleRepository(db)
	maintenanceRepo := NewSQLiteMaintenanceRepository(db)

	user := &models.User{
		Username:     "sumcostowner",
		Email:        "sumcost@example.com",
		PasswordHash: "hashed_password",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	vehicle := &models.Vehicle{
		UserID: user.ID, VIN: "SUMCST12345678901",
		Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
	}
	require.NoError(t, vehicleRepo.Create(ctx, vehicle))

	t.Run("returns nil when no records exist", func(t *testing.T) {
		result, err := maintenanceRepo.SumCostByVehicleID(ctx, vehicle.ID)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns nil when records have no cost", func(t *testing.T) {
		record := &models.MaintenanceRecord{
			VehicleID:   vehicle.ID,
			ServiceType: "Inspection",
			ServiceDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		}
		require.NoError(t, maintenanceRepo.Create(ctx, record))

		result, err := maintenanceRepo.SumCostByVehicleID(ctx, vehicle.ID)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("sums costs correctly", func(t *testing.T) {
		cost1 := 89.99
		cost2 := 150.00
		record1 := &models.MaintenanceRecord{
			VehicleID:   vehicle.ID,
			ServiceType: "Oil Change",
			ServiceDate: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			Cost:        &cost1,
		}
		record2 := &models.MaintenanceRecord{
			VehicleID:   vehicle.ID,
			ServiceType: "Brake Pad Replace",
			ServiceDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			Cost:        &cost2,
		}
		require.NoError(t, maintenanceRepo.Create(ctx, record1))
		require.NoError(t, maintenanceRepo.Create(ctx, record2))

		result, err := maintenanceRepo.SumCostByVehicleID(ctx, vehicle.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.InDelta(t, 239.99, *result, 0.01)
	})
}
