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

func newTestFuelEditPageHandler(
	t *testing.T,
	vehicleSvc *stubVehicleSvc,
	fuelSvc *stubFuelSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, &stubMaintenanceSvc{}, fuelSvc, nil, nil)
}

func TestPageHandler_FuelUpdate(t *testing.T) {
	fillDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	pricePerUnit := 3.499

	baseRecord := &models.FuelRecord{
		ID: "f1", VehicleID: "v1", FuelType: "gasoline",
		FillDate: fillDate, Volume: 12.5, Mileage: 45000,
		PricePerUnit: &pricePerUnit,
	}
	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}

	postForm := func(vals url.Values) *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/fuel/f1/edit",
			strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req
	}

	validForm := url.Values{
		"fuel_type": {"gasoline"},
		"fill_date": {"2024-06-15"},
		"mileage":   {"46000"},
		"volume":    {"13.00"},
	}

	t.Run("redirects to detail page on success", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getResult: baseRecord, updateRes: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestFuelEditPageHandler(t, vehicleStub, fuelStub)

		req := postForm(validForm)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelUpdate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/fuel/f1?updated=true", rec.Header().Get("Location"))
	})

	t.Run("returns 400 when fill date is invalid", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestFuelEditPageHandler(t, vehicleStub, fuelStub)

		form := url.Values{
			"fuel_type": {"gasoline"},
			"fill_date": {"not-a-date"},
			"mileage":   {"45000"},
			"volume":    {"12.50"},
		}
		req := postForm(form)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid date format")
	})

	t.Run("returns 400 when fuel type is missing", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestFuelEditPageHandler(t, vehicleStub, fuelStub)

		form := url.Values{
			"fuel_type": {""},
			"fill_date": {"2024-06-15"},
			"mileage":   {"45000"},
			"volume":    {"12.50"},
		}
		req := postForm(form)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("repopulates form fields on validation error", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestFuelEditPageHandler(t, vehicleStub, fuelStub)

		form := url.Values{
			"fuel_type":      {"diesel"},
			"fill_date":      {""},
			"mileage":        {"50000"},
			"volume":         {"14.00"},
			"price_per_unit": {"3.599"},
			"location":       {"Chevron"},
		}
		req := postForm(form)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Diesel")
		assert.Contains(t, body, "50000")
		assert.Contains(t, body, "14.00")
		assert.Contains(t, body, "3.599")
		assert.Contains(t, body, "Chevron")
	})

	t.Run("returns 404 when record not found", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getErr: models.NewNotFoundError("FuelRecord", "f99")}
		handler := newTestFuelEditPageHandler(t, &stubVehicleSvc{}, fuelStub)

		req := postForm(validForm)
		req = mux.SetURLVars(req, map[string]string{"id": "f99"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelUpdate(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 403 when vehicle belongs to another user", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		handler := newTestFuelEditPageHandler(t, vehicleStub, fuelStub)

		req := postForm(validForm)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelUpdate(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestFuelEditPageHandler(t, &stubVehicleSvc{}, &stubFuelSvc{})

		req := postForm(validForm)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		rec := httptest.NewRecorder()

		handler.FuelUpdate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when update fails", func(t *testing.T) {
		fuelStub := &stubFuelSvc{
			getResult: baseRecord,
			updateErr: models.NewDatabaseError("update fuel", assert.AnError),
		}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestFuelEditPageHandler(t, vehicleStub, fuelStub)

		req := postForm(validForm)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelUpdate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to update fuel record")
	})

	t.Run("accepts optional fields", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getResult: baseRecord, updateRes: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestFuelEditPageHandler(t, vehicleStub, fuelStub)

		form := url.Values{
			"fuel_type":               {"gasoline"},
			"fill_date":               {"2024-06-15"},
			"mileage":                 {"50000"},
			"volume":                  {"15.00"},
			"price_per_unit":          {"3.599"},
			"octane_rating":           {"91"},
			"location":                {"Premium Auto"},
			"brand":                   {"Chevron"},
			"notes":                   {"Premium fill"},
			"city_driving_percentage": {"60"},
			"vehicle_reported_mpg":    {"30.5"},
		}
		req := postForm(form)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelUpdate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
	})
}
