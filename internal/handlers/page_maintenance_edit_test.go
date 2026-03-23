package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, maintenanceSvc, nil, nil)
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

		handler.MaintenanceEdit(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Edit Maintenance Record")
		assert.Contains(t, body, `action="/maintenance/m1/edit"`)
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

func TestPageHandler_MaintenanceUpdate(t *testing.T) {
	serviceDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	cost := 99.99
	mileage := 30000

	baseRecord := &models.MaintenanceRecord{
		ID: "m1", VehicleID: "v1", ServiceType: "oil_change",
		ServiceDate: serviceDate, Cost: &cost, MileageAtService: &mileage,
		ServiceProvider: "Quick Lube",
	}
	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}

	postForm := func(vals url.Values) *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/maintenance/m1/edit",
			strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req
	}

	validForm := url.Values{
		"service_type": {"oil_change"},
		"service_date": {"2024-06-15"},
	}

	t.Run("redirects to detail page on success", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getResult: baseRecord, updateRes: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestMaintenanceEditPageHandler(t, vehicleStub, maintenanceStub)

		req := postForm(validForm)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceUpdate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/maintenance/m1?updated=true", rec.Header().Get("Location"))
	})

	t.Run("returns 400 when service date is invalid", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestMaintenanceEditPageHandler(t, vehicleStub, maintenanceStub)

		form := url.Values{
			"service_type": {"oil_change"},
			"service_date": {"not-a-date"},
		}
		req := postForm(form)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid date format")
	})

	t.Run("returns 400 when service type is missing", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestMaintenanceEditPageHandler(t, vehicleStub, maintenanceStub)

		form := url.Values{
			"service_type": {""},
			"service_date": {"2024-06-15"},
		}
		req := postForm(form)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("repopulates form fields on validation error", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestMaintenanceEditPageHandler(t, vehicleStub, maintenanceStub)

		form := url.Values{
			"service_type":       {"brakes"},
			"service_date":       {""},
			"mileage_at_service": {"50000"},
			"cost":               {"120.50"},
			"service_provider":   {"Auto Shop"},
		}
		req := postForm(form)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Brakes")
		assert.Contains(t, body, "50000")
		assert.Contains(t, body, "120.50")
		assert.Contains(t, body, "Auto Shop")
	})

	t.Run("returns 404 when record not found", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getErr: models.NewNotFoundError("MaintenanceRecord", "m99")}
		handler := newTestMaintenanceEditPageHandler(t, &stubVehicleSvc{}, maintenanceStub)

		req := postForm(validForm)
		req = mux.SetURLVars(req, map[string]string{"id": "m99"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceUpdate(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 403 when vehicle belongs to another user", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		handler := newTestMaintenanceEditPageHandler(t, vehicleStub, maintenanceStub)

		req := postForm(validForm)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceUpdate(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestMaintenanceEditPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := postForm(validForm)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		rec := httptest.NewRecorder()

		handler.MaintenanceUpdate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when update fails", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{
			getResult: baseRecord,
			updateErr: models.NewDatabaseError("update", assert.AnError),
		}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestMaintenanceEditPageHandler(t, vehicleStub, maintenanceStub)

		req := postForm(validForm)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceUpdate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to update maintenance record")
	})

	t.Run("accepts optional mileage and cost fields", func(t *testing.T) {
		maintenanceStub := &stubMaintenanceSvc{getResult: baseRecord, updateRes: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestMaintenanceEditPageHandler(t, vehicleStub, maintenanceStub)

		form := url.Values{
			"service_type":       {"oil_change"},
			"service_date":       {"2024-06-15"},
			"mileage_at_service": {"50000"},
			"cost":               {"79.99"},
			"service_provider":   {"Premium Auto"},
		}
		req := postForm(form)
		req = mux.SetURLVars(req, map[string]string{"id": "m1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceUpdate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
	})
}
