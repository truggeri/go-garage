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

// mockMaintenanceRepository is a mock implementation of MaintenanceRepository for testing
type mockMaintenanceRepository struct {
	records    map[string]*models.MaintenanceRecord
	createErr  error
	findErr    error
	updateErr  error
	deleteErr  error
	listResult []*models.MaintenanceRecord
}

func newMockMaintenanceRepository() *mockMaintenanceRepository {
	return &mockMaintenanceRepository{
		records: make(map[string]*models.MaintenanceRecord),
	}
}

func (m *mockMaintenanceRepository) Create(ctx context.Context, record *models.MaintenanceRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	if record.ID == "" {
		record.ID = "test-maintenance-id"
	}
	m.records[record.ID] = record
	return nil
}

func (m *mockMaintenanceRepository) FindByID(ctx context.Context, id string) (*models.MaintenanceRecord, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	record, exists := m.records[id]
	if !exists {
		return nil, models.NewNotFoundError("MaintenanceRecord", id)
	}
	return record, nil
}

func (m *mockMaintenanceRepository) FindByVehicleID(ctx context.Context, vehicleID string) ([]*models.MaintenanceRecord, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []*models.MaintenanceRecord
	for _, r := range m.records {
		if r.VehicleID == vehicleID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockMaintenanceRepository) Update(ctx context.Context, record *models.MaintenanceRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, exists := m.records[record.ID]; !exists {
		return models.NewNotFoundError("MaintenanceRecord", record.ID)
	}
	m.records[record.ID] = record
	return nil
}

func (m *mockMaintenanceRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, exists := m.records[id]; !exists {
		return models.NewNotFoundError("MaintenanceRecord", id)
	}
	delete(m.records, id)
	return nil
}

func (m *mockMaintenanceRepository) List(ctx context.Context, filters repositories.MaintenanceFilters, pagination repositories.PaginationParams) ([]*models.MaintenanceRecord, error) {
	if m.listResult != nil {
		return m.listResult, nil
	}
	var result []*models.MaintenanceRecord
	for _, r := range m.records {
		result = append(result, r)
	}
	return result, nil
}

func TestMaintenanceService_CreateMaintenance(t *testing.T) {
	ctx := context.Background()

	t.Run("creates maintenance record successfully", func(t *testing.T) {
		maintenanceRepo := newMockMaintenanceRepository()
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
		service := NewMaintenanceService(maintenanceRepo, vehicleRepo)

		record := &models.MaintenanceRecord{
			VehicleID:   "vehicle-123",
			ServiceType: "Oil Change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}

		err := service.CreateMaintenance(ctx, record)
		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
	})

	t.Run("returns error for non-existent vehicle", func(t *testing.T) {
		maintenanceRepo := newMockMaintenanceRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewMaintenanceService(maintenanceRepo, vehicleRepo)

		record := &models.MaintenanceRecord{
			VehicleID:   "non-existent-vehicle",
			ServiceType: "Oil Change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}

		err := service.CreateMaintenance(ctx, record)
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestMaintenanceService_GetMaintenance(t *testing.T) {
	ctx := context.Background()

	t.Run("returns maintenance record by ID", func(t *testing.T) {
		maintenanceRepo := newMockMaintenanceRepository()
		vehicleRepo := newMockVehicleRepository()
		existingRecord := &models.MaintenanceRecord{
			ID:          "record-123",
			VehicleID:   "vehicle-123",
			ServiceType: "Oil Change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}
		maintenanceRepo.records[existingRecord.ID] = existingRecord
		service := NewMaintenanceService(maintenanceRepo, vehicleRepo)

		record, err := service.GetMaintenance(ctx, "record-123")
		require.NoError(t, err)
		assert.Equal(t, "Oil Change", record.ServiceType)
	})

	t.Run("returns not found error for non-existent record", func(t *testing.T) {
		maintenanceRepo := newMockMaintenanceRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewMaintenanceService(maintenanceRepo, vehicleRepo)

		_, err := service.GetMaintenance(ctx, "non-existent")
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestMaintenanceService_GetVehicleMaintenance(t *testing.T) {
	ctx := context.Background()

	t.Run("returns all maintenance records for vehicle", func(t *testing.T) {
		maintenanceRepo := newMockMaintenanceRepository()
		vehicleRepo := newMockVehicleRepository()
		record1 := &models.MaintenanceRecord{
			ID:          "record-1",
			VehicleID:   "vehicle-123",
			ServiceType: "Oil Change",
			ServiceDate: time.Now().Add(-48 * time.Hour),
		}
		record2 := &models.MaintenanceRecord{
			ID:          "record-2",
			VehicleID:   "vehicle-123",
			ServiceType: "Tire Rotation",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}
		maintenanceRepo.records[record1.ID] = record1
		maintenanceRepo.records[record2.ID] = record2
		service := NewMaintenanceService(maintenanceRepo, vehicleRepo)

		records, err := service.GetVehicleMaintenance(ctx, "vehicle-123")
		require.NoError(t, err)
		assert.Len(t, records, 2)
	})

	t.Run("returns empty slice for vehicle with no records", func(t *testing.T) {
		maintenanceRepo := newMockMaintenanceRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewMaintenanceService(maintenanceRepo, vehicleRepo)

		records, err := service.GetVehicleMaintenance(ctx, "vehicle-with-no-records")
		require.NoError(t, err)
		assert.Empty(t, records)
	})
}

func TestMaintenanceService_UpdateMaintenance(t *testing.T) {
	ctx := context.Background()

	t.Run("updates maintenance record fields", func(t *testing.T) {
		maintenanceRepo := newMockMaintenanceRepository()
		vehicleRepo := newMockVehicleRepository()
		existingRecord := &models.MaintenanceRecord{
			ID:          "record-123",
			VehicleID:   "vehicle-123",
			ServiceType: "Oil Change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}
		maintenanceRepo.records[existingRecord.ID] = existingRecord
		service := NewMaintenanceService(maintenanceRepo, vehicleRepo)

		newServiceType := "Synthetic Oil Change"
		newCost := 89.99
		updates := MaintenanceUpdates{
			ServiceType: &newServiceType,
			Cost:        &newCost,
		}

		updatedRecord, err := service.UpdateMaintenance(ctx, "record-123", updates)
		require.NoError(t, err)
		assert.Equal(t, "Synthetic Oil Change", updatedRecord.ServiceType)
		assert.Equal(t, 89.99, *updatedRecord.Cost)
	})

	t.Run("updates service date field", func(t *testing.T) {
		maintenanceRepo := newMockMaintenanceRepository()
		vehicleRepo := newMockVehicleRepository()
		existingRecord := &models.MaintenanceRecord{
			ID:          "record-123",
			VehicleID:   "vehicle-123",
			ServiceType: "Oil Change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}
		maintenanceRepo.records[existingRecord.ID] = existingRecord
		service := NewMaintenanceService(maintenanceRepo, vehicleRepo)

		newServiceDate := time.Now().Add(-48 * time.Hour)
		updates := MaintenanceUpdates{
			ServiceDate: &newServiceDate,
		}

		updatedRecord, err := service.UpdateMaintenance(ctx, "record-123", updates)
		require.NoError(t, err)
		assert.Equal(t, newServiceDate.Unix(), updatedRecord.ServiceDate.Unix())
		assert.Equal(t, "Oil Change", updatedRecord.ServiceType) // unchanged
	})

	t.Run("partial update preserves unchanged fields", func(t *testing.T) {
		maintenanceRepo := newMockMaintenanceRepository()
		vehicleRepo := newMockVehicleRepository()
		mileage := 50000
		existingRecord := &models.MaintenanceRecord{
			ID:               "record-123",
			VehicleID:        "vehicle-123",
			ServiceType:      "Oil Change",
			ServiceDate:      time.Now().Add(-24 * time.Hour),
			MileageAtService: &mileage,
			ServiceProvider:  "Quick Lube",
		}
		maintenanceRepo.records[existingRecord.ID] = existingRecord
		service := NewMaintenanceService(maintenanceRepo, vehicleRepo)

		newNotes := "Used synthetic oil"
		updates := MaintenanceUpdates{
			Notes: &newNotes,
		}

		updatedRecord, err := service.UpdateMaintenance(ctx, "record-123", updates)
		require.NoError(t, err)
		assert.Equal(t, "Oil Change", updatedRecord.ServiceType)     // unchanged
		assert.Equal(t, "Quick Lube", updatedRecord.ServiceProvider) // unchanged
		assert.Equal(t, 50000, *updatedRecord.MileageAtService)      // unchanged
		assert.Equal(t, "Used synthetic oil", updatedRecord.Notes)   // updated
	})

	t.Run("returns not found for non-existent record", func(t *testing.T) {
		maintenanceRepo := newMockMaintenanceRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewMaintenanceService(maintenanceRepo, vehicleRepo)

		_, err := service.UpdateMaintenance(ctx, "non-existent", MaintenanceUpdates{})
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

func TestMaintenanceService_DeleteMaintenance(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes maintenance record successfully", func(t *testing.T) {
		maintenanceRepo := newMockMaintenanceRepository()
		vehicleRepo := newMockVehicleRepository()
		existingRecord := &models.MaintenanceRecord{
			ID:          "record-123",
			VehicleID:   "vehicle-123",
			ServiceType: "Oil Change",
			ServiceDate: time.Now().Add(-24 * time.Hour),
		}
		maintenanceRepo.records[existingRecord.ID] = existingRecord
		service := NewMaintenanceService(maintenanceRepo, vehicleRepo)

		err := service.DeleteMaintenance(ctx, "record-123")
		require.NoError(t, err)

		// Verify record is deleted
		_, exists := maintenanceRepo.records["record-123"]
		assert.False(t, exists)
	})

	t.Run("returns not found for non-existent record", func(t *testing.T) {
		maintenanceRepo := newMockMaintenanceRepository()
		vehicleRepo := newMockVehicleRepository()
		service := NewMaintenanceService(maintenanceRepo, vehicleRepo)

		err := service.DeleteMaintenance(ctx, "non-existent")
		assert.Error(t, err)

		var notFoundErr *models.NotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}
