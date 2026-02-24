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

func newTestVehicleListPageHandler(
	t *testing.T,
	vehicleSvc *stubVehicleSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, &stubMaintenanceSvc{})
}

func newTestVehicleDetailPageHandler(
	t *testing.T,
	vehicleSvc *stubVehicleSvc,
	maintenanceSvc *stubMaintenanceSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, maintenanceSvc)
}

func TestPageHandler_VehicleList(t *testing.T) {
	mileage := 45000

	t.Run("renders vehicle list for authenticated user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			countResult: 2,
			listResult: []*models.Vehicle{
				{ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive, CurrentMileage: &mileage},
				{ID: "v2", UserID: "u1", Make: "Toyota", Model: "Camry", Year: 2022, Status: models.VehicleStatusActive},
			},
		}
		handler := newTestVehicleListPageHandler(t, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/vehicles", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "My Vehicles")
		assert.Contains(t, body, "Ford")
		assert.Contains(t, body, "Toyota")
		assert.Contains(t, body, "45,000 mi")
	})

	t.Run("renders empty state when no vehicles", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			countResult: 0,
			listResult:  []*models.Vehicle{},
		}
		handler := newTestVehicleListPageHandler(t, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/vehicles", nil)
		req = addAuthContext(req, "u1", "newuser")
		rec := httptest.NewRecorder()

		handler.VehicleList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "No vehicles found")
		assert.Contains(t, body, "Add Your First Vehicle")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestVehicleListPageHandler(t, &stubVehicleSvc{})

		req := httptest.NewRequest(http.MethodGet, "/vehicles", nil)
		rec := httptest.NewRecorder()

		handler.VehicleList(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when count fails", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			countErr: models.NewDatabaseError("count vehicles", assert.AnError),
		}
		handler := newTestVehicleListPageHandler(t, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/vehicles", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleList(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when list fails", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			countResult: 1,
			listErr:     models.NewDatabaseError("list vehicles", assert.AnError),
		}
		handler := newTestVehicleListPageHandler(t, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/vehicles", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleList(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("passes filter parameters", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			countResult: 1,
			listResult: []*models.Vehicle{
				{ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
			},
		}
		handler := newTestVehicleListPageHandler(t, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/vehicles?make=Ford&status=active", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Ford")
		assert.Contains(t, body, "Clear")
	})

	t.Run("handles pagination parameter", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			countResult: 25,
			listResult:  []*models.Vehicle{},
		}
		handler := newTestVehicleListPageHandler(t, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/vehicles?page=2", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("shows success flash when added=true", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			countResult: 1,
			listResult: []*models.Vehicle{
				{ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
			},
		}
		handler := newTestVehicleListPageHandler(t, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/vehicles?added=true", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Vehicle added successfully")
	})
}

