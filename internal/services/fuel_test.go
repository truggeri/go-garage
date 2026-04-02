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

func TestFuelService_CreateFuel(t *testing.T) {
	ctx := context.Background()

	t.Run("creates fuel record successfully", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		vehicleRepo.vehicles["vehicle-123"] = &models.Vehicle{
			ID:     "vehicle-123",
			UserID: "user-123",
			VIN:    "1HGBH41JXMN109186",
			Make:   "Honda",
			Model:  "Civic",
			Year:   2021,
			Status: models.VehicleStatusActive,
		}
		service := NewFuelService(fuelRepo, vehicleRepo)

		record := &models.FuelRecord{
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  "gasoline",
		}

		err := service.CreateFuel(ctx, record)
		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
	})

	t.Run("returns error for non-existent vehicle", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewFuelService(fuelRepo, vehicleRepo)

		record := &models.FuelRecord{
			VehicleID: "non-existent-vehicle",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  "gasoline",
		}

		err := service.CreateFuel(ctx, record)
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestFuelService_GetFuel(t *testing.T) {
	ctx := context.Background()

	t.Run("returns fuel record by ID", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		existingRecord := &models.FuelRecord{
			ID:        "record-123",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  "gasoline",
		}
		fuelRepo.records[existingRecord.ID] = existingRecord
		service := NewFuelService(fuelRepo, vehicleRepo)

		record, err := service.GetFuel(ctx, "record-123")
		require.NoError(t, err)
		assert.Equal(t, "gasoline", record.FuelType)
		assert.Equal(t, 50000, record.Mileage)
	})

	t.Run("returns not found error for non-existent record", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewFuelService(fuelRepo, vehicleRepo)

		_, err := service.GetFuel(ctx, "non-existent")
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestFuelService_GetVehicleFuel(t *testing.T) {
	ctx := context.Background()

	t.Run("returns all fuel records for vehicle", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		record1 := &models.FuelRecord{
			ID:        "record-1",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-48 * time.Hour),
			Mileage:   49500,
			Volume:    11.0,
			FuelType:  "gasoline",
		}
		record2 := &models.FuelRecord{
			ID:        "record-2",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  "gasoline",
		}
		fuelRepo.records[record1.ID] = record1
		fuelRepo.records[record2.ID] = record2
		service := NewFuelService(fuelRepo, vehicleRepo)

		records, err := service.GetVehicleFuel(ctx, "vehicle-123")
		require.NoError(t, err)
		assert.Len(t, records, 2)
	})

	t.Run("returns empty slice for vehicle with no records", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewFuelService(fuelRepo, vehicleRepo)

		records, err := service.GetVehicleFuel(ctx, "vehicle-with-no-records")
		require.NoError(t, err)
		assert.Empty(t, records)
	})
}

func TestFuelService_UpdateFuel(t *testing.T) {
	ctx := context.Background()

	t.Run("updates fuel record fields", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		existingRecord := &models.FuelRecord{
			ID:        "record-123",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  "gasoline",
		}
		fuelRepo.records[existingRecord.ID] = existingRecord
		service := NewFuelService(fuelRepo, vehicleRepo)

		newVolume := 15.0
		newFuelType := "diesel"
		updates := FuelUpdates{
			Volume:   &newVolume,
			FuelType: &newFuelType,
		}

		updatedRecord, err := service.UpdateFuel(ctx, "record-123", updates)
		require.NoError(t, err)
		assert.Equal(t, 15.0, updatedRecord.Volume)
		assert.Equal(t, "diesel", updatedRecord.FuelType)
	})

	t.Run("updates fill date field", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		existingRecord := &models.FuelRecord{
			ID:        "record-123",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  "gasoline",
		}
		fuelRepo.records[existingRecord.ID] = existingRecord
		service := NewFuelService(fuelRepo, vehicleRepo)

		newFillDate := time.Now().Add(-48 * time.Hour)
		updates := FuelUpdates{
			FillDate: &newFillDate,
		}

		updatedRecord, err := service.UpdateFuel(ctx, "record-123", updates)
		require.NoError(t, err)
		assert.Equal(t, newFillDate.Unix(), updatedRecord.FillDate.Unix())
		assert.Equal(t, "gasoline", updatedRecord.FuelType) // unchanged
	})

	t.Run("partial update preserves unchanged fields", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		pricePerUnit := 3.99
		existingRecord := &models.FuelRecord{
			ID:           "record-123",
			VehicleID:    "vehicle-123",
			FillDate:     time.Now().Add(-24 * time.Hour),
			Mileage:      50000,
			Volume:       12.5,
			FuelType:     "gasoline",
			PricePerUnit: &pricePerUnit,
			Location:     "Shell Station",
		}
		fuelRepo.records[existingRecord.ID] = existingRecord
		service := NewFuelService(fuelRepo, vehicleRepo)

		newNotes := "Highway trip"
		updates := FuelUpdates{
			Notes: &newNotes,
		}

		updatedRecord, err := service.UpdateFuel(ctx, "record-123", updates)
		require.NoError(t, err)
		assert.Equal(t, "gasoline", updatedRecord.FuelType)      // unchanged
		assert.Equal(t, "Shell Station", updatedRecord.Location) // unchanged
		assert.Equal(t, 3.99, *updatedRecord.PricePerUnit)       // unchanged
		assert.Equal(t, 50000, updatedRecord.Mileage)            // unchanged
		assert.Equal(t, "Highway trip", updatedRecord.Notes)     // updated
	})

	t.Run("returns not found for non-existent record", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewFuelService(fuelRepo, vehicleRepo)

		_, err := service.UpdateFuel(ctx, "non-existent", FuelUpdates{})
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestFuelService_DeleteFuel(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes fuel record successfully", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		existingRecord := &models.FuelRecord{
			ID:        "record-123",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  "gasoline",
		}
		fuelRepo.records[existingRecord.ID] = existingRecord
		service := NewFuelService(fuelRepo, vehicleRepo)

		err := service.DeleteFuel(ctx, "record-123")
		require.NoError(t, err)

		// Verify record is deleted
		_, exists := fuelRepo.records["record-123"]
		assert.False(t, exists)
	})

	t.Run("returns not found for non-existent record", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewFuelService(fuelRepo, vehicleRepo)

		err := service.DeleteFuel(ctx, "non-existent")
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestFuelService_ListFuel(t *testing.T) {
	ctx := context.Background()

	t.Run("lists fuel records with filters and pagination", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		record1 := &models.FuelRecord{
			ID:        "record-1",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  "gasoline",
		}
		fuelRepo.listResult = []*models.FuelRecord{record1}
		service := NewFuelService(fuelRepo, vehicleRepo)

		filters := repositories.FuelFilters{}
		pagination := repositories.PaginationParams{Limit: 10, Offset: 0}

		records, err := service.ListFuel(ctx, filters, pagination)
		require.NoError(t, err)
		assert.Len(t, records, 1)
	})
}

