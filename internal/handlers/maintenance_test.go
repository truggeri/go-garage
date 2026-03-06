package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
	"github.com/truggeri/go-garage/internal/services"
)

type stubMaintenanceSvc struct {
	createErr   error
	getResult   *models.MaintenanceRecord
	getErr      error
	listResult  []*models.MaintenanceRecord
	listErr     error
	countResult int
	countErr    error
	updateRes   *models.MaintenanceRecord
	updateErr   error
	deleteErr   error
}

func (s *stubMaintenanceSvc) CreateMaintenance(_ context.Context, m *models.MaintenanceRecord) error {
	if s.createErr != nil {
		return s.createErr
	}
	m.ID = "generated-id"
	return nil
}

func (s *stubMaintenanceSvc) GetMaintenance(_ context.Context, _ string) (*models.MaintenanceRecord, error) {
	return s.getResult, s.getErr
}

func (s *stubMaintenanceSvc) GetVehicleMaintenance(_ context.Context, _ string) ([]*models.MaintenanceRecord, error) {
	return s.listResult, s.listErr
}

func (s *stubMaintenanceSvc) ListMaintenance(_ context.Context, _ repositories.MaintenanceFilters, _ repositories.PaginationParams) ([]*models.MaintenanceRecord, error) {
	return s.listResult, s.listErr
}

func (s *stubMaintenanceSvc) CountMaintenance(_ context.Context, _ repositories.MaintenanceFilters) (int, error) {
	return s.countResult, s.countErr
}

func (s *stubMaintenanceSvc) UpdateMaintenance(_ context.Context, _ string, _ services.MaintenanceUpdates) (*models.MaintenanceRecord, error) {
	return s.updateRes, s.updateErr
}

func (s *stubMaintenanceSvc) DeleteMaintenance(_ context.Context, _ string) error {
	return s.deleteErr
}

