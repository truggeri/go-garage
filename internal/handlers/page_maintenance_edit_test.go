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

func newTestMaintenanceEditPageHandler(
	t *testing.T,
	vehicleSvc *stubVehicleSvc,
	maintenanceSvc *stubMaintenanceSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, maintenanceSvc, nil, nil, nil)
}

func TestPageHandler_MaintenanceEdit(t *testing.T) {
	serviceDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	cost := 99.99
	mileage := 30000

	baseRecord := &models.MaintenanceRecord{
		ID: "m1", VehicleID: "v1", ServiceType: "oil_change",
		ServiceDate: serviceDate, Cost: &cost, MileageAtService: &mileage,
		ServiceProvider: "Quick Lube", Notes: "Used synthetic oil",
	}
	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}

	t.Run("renders edit form with pre-populated data", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestMaintenanceEditPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance/m1/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		serveHTMXPage(t, handler.MaintenanceEdit, rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Edit Maintenance Record")
		assert.Contains(t, body, `hx-put="/api/v1/maintenance/m1"`)
		assert.Contains(t, body, `hx-ext="json-enc"`)
		assert.Contains(t, body, `hx-swap="none"`)
		assert.Contains(t, body, `hx-validate="true"`)
		assert.NotContains(t, body, `novalidate`)
		assert.Contains(t, body, `data-form-errors role="alert" aria-live="assertive"`)
		assert.NotContains(t, body, `action="/maintenance/m1/edit"`)
		for _, field := range []string{"service_type", "custom_service_type", "service_date", "mileage_at_service", "cost", "service_provider", "notes"} {
			assert.Contains(t, body, `data-field="`+field+`"`)
		}
		assert.Contains(t, body, "Oil Change")
		assert.Contains(t, body, "2024-06-01")
		assert.Contains(t, body, "30000")
		assert.Contains(t, body, "99.99")
		assert.Contains(t, body, "Quick Lube")
		assert.Contains(t, body, "Used synthetic oil")
		assert.Contains(t, body, "2020 Ford Focus")
	})

	t.Run("returns 404 when record not found", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getErr: models.NewNotFoundError("MaintenanceRecord", "m99")}
		handler := newTestMaintenanceEditPageHandler(t, &stubVehicleSvc{}, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance/m99/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "m99"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceEdit(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 403 when vehicle belongs to another user", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		handler := newTestMaintenanceEditPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance/m1/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceEdit(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestMaintenanceEditPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/maintenance/m1/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		rec := httptest.NewRecorder()

		handler.MaintenanceEdit(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when maintenance service fails", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getErr: models.NewDatabaseError("get maintenance", assert.AnError)}
		handler := newTestMaintenanceEditPageHandler(t, &stubVehicleSvc{}, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/maintenance/m1/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceEdit(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
