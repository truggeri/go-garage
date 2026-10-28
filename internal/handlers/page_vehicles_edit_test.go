package handlers

import (
	"net/http"
	"net/http/httptest"
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

		serveHTMXPage(t, handler.VehicleEdit, rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Edit Vehicle")
		assert.Contains(t, body, `hx-put="/api/v1/vehicles/v1"`)
		assert.Contains(t, body, `hx-ext="json-enc"`)
		assert.Contains(t, body, `hx-swap="none"`)
		assert.Contains(t, body, `data-form-errors`)
		for _, field := range []string{"vin", "make", "model", "year", "display_name", "color", "license_plate", "purchase_date", "purchase_price", "purchase_mileage", "current_mileage", "notes"} {
			assert.Contains(t, body, `data-field="`+field+`"`)
		}
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
