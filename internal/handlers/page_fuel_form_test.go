package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/templateengine"
)

func newTestFuelFormPageHandler(
	t *testing.T,
	vehicleSvc *stubVehicleSvc,
	fuelSvc *stubFuelSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, &stubMaintenanceSvc{}, fuelSvc, nil, nil)
}

func TestPageHandler_FuelCreate(t *testing.T) {
	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}

	postForm := func(vals url.Values) *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/fuel/new",
			strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req
	}

	validForm := url.Values{
		"vehicle_id": {"v1"},
		"fuel_type":  {"gasoline"},
		"fill_date":  {"2024-06-15"},
		"mileage":    {"45000"},
		"volume":     {"12.50"},
	}

	t.Run("redirects on successful creation", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestFuelFormPageHandler(t, vehicleStub, &stubFuelSvc{})

		req := postForm(validForm)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelCreate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/fuel?added=true", rec.Header().Get("Location"))
	})

	t.Run("returns 400 when vehicle not owned by user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestFuelFormPageHandler(t, vehicleStub, &stubFuelSvc{})

		form := url.Values{
			"vehicle_id": {"other-vehicle"},
			"fuel_type":  {"gasoline"},
			"fill_date":  {"2024-06-15"},
			"mileage":    {"45000"},
			"volume":     {"12.50"},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 when fill date is invalid", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestFuelFormPageHandler(t, vehicleStub, &stubFuelSvc{})

		form := url.Values{
			"vehicle_id": {"v1"},
			"fuel_type":  {"gasoline"},
			"fill_date":  {"not-a-date"},
			"mileage":    {"45000"},
			"volume":     {"12.50"},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid date format")
	})

	t.Run("returns 400 when fuel type is missing", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestFuelFormPageHandler(t, vehicleStub, &stubFuelSvc{})

		form := url.Values{
			"vehicle_id": {"v1"},
			"fuel_type":  {""},
			"fill_date":  {"2024-06-15"},
			"mileage":    {"45000"},
			"volume":     {"12.50"},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 when volume is missing", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestFuelFormPageHandler(t, vehicleStub, &stubFuelSvc{})

		form := url.Values{
			"vehicle_id": {"v1"},
			"fuel_type":  {"gasoline"},
			"fill_date":  {"2024-06-15"},
			"mileage":    {"45000"},
			"volume":     {""},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 500 when create fails", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		fuelStub := &stubFuelSvc{createErr: models.NewDatabaseError("create fuel", assert.AnError)}
		handler := newTestFuelFormPageHandler(t, vehicleStub, fuelStub)

		req := postForm(validForm)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelCreate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to add fuel record")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestFuelFormPageHandler(t, &stubVehicleSvc{}, &stubFuelSvc{})

		req := postForm(validForm)
		rec := httptest.NewRecorder()

		handler.FuelCreate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when vehicle service fails", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listErr: models.NewDatabaseError("get vehicles", assert.AnError)}
		handler := newTestFuelFormPageHandler(t, vehicleStub, &stubFuelSvc{})

		req := postForm(validForm)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelCreate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("accepts optional fields", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestFuelFormPageHandler(t, vehicleStub, &stubFuelSvc{})

		form := url.Values{
			"vehicle_id":              {"v1"},
			"fuel_type":               {"gasoline"},
			"fill_date":               {"2024-06-15"},
			"mileage":                 {"45000"},
			"volume":                  {"12.50"},
			"price_per_unit":          {"3.499"},
			"octane_rating":           {"87"},
			"location":                {"Shell Station"},
			"brand":                   {"Shell"},
			"notes":                   {"Regular fill-up"},
			"city_driving_percentage": {"50"},
			"vehicle_reported_mpg":    {"28.5"},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelCreate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
	})

	t.Run("returns 400 when mileage is invalid", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestFuelFormPageHandler(t, vehicleStub, &stubFuelSvc{})

		form := url.Values{
			"vehicle_id": {"v1"},
			"fuel_type":  {"gasoline"},
			"fill_date":  {"2024-06-15"},
			"mileage":    {"abc"},
			"volume":     {"12.50"},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Mileage must be a valid number")
	})
}