func TestMaintenanceHandler_ListAll(t *testing.T) {
	serviceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("returns maintenance records for authenticated user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "u1", VIN: "ABC12345678901234", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		maintStub := &stubMaintenanceSvc{
			countResult: 1,
			listResult: []*models.MaintenanceRecord{
				{ID: "m1", VehicleID: "v1", ServiceType: models.ServiceTypeOilChange, ServiceDate: serviceDate},
			},
		}
		h := MakeMaintenanceAPIHandler(maintStub, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/v1/maintenance", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"vehicleId": "v1"})
		rec := httptest.NewRecorder()

		h.ListAll(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.True(t, resp["success"].(bool))
		data := resp["data"].([]interface{})
		assert.Len(t, data, 1)
	})

	t.Run("rejects unauthenticated request", func(t *testing.T) {
		h := MakeMaintenanceAPIHandler(&stubMaintenanceSvc{}, &stubVehicleSvc{})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/v1/maintenance", nil)
		req = mux.SetURLVars(req, map[string]string{"vehicleId": "v1"})
		rec := httptest.NewRecorder()

		h.ListAll(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("returns forbidden for vehicle owned by another user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", VIN: "ABC12345678901234", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		h := MakeMaintenanceAPIHandler(&stubMaintenanceSvc{}, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/v1/maintenance", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"vehicleId": "v1"})
		rec := httptest.NewRecorder()

		h.ListAll(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns not found for non-existent vehicle", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getErr: models.NewNotFoundError("Vehicle", "v999"),
		}
		h := MakeMaintenanceAPIHandler(&stubMaintenanceSvc{}, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/v999/maintenance", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"vehicleId": "v999"})
		rec := httptest.NewRecorder()

		h.ListAll(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestMaintenanceHandler_CreateOne(t *testing.T) {
	t.Run("creates maintenance record with valid input", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "u1", VIN: "ABC12345678901234", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		maintStub := &stubMaintenanceSvc{}
		h := MakeMaintenanceAPIHandler(maintStub, vehicleStub)

		body := map[string]interface{}{
			"service_type": "oil_change", "service_date": "2024-01-15",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/v1/maintenance", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"vehicleId": "v1"})
		rec := httptest.NewRecorder()

		h.CreateOne(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.True(t, resp["success"].(bool))
		assert.Equal(t, "Maintenance record created successfully", resp["message"])
	})

	t.Run("rejects missing required fields", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "u1", VIN: "ABC12345678901234", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		h := MakeMaintenanceAPIHandler(&stubMaintenanceSvc{}, vehicleStub)

		body := map[string]interface{}{"service_type": "oil_change"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/v1/maintenance", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"vehicleId": "v1"})
		rec := httptest.NewRecorder()

		h.CreateOne(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects unauthenticated request", func(t *testing.T) {
		h := MakeMaintenanceAPIHandler(&stubMaintenanceSvc{}, &stubVehicleSvc{})

		body := map[string]interface{}{"service_type": "oil_change", "service_date": "2024-01-15"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/v1/maintenance", bytes.NewReader(jsonBody))
		req = mux.SetURLVars(req, map[string]string{"vehicleId": "v1"})
		rec := httptest.NewRecorder()

		h.CreateOne(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("returns forbidden for vehicle owned by another user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", VIN: "ABC12345678901234", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		h := MakeMaintenanceAPIHandler(&stubMaintenanceSvc{}, vehicleStub)

		body := map[string]interface{}{"service_type": "oil_change", "service_date": "2024-01-15"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/v1/maintenance", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"vehicleId": "v1"})
		rec := httptest.NewRecorder()

		h.CreateOne(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestMaintenanceHandler_GetOne(t *testing.T) {
	serviceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("returns maintenance record when owned by user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "u1", VIN: "ABC12345678901234", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		maintStub := &stubMaintenanceSvc{
			getResult: &models.MaintenanceRecord{ID: "m1", VehicleID: "v1", ServiceType: models.ServiceTypeOilChange, ServiceDate: serviceDate},
		}
		h := MakeMaintenanceAPIHandler(maintStub, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/maintenance/m1", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.True(t, resp["success"].(bool))
	})

	t.Run("returns forbidden for record on vehicle owned by another user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", VIN: "ABC12345678901234", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		maintStub := &stubMaintenanceSvc{
			getResult: &models.MaintenanceRecord{ID: "m1", VehicleID: "v1", ServiceType: models.ServiceTypeOilChange, ServiceDate: serviceDate},
		}
		h := MakeMaintenanceAPIHandler(maintStub, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/maintenance/m1", nil)
		req = addAuthContext(req, "different-user", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns not found for non-existent record", func(t *testing.T) {
		maintStub := &stubMaintenanceSvc{
			getErr: models.NewNotFoundError("MaintenanceRecord", "m999"),
		}
		h := MakeMaintenanceAPIHandler(maintStub, &stubVehicleSvc{})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/maintenance/m999", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "m999"})
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestMaintenanceHandler_ReplaceOne(t *testing.T) {
	serviceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("updates maintenance record successfully", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "u1", VIN: "ABC12345678901234", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		cost := 89.99
		maintStub := &stubMaintenanceSvc{
			getResult: &models.MaintenanceRecord{ID: "m1", VehicleID: "v1", ServiceType: models.ServiceTypeOilChange, ServiceDate: serviceDate},
			updateRes: &models.MaintenanceRecord{ID: "m1", VehicleID: "v1", ServiceType: models.ServiceTypeBrakes, ServiceDate: serviceDate, Cost: &cost},
		}
		h := MakeMaintenanceAPIHandler(maintStub, vehicleStub)

		body := map[string]interface{}{"service_type": "brakes", "cost": 89.99}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/maintenance/m1", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		rec := httptest.NewRecorder()

		h.ReplaceOne(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "Maintenance record updated successfully", resp["message"])
	})

	t.Run("returns forbidden for record on vehicle owned by another user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", VIN: "ABC12345678901234", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		maintStub := &stubMaintenanceSvc{
			getResult: &models.MaintenanceRecord{ID: "m1", VehicleID: "v1", ServiceType: models.ServiceTypeOilChange, ServiceDate: serviceDate},
		}
		h := MakeMaintenanceAPIHandler(maintStub, vehicleStub)

		body := map[string]interface{}{"service_type": "brakes"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/maintenance/m1", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		rec := httptest.NewRecorder()

		h.ReplaceOne(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestMaintenanceHandler_RemoveOne(t *testing.T) {
	serviceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("deletes maintenance record successfully", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "u1", VIN: "ABC12345678901234", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		maintStub := &stubMaintenanceSvc{
			getResult: &models.MaintenanceRecord{ID: "m1", VehicleID: "v1", ServiceType: models.ServiceTypeOilChange, ServiceDate: serviceDate},
		}
		h := MakeMaintenanceAPIHandler(maintStub, vehicleStub)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/maintenance/m1", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		rec := httptest.NewRecorder()

		h.RemoveOne(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "Maintenance record deleted successfully", resp["message"])
	})

	t.Run("returns forbidden for record on vehicle owned by another user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", VIN: "ABC12345678901234", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		maintStub := &stubMaintenanceSvc{
			getResult: &models.MaintenanceRecord{ID: "m1", VehicleID: "v1", ServiceType: models.ServiceTypeOilChange, ServiceDate: serviceDate},
		}
		h := MakeMaintenanceAPIHandler(maintStub, vehicleStub)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/maintenance/m1", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		rec := httptest.NewRecorder()

		h.RemoveOne(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns not found for non-existent record", func(t *testing.T) {
		maintStub := &stubMaintenanceSvc{
			getErr: models.NewNotFoundError("MaintenanceRecord", "m999"),
		}
		h := MakeMaintenanceAPIHandler(maintStub, &stubVehicleSvc{})

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/maintenance/m999", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "m999"})
		rec := httptest.NewRecorder()

		h.RemoveOne(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
