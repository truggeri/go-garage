package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/truggeri/go-garage/internal/models"
)

func TestPageHandler_VehicleEdit(t *testing.T) {
	mileage := 45000
	price := 25000.00
	purchaseDate := time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC)
	purchaseMileage := 5

	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive, CurrentMileage: &mileage,
		PurchaseDate: &purchaseDate, PurchasePrice: &price,
		PurchaseMileage: &purchaseMileage, VIN: "1HGBH41JXMN109186",
		Color: "Blue", LicensePlate: "ABC-1234", Notes: "Great car",
	}

	t.Run("renders edit form with pre-populated data", func(t *testing.T) {
		handler := newTestVehicleDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1/edit", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = addResourceContext(req, baseVehicle)
		rec := httptest.NewRecorder()

		handler.VehicleEdit(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Edit Vehicle")
		assert.Contains(t, body, `action="/vehicles/v1/edit"`)
		assert.Contains(t, body, "Ford")
		assert.Contains(t, body, "Focus")
		assert.Contains(t, body, "2020")
		assert.Contains(t, body, "1HGBH41JXMN109186")
		assert.Contains(t, body, "Blue")
		assert.Contains(t, body, "ABC-1234")
		assert.Contains(t, body, "45000")
		assert.Contains(t, body, "2021-06-15")
		assert.Contains(t, body, "25000.00")
		assert.Contains(t, body, "Great car")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestVehicleDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1/edit", nil)
		req = addResourceContext(req, baseVehicle)
		rec := httptest.NewRecorder()

		handler.VehicleEdit(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when resource missing from context", func(t *testing.T) {
		handler := newTestVehicleDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1/edit", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleEdit(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPageHandler_VehicleUpdate(t *testing.T) {
	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive, VIN: "1HGBH41JXMN109186",
	}

	validForm := func() url.Values {
		form := url.Values{}
		form.Set("make", "Toyota")
		form.Set("model", "Camry")
		form.Set("year", "2021")
		form.Set("vin", "1HGBH41JXMN109186")
		return form
	}

	t.Run("redirects to vehicle detail on success", func(t *testing.T) {
		handler := newTestVehicleDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(validForm().Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		req = addResourceContext(req, baseVehicle)
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/vehicles/v1?updated=true", rec.Header().Get("Location"))
	})

	t.Run("returns 400 when required fields missing", func(t *testing.T) {
		handler := newTestVehicleDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		form := url.Values{}
		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		req = addResourceContext(req, baseVehicle)
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "make is required")
		assert.Contains(t, body, "model is required")
		assert.Contains(t, body, "year is required")
	})

	t.Run("returns 400 for invalid year", func(t *testing.T) {
		handler := newTestVehicleDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		form := validForm()
		form.Set("year", "abc")
		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		req = addResourceContext(req, baseVehicle)
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Year must be a valid")
	})

	t.Run("repopulates form fields on validation error", func(t *testing.T) {
		handler := newTestVehicleDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		form := validForm()
		form.Set("make", "Honda")
		form.Set("model", "")
		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		req = addResourceContext(req, baseVehicle)
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Honda")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestVehicleDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(validForm().Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addResourceContext(req, baseVehicle)
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when resource missing from context", func(t *testing.T) {
		handler := newTestVehicleDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(validForm().Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when update service fails", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			updateErr: models.NewDatabaseError("update vehicle", assert.AnError),
		}
		handler := newTestVehicleDetailPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(validForm().Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		req = addResourceContext(req, baseVehicle)
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("accepts optional fields", func(t *testing.T) {
		handler := newTestVehicleDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		form := validForm()
		form.Set("color", "Red")
		form.Set("license_plate", "XYZ-5678")
		form.Set("purchase_date", "2021-06-15")
		form.Set("purchase_price", "25000.00")
		form.Set("purchase_mileage", "5")
		form.Set("current_mileage", "12000")
		form.Set("notes", "Updated notes")

		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		req = addResourceContext(req, baseVehicle)
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
	})
}