func TestFuelService_CountFuel(t *testing.T) {
	ctx := context.Background()

	t.Run("counts fuel records successfully", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		fuelRepo.records["record-1"] = &models.FuelRecord{
			ID:        "record-1",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   49500,
			Volume:    11.0,
			FuelType:  "gasoline",
		}
		fuelRepo.records["record-2"] = &models.FuelRecord{
			ID:        "record-2",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-48 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  "gasoline",
		}
		service := NewFuelService(fuelRepo, vehicleRepo)

		count, err := service.CountFuel(ctx, repositories.FuelFilters{})
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}

func TestFuelService_UpdateFuel_AllFields(t *testing.T) {
	ctx := context.Background()

	t.Run("updates mileage and optional fields", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		existingRecord := &models.FuelRecord{
			ID:        "record-123",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  "gasoline",
		}
		fuelRepo.records[existingRecord.ID] = existingRecord
		service := NewFuelService(fuelRepo, vehicleRepo)

		newMileage := 55000
		newBrand := "Shell"
		newPartialFill := true
		newOctane := 93
		newCityPct := 75
		newMPG := 28.5
		updates := FuelUpdates{
			Mileage:               &newMileage,
			Brand:                 &newBrand,
			PartialFill:           &newPartialFill,
			OctaneRating:          &newOctane,
			CityDrivingPercentage: &newCityPct,
			VehicleReportedMPG:    &newMPG,
		}

		updatedRecord, err := service.UpdateFuel(ctx, "record-123", updates)
		require.NoError(t, err)
		assert.Equal(t, 55000, updatedRecord.Mileage)
		assert.Equal(t, "Shell", updatedRecord.Brand)
		assert.True(t, updatedRecord.PartialFill)
		assert.Equal(t, 93, *updatedRecord.OctaneRating)
		assert.Equal(t, 75, *updatedRecord.CityDrivingPercentage)
		assert.Equal(t, 28.5, *updatedRecord.VehicleReportedMPG)
		assert.Equal(t, "gasoline", updatedRecord.FuelType) // unchanged
	})

	t.Run("returns error on update failure", func(t *testing.T) {
		fuelRepo := newMockFuelRepository()
		vehicleRepo := newMockVehicleRepository()
		existingRecord := &models.FuelRecord{
			ID:        "record-123",
			VehicleID: "vehicle-123",
			FillDate:  time.Now().Add(-24 * time.Hour),
			Mileage:   50000,
			Volume:    12.5,
			FuelType:  "gasoline",
		}
		fuelRepo.records[existingRecord.ID] = existingRecord
		fuelRepo.updateErr = models.NewDatabaseError("update", assert.AnError)
		service := NewFuelService(fuelRepo, vehicleRepo)

		_, err := service.UpdateFuel(ctx, "record-123", FuelUpdates{})
		assert.Error(t, err)
	})
}
