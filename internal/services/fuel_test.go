package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
)

// mockFuelRepository is a mock implementation of FuelRepository for testing
type mockFuelRepository struct {
	records    map[string]*models.FuelRecord
	createErr  error
	findErr    error
	updateErr  error
	deleteErr  error
	listResult []*models.FuelRecord
}

func newMockFuelRepository() *mockFuelRepository {
	return &mockFuelRepository{
		records: make(map[string]*models.FuelRecord),
	}
}

func (m *mockFuelRepository) Create(ctx context.Context, record *models.FuelRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	if record.ID == "" {
		record.ID = "test-fuel-id"
	}
	m.records[record.ID] = record
	return nil
}

func (m *mockFuelRepository) FindByID(ctx context.Context, id string) (*models.FuelRecord, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	record, exists := m.records[id]
	if !exists {
		return nil, models.NewNotFoundError("FuelRecord", id)
	}
	return record, nil
}

func (m *mockFuelRepository) FindByVehicleID(ctx context.Context, vehicleID string) ([]*models.FuelRecord, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []*models.FuelRecord
	for _, r := range m.records {
		if r.VehicleID == vehicleID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockFuelRepository) Update(ctx context.Context, record *models.FuelRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, exists := m.records[record.ID]; !exists {
		return models.NewNotFoundError("FuelRecord", record.ID)
	}
	m.records[record.ID] = record
	return nil
}

func (m *mockFuelRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, exists := m.records[id]; !exists {
		return models.NewNotFoundError("FuelRecord", id)
	}
	delete(m.records, id)
	return nil
}

func (m *mockFuelRepository) List(ctx context.Context, filters repositories.FuelFilters, pagination repositories.PaginationParams) ([]*models.FuelRecord, error) {
	if m.listResult != nil {
		return m.listResult, nil
	}
	var result []*models.FuelRecord
	for _, r := range m.records {
		result = append(result, r)
	}
	return result, nil
}

func (m *mockFuelRepository) Count(ctx context.Context, filters repositories.FuelFilters) (int, error) {
	if m.listResult != nil {
		return len(m.listResult), nil
	}
	return len(m.records), nil
}

func (m *mockFuelRepository) SumCostByVehicleID(ctx context.Context, vehicleID string) (*float64, error) {
	var total float64
	var hasCost bool
	for _, r := range m.records {
		if r.VehicleID == vehicleID && r.PricePerUnit != nil {
			total += *r.PricePerUnit * r.Volume
			hasCost = true
		}
	}
	if !hasCost {
		return nil, nil
	}
	return &total, nil
}

func TestFuelService_CreateFuel(t *testing.T) {
	ctx := context.Background()

	t.Run("creates fuel record successfully", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		vehicleRepo.vehicles["vehicle-123"] = &models.Vehicle{
			ID: "vehicle-123", UserID: "user-123", VIN: "1HGBH41JXMN109186",
			Make: "Honda", Model: "Civic", Year: 2021, Status: models.VehicleStatusActive,
		}
		service := NewFuelService(fuelRepo, vehicleRepo, nil)

		price := 3.50
		record := &models.FuelRecord{
			VehicleID:    "vehicle-123",
			FillDate:     time.Now().Add(-24 * time.Hour),
			Mileage:      50000,
			Volume:       12.5,
			FuelType:     string(models.FuelTypeGasoline),
			PricePerUnit: &price,
		}

		err := service.CreateFuel(ctx, record)
		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
	})

	t.Run("returns error for non-existent vehicle", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewFuelService(fuelRepo, vehicleRepo, nil)

		record := &models.FuelRecord{
			VehicleID: "non-existent",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  string(models.FuelTypeGasoline),
		}

		err := service.CreateFuel(ctx, record)
		require.Error(t, err)
	})
}

func TestFuelService_GetFuel(t *testing.T) {
	ctx := context.Background()

	t.Run("gets fuel record successfully", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		fuelRepo.records["fuel-123"] = &models.FuelRecord{
			ID:        "fuel-123",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  string(models.FuelTypeGasoline),
		}
		service := NewFuelService(fuelRepo, vehicleRepo, nil)

		result, err := service.GetFuel(ctx, "fuel-123")
		require.NoError(t, err)
		assert.Equal(t, "fuel-123", result.ID)
	})

	t.Run("returns error for non-existent record", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewFuelService(fuelRepo, vehicleRepo, nil)

		_, err := service.GetFuel(ctx, "non-existent")
		require.Error(t, err)
	})
}

func TestFuelService_UpdateFuel(t *testing.T) {
	ctx := context.Background()

	t.Run("updates fuel record successfully", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		price := 3.50
		fuelRepo.records["fuel-123"] = &models.FuelRecord{
			ID:           "fuel-123",
			VehicleID:    "vehicle-123",
			FillDate:     time.Now().Add(-24 * time.Hour),
			Mileage:      50000,
			Volume:       12.5,
			FuelType:     string(models.FuelTypeGasoline),
			PricePerUnit: &price,
		}
		service := NewFuelService(fuelRepo, vehicleRepo, nil)

		newVolume := 15.0
		result, err := service.UpdateFuel(ctx, "fuel-123", FuelUpdates{Volume: &newVolume})
		require.NoError(t, err)
		assert.Equal(t, 15.0, result.Volume)
	})

	t.Run("returns error for non-existent record", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewFuelService(fuelRepo, vehicleRepo, nil)

		newVolume := 15.0
		_, err := service.UpdateFuel(ctx, "non-existent", FuelUpdates{Volume: &newVolume})
		require.Error(t, err)
	})
}

