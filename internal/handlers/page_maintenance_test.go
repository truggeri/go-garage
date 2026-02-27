package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/templateengine"
)

func newTestMaintenanceListPageHandler(
	t *testing.T,
	vehicleSvc *stubVehicleSvc,
	maintenanceSvc *stubMaintenanceSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, maintenanceSvc, nil)
}

func TestPageHandler_MaintenanceList(t *testing.T) {
	serviceDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	cost := 99.99
	mileage := 30000

	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}
	baseRecord := &models.MaintenanceRecord{
		ID: "m1", VehicleID: "v1", ServiceType: "Oil Change",
		ServiceDate: serviceDate, Cost: &cost, MileageAtService: &mileage,
		ServiceProvider: "Quick Lube",
	}

	t.Run("renders maintenance list for authenticated user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		maintenanceStub := &stubMaintenanceSvc{listResult: []*models.MaintenanceRecord{baseRecord}}
		handler := newTestMaintenanceListPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Maintenance Records")
		assert.Contains(t, body, "Oil Change")
		assert.Contains(t, body, "Quick Lube")
		assert.Contains(t, body, "Ford Focus")
	})

	t.Run("renders empty state when no records", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		maintenanceStub := &stubMaintenanceSvc{listResult: []*models.MaintenanceRecord{}}
		handler := newTestMaintenanceListPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "No maintenance records yet")
	})

	t.Run("renders empty state with no vehicles", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{}}
		maintenanceStub := &stubMaintenanceSvc{}
		handler := newTestMaintenanceListPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Add a Vehicle")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestMaintenanceListPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/maintenance", nil)
		rec := httptest.NewRecorder()

		handler.MaintenanceList(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when vehicle service fails", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listErr: models.NewDatabaseError("get vehicles", assert.AnError)}
		handler := newTestMaintenanceListPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/maintenance", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceList(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("filters by vehicle ID", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			listResult:  []*models.Vehicle{baseVehicle},
			countResult: 1,
		}
		maintenanceStub := &stubMaintenanceSvc{
			countResult: 1,
			listResult:  []*models.MaintenanceRecord{baseRecord},
		}
		handler := newTestMaintenanceListPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance?vehicle=v1", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Oil Change")
	})

	t.Run("returns 403 when filtering by vehicle not owned by user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			listResult: []*models.Vehicle{baseVehicle},
		}
		handler := newTestMaintenanceListPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/maintenance?vehicle=other-vehicle-id", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceList(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("shows success flash when added=true", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		maintenanceStub := &stubMaintenanceSvc{listResult: []*models.MaintenanceRecord{baseRecord}}
		handler := newTestMaintenanceListPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance?added=true", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Maintenance record added successfully")
	})

	t.Run("returns 500 when vehicle-filtered maintenance count fails", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		maintenanceStub := &stubMaintenanceSvc{
			countErr: models.NewDatabaseError("count maintenance", assert.AnError),
		}
		handler := newTestMaintenanceListPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance?vehicle=v1", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceList(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("filters by service type when no vehicle selected", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		maintenanceStub := &stubMaintenanceSvc{
			listResult: []*models.MaintenanceRecord{
				baseRecord,
				{ID: "m2", VehicleID: "v1", ServiceType: "Tire Rotation", ServiceDate: serviceDate},
			},
		}
		handler := newTestMaintenanceListPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance?service_type=oil", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Oil Change")
		assert.NotContains(t, body, "Tire Rotation")
	})
}