func TestPageHandler_VehicleNew(t *testing.T) {
	t.Run("renders add vehicle form for authenticated user", func(t *testing.T) {
		handler := newTestVehicleListPageHandler(t, &stubVehicleSvc{})

		req := httptest.NewRequest(http.MethodGet, "/vehicles/new", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleNew(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Add Vehicle")
		assert.Contains(t, body, `action="/vehicles/new"`)
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestVehicleListPageHandler(t, &stubVehicleSvc{})

		req := httptest.NewRequest(http.MethodGet, "/vehicles/new", nil)
		rec := httptest.NewRecorder()

		handler.VehicleNew(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPageHandler_VehicleCreate(t *testing.T) {
	validForm := func() url.Values {
		form := url.Values{}
		form.Set("make", "Toyota")
		form.Set("model", "Camry")
		form.Set("year", "2021")
		return form
	}

	t.Run("redirects to vehicles list on success", func(t *testing.T) {
		handler := newTestVehicleListPageHandler(t, &stubVehicleSvc{})

		req := httptest.NewRequest(http.MethodPost, "/vehicles/new", strings.NewReader(validForm().Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleCreate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/vehicles?added=true", rec.Header().Get("Location"))
	})

	t.Run("returns 400 when required fields missing", func(t *testing.T) {
		handler := newTestVehicleListPageHandler(t, &stubVehicleSvc{})

		form := url.Values{}
		req := httptest.NewRequest(http.MethodPost, "/vehicles/new", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Make is required")
		assert.Contains(t, body, "Model is required")
		assert.Contains(t, body, "Year is required")
	})

	t.Run("returns 400 for invalid year", func(t *testing.T) {
		handler := newTestVehicleListPageHandler(t, &stubVehicleSvc{})

		form := url.Values{}
		form.Set("make", "Toyota")
		form.Set("model", "Camry")
		form.Set("year", "abc")
		req := httptest.NewRequest(http.MethodPost, "/vehicles/new", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Year must be a valid year")
	})

	t.Run("repopulates form fields on validation error", func(t *testing.T) {
		handler := newTestVehicleListPageHandler(t, &stubVehicleSvc{})

		form := url.Values{}
		form.Set("make", "Honda")
		form.Set("model", "")
		form.Set("year", "2022")
		req := httptest.NewRequest(http.MethodPost, "/vehicles/new", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Honda")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestVehicleListPageHandler(t, &stubVehicleSvc{})

		req := httptest.NewRequest(http.MethodPost, "/vehicles/new", strings.NewReader(validForm().Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.VehicleCreate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when service fails", func(t *testing.T) {
		stub := &stubVehicleSvc{
			createErr: models.NewDatabaseError("create vehicle", assert.AnError),
		}
		handler := newTestVehicleListPageHandler(t, stub)

		req := httptest.NewRequest(http.MethodPost, "/vehicles/new", strings.NewReader(validForm().Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleCreate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("accepts optional fields", func(t *testing.T) {
		handler := newTestVehicleListPageHandler(t, &stubVehicleSvc{})

		form := validForm()
		form.Set("vin", "1HGBH41JXMN109186")
		form.Set("color", "Blue")
		form.Set("license_plate", "ABC-1234")
		form.Set("purchase_date", "2021-06-15")
		form.Set("purchase_price", "25000.00")
		form.Set("purchase_mileage", "5")
		form.Set("current_mileage", "12000")
		form.Set("notes", "Great car")

		req := httptest.NewRequest(http.MethodPost, "/vehicles/new", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleCreate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
	})
}

func TestPageHandler_VehicleDetail(t *testing.T) {
	cost := 120.50
	serviceDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	mileage := 45000

	t.Run("renders vehicle detail page for authenticated user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{
				ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
				Status: models.VehicleStatusActive, CurrentMileage: &mileage,
			},
		}
		maintenanceStub := &stubMaintenanceSvc{
			listResult: []*models.MaintenanceRecord{
				{ID: "m1", VehicleID: "v1", ServiceType: "Oil Change", ServiceDate: serviceDate, Cost: &cost},
			},
		}
		handler := newTestVehicleDetailPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleDetail(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Ford")
		assert.Contains(t, body, "Focus")
		assert.Contains(t, body, "2020")
		assert.Contains(t, body, "Oil Change")
	})

	t.Run("shows update flash when updated=true", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{
				ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
				Status: models.VehicleStatusActive,
			},
		}
		handler := newTestVehicleDetailPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1?updated=true", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleDetail(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Vehicle updated successfully")
	})

	t.Run("returns 404 when vehicle not found", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getErr: models.NewNotFoundError("Vehicle", "v99"),
		}
		handler := newTestVehicleDetailPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v99", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "v99"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleDetail(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 403 when vehicle belongs to another user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", Make: "Toyota", Model: "Camry", Year: 2021, Status: models.VehicleStatusActive},
		}
		handler := newTestVehicleDetailPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleDetail(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestVehicleDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		handler.VehicleDetail(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func newTestVehicleEditPageHandler(
	t *testing.T,
	vehicleSvc *stubVehicleSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, &stubMaintenanceSvc{})
}

func TestPageHandler_VehicleEdit(t *testing.T) {
	mileage := 45000
	price := 25000.0
	purchaseMileage := 5000
	purchaseDate := time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC)

	t.Run("renders pre-populated edit form for authenticated user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{
				ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
				Status: models.VehicleStatusActive, CurrentMileage: &mileage,
				VIN: "1HGBH41JXMN109186", Color: "Blue", LicensePlate: "ABC-1234",
				PurchaseDate: &purchaseDate, PurchasePrice: &price, PurchaseMileage: &purchaseMileage,
				Notes: "Great car",
			},
		}
		handler := newTestVehicleEditPageHandler(t, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleEdit(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Edit Vehicle")
		assert.Contains(t, body, `value="Ford"`)
		assert.Contains(t, body, `value="Focus"`)
		assert.Contains(t, body, `value="2020"`)
		assert.Contains(t, body, "1HGBH41JXMN109186")
		assert.Contains(t, body, "Great car")
		assert.Contains(t, body, `action="/vehicles/v1/edit"`)
	})

	t.Run("returns 404 when vehicle not found", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getErr: models.NewNotFoundError("Vehicle", "v99"),
		}
		handler := newTestVehicleEditPageHandler(t, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v99/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "v99"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleEdit(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 403 when vehicle belongs to another user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", Make: "Toyota", Model: "Camry", Year: 2021, Status: models.VehicleStatusActive},
		}
		handler := newTestVehicleEditPageHandler(t, vehicleStub)

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleEdit(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestVehicleEditPageHandler(t, &stubVehicleSvc{})

		req := httptest.NewRequest(http.MethodGet, "/vehicles/v1/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		handler.VehicleEdit(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPageHandler_VehicleUpdate(t *testing.T) {
	validVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}

	validForm := func() url.Values {
		form := url.Values{}
		form.Set("make", "Toyota")
		form.Set("model", "Camry")
		form.Set("year", "2021")
		form.Set("status", "active")
		return form
	}

	t.Run("redirects to vehicle detail on success", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getResult: validVehicle,
			updateRes: &models.Vehicle{ID: "v1", UserID: "u1", Make: "Toyota", Model: "Camry", Year: 2021, Status: models.VehicleStatusActive},
		}
		handler := newTestVehicleEditPageHandler(t, stub)

		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(validForm().Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/vehicles/v1?updated=true", rec.Header().Get("Location"))
	})

	t.Run("returns 400 when required fields missing", func(t *testing.T) {
		stub := &stubVehicleSvc{getResult: validVehicle}
		handler := newTestVehicleEditPageHandler(t, stub)

		form := url.Values{}
		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Make is required")
		assert.Contains(t, body, "Model is required")
		assert.Contains(t, body, "Year is required")
	})

	t.Run("repopulates form fields on validation error", func(t *testing.T) {
		stub := &stubVehicleSvc{getResult: validVehicle}
		handler := newTestVehicleEditPageHandler(t, stub)

		form := url.Values{}
		form.Set("make", "Honda")
		form.Set("model", "")
		form.Set("year", "2022")
		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Honda")
	})

	t.Run("returns 404 when vehicle not found", func(t *testing.T) {
		stub := &stubVehicleSvc{getErr: models.NewNotFoundError("Vehicle", "v99")}
		handler := newTestVehicleEditPageHandler(t, stub)

		req := httptest.NewRequest(http.MethodPost, "/vehicles/v99/edit", strings.NewReader(validForm().Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"id": "v99"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 403 when vehicle belongs to another user", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		handler := newTestVehicleEditPageHandler(t, stub)

		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(validForm().Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestVehicleEditPageHandler(t, &stubVehicleSvc{})

		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(validForm().Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when update service fails", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getResult: validVehicle,
			updateErr: models.NewDatabaseError("update vehicle", assert.AnError),
		}
		handler := newTestVehicleEditPageHandler(t, stub)

		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(validForm().Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("accepts all optional fields", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getResult: validVehicle,
			updateRes: &models.Vehicle{ID: "v1", UserID: "u1", Make: "Toyota", Model: "Camry", Year: 2021, Status: models.VehicleStatusActive},
		}
		handler := newTestVehicleEditPageHandler(t, stub)

		form := validForm()
		form.Set("vin", "1HGBH41JXMN109186")
		form.Set("color", "Red")
		form.Set("license_plate", "XYZ-9999")
		form.Set("purchase_date", "2021-06-15")
		form.Set("purchase_price", "30000.00")
		form.Set("purchase_mileage", "100")
		form.Set("current_mileage", "15000")
		form.Set("notes", "Updated notes")
		req := httptest.NewRequest(http.MethodPost, "/vehicles/v1/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleUpdate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
	})
}