func TestFuelService_DeleteFuel(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes fuel record successfully", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		fuelRepo.records["fuel-123"] = &models.FuelRecord{
			ID:        "fuel-123",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  string(models.FuelTypeGasoline),
		}
		service := NewFuelService(fuelRepo, vehicleRepo, nil)

		err := service.DeleteFuel(ctx, "fuel-123")
		require.NoError(t, err)
	})

	t.Run("returns error for non-existent record", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewFuelService(fuelRepo, vehicleRepo, nil)

		err := service.DeleteFuel(ctx, "non-existent")
		require.Error(t, err)
	})
}

func TestFuelService_MetricsRecalculation(t *testing.T) {
	ctx := context.Background()

	t.Run("recalculates metrics on create", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		metricsRepo := newMockMetricsRepository()
		vehicleRepo.vehicles["vehicle-123"] = &models.Vehicle{
			ID: "vehicle-123", UserID: "user-123", VIN: "1HGBH41JXMN109186",
			Make: "Honda", Model: "Civic", Year: 2021, Status: models.VehicleStatusActive,
		}
		service := NewFuelService(fuelRepo, vehicleRepo, metricsRepo)

		price := 3.50
		record := &models.FuelRecord{
			VehicleID:    "vehicle-123",
			FillDate:     time.Now().Add(-24 * time.Hour),
			Mileage:      50000,
			Volume:       10.0,
			FuelType:     string(models.FuelTypeGasoline),
			PricePerUnit: &price,
		}

		err := service.CreateFuel(ctx, record)
		require.NoError(t, err)

		// Verify metrics were updated
		metrics, ok := metricsRepo.metrics["vehicle-123"]
		require.True(t, ok)
		require.NotNil(t, metrics.TotalFuelSpent)
		assert.InDelta(t, 35.00, *metrics.TotalFuelSpent, 0.01)
	})

	t.Run("recalculates metrics on update", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		metricsRepo := newMockMetricsRepository()
		price := 3.50
		existingRecord := &models.FuelRecord{
			ID: "fuel-123", VehicleID: "vehicle-123",
			FillDate: time.Now().Add(-24 * time.Hour),
			Mileage:  50000, Volume: 10.0,
			FuelType:     string(models.FuelTypeGasoline),
			PricePerUnit: &price,
		}
		fuelRepo.records[existingRecord.ID] = existingRecord
		service := NewFuelService(fuelRepo, vehicleRepo, metricsRepo)

		newPrice := 4.00
		_, err := service.UpdateFuel(ctx, "fuel-123", FuelUpdates{PricePerUnit: &newPrice})
		require.NoError(t, err)

		// Verify metrics were updated
		metrics, ok := metricsRepo.metrics["vehicle-123"]
		require.True(t, ok)
		require.NotNil(t, metrics.TotalFuelSpent)
		assert.InDelta(t, 40.00, *metrics.TotalFuelSpent, 0.01)
	})

	t.Run("recalculates metrics on delete", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		metricsRepo := newMockMetricsRepository()
		price := 3.50
		existingRecord := &models.FuelRecord{
			ID: "fuel-123", VehicleID: "vehicle-123",
			FillDate: time.Now().Add(-24 * time.Hour),
			Mileage:  50000, Volume: 10.0,
			FuelType:     string(models.FuelTypeGasoline),
			PricePerUnit: &price,
		}
		fuelRepo.records[existingRecord.ID] = existingRecord
		service := NewFuelService(fuelRepo, vehicleRepo, metricsRepo)

		err := service.DeleteFuel(ctx, "fuel-123")
		require.NoError(t, err)

		// Verify metrics were updated (should be 0 since no more records)
		metrics, ok := metricsRepo.metrics["vehicle-123"]
		require.True(t, ok)
		require.NotNil(t, metrics.TotalFuelSpent)
		assert.Equal(t, 0.0, *metrics.TotalFuelSpent)
	})

	t.Run("handles nil metrics repo gracefully", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		vehicleRepo.vehicles["vehicle-123"] = &models.Vehicle{
			ID: "vehicle-123", UserID: "user-123", VIN: "1HGBH41JXMN109186",
			Make: "Honda", Model: "Civic", Year: 2021, Status: models.VehicleStatusActive,
		}
		service := NewFuelService(fuelRepo, vehicleRepo, nil)

		record := &models.FuelRecord{
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  string(models.FuelTypeGasoline),
		}

		// Should not panic even with nil metricsRepo
		err := service.CreateFuel(ctx, record)
		require.NoError(t, err)
	})
}

func TestFuelService_ListFuel(t *testing.T) {
	ctx := context.Background()

	fuelRepo := newMockFuelRepository()
	vehicleRepo := newMockVehicleRepository()
	service := NewFuelService(fuelRepo, vehicleRepo, nil)

	fuelRepo.records["fuel-1"] = &models.FuelRecord{
		ID: "fuel-1", VehicleID: "vehicle-123",
		FillDate: time.Now().Add(-24 * time.Hour),
		Mileage:  50000, Volume: 12.5,
		FuelType: string(models.FuelTypeGasoline),
	}

	results, err := service.ListFuel(ctx, repositories.FuelFilters{}, repositories.PaginationParams{})
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestFuelService_CountFuel(t *testing.T) {
	ctx := context.Background()

	fuelRepo := newMockFuelRepository()
	vehicleRepo := newMockVehicleRepository()
	service := NewFuelService(fuelRepo, vehicleRepo, nil)

	fuelRepo.records["fuel-1"] = &models.FuelRecord{
		ID: "fuel-1", VehicleID: "vehicle-123",
		FillDate: time.Now().Add(-24 * time.Hour),
		Mileage:  50000, Volume: 12.5,
		FuelType: string(models.FuelTypeGasoline),
	}

	count, err := service.CountFuel(ctx, repositories.FuelFilters{})
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
