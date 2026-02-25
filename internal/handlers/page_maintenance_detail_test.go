package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/templateengine"
)

func newTestMaintenanceDetailPageHandler(
	t *testing.T,
	vehicleSvc *stubVehicleSvc,
	maintenanceSvc *stubMaintenanceSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, maintenanceSvc)
}

func TestPageHandler_MaintenanceDetail(t *testing.T) {
	serviceDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	cost := 99.99
	mileage := 30000

	baseRecord := &models.MaintenanceRecord{
		ID: "m1", VehicleID: "v1", ServiceType: "Oil Change",
		ServiceDate: serviceDate, Cost: &cost, MileageAtService: &mileage,
		ServiceProvider: "Quick Lube", Notes: "Used synthetic oil",
	}
	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}

	t.Run("renders maintenance detail page for authenticated user", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestMaintenanceDetailPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance/m1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceDetail(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Oil Change")
		assert.Contains(t, body, "Quick Lube")
		assert.Contains(t, body, "Used synthetic oil")
		assert.Contains(t, body, "Ford")
		assert.Contains(t, body, "Focus")
		assert.Contains(t, body, "2020")
	})

	t.Run("returns 404 when record not found", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getErr: models.NewNotFoundError("MaintenanceRecord", "m99")}
		handler := newTestMaintenanceDetailPageHandler(t, &stubVehicleSvc{}, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance/m99", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "m99"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceDetail(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 403 when vehicle belongs to another user", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		handler := newTestMaintenanceDetailPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance/m1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceDetail(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestMaintenanceDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/maintenance/m1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		rec := httptest.NewRecorder()

		handler.MaintenanceDetail(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when maintenance service fails", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getErr: models.NewDatabaseError("get maintenance", assert.AnError)}
		handler := newTestMaintenanceDetailPageHandler(t, &stubVehicleSvc{}, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance/m1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceDetail(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when vehicle service fails", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{getErr: models.NewDatabaseError("get vehicle", assert.AnError)}
		handler := newTestMaintenanceDetailPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance/m1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceDetail(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
