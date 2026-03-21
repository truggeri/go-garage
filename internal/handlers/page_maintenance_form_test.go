package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/templateengine"
)

func newTestMaintenanceFormPageHandler(
	t *testing.T,
	vehicleSvc *stubVehicleSvc,
	maintenanceSvc *stubMaintenanceSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, maintenanceSvc, nil, nil)
}

func TestSafeIndex(t *testing.T) {
	slice := []string{"a", "b", "c"}

	t.Run("returns value at valid index", func(t *testing.T) {
		assert.Equal(t, "a", safeIndex(slice, 0))
		assert.Equal(t, "c", safeIndex(slice, 2))
	})

	t.Run("returns empty string for out of bounds index", func(t *testing.T) {
		assert.Equal(t, "", safeIndex(slice, 3))
		assert.Equal(t, "", safeIndex(slice, 100))
	})

	t.Run("returns empty string for negative index", func(t *testing.T) {
		assert.Equal(t, "", safeIndex(slice, -1))
	})

	t.Run("returns empty string for nil slice", func(t *testing.T) {
		assert.Equal(t, "", safeIndex(nil, 0))
	})
}

func TestParseMaintenanceNewForm(t *testing.T) {
	t.Run("parses valid service date", func(t *testing.T) {
		result := parseMaintenanceNewForm("2024-06-15", "", "")
		assert.Empty(t, result.Errors)
		assert.Equal(t, 2024, result.ServiceDate.Year())
		assert.Equal(t, 6, int(result.ServiceDate.Month()))
		assert.Equal(t, 15, result.ServiceDate.Day())
	})

	t.Run("returns error for invalid service date", func(t *testing.T) {
		result := parseMaintenanceNewForm("not-a-date", "", "")
		assert.Contains(t, result.Errors, "service_date")
	})

	t.Run("leaves service date zero when empty", func(t *testing.T) {
		result := parseMaintenanceNewForm("", "", "")
		assert.Empty(t, result.Errors)
		assert.True(t, result.ServiceDate.IsZero())
	})

	t.Run("parses valid mileage", func(t *testing.T) {
		result := parseMaintenanceNewForm("2024-06-15", "45000", "")
		assert.Empty(t, result.Errors)
		require.NotNil(t, result.MileageAtService)
		assert.Equal(t, 45000, *result.MileageAtService)
	})

	t.Run("returns error for invalid mileage", func(t *testing.T) {
		result := parseMaintenanceNewForm("2024-06-15", "abc", "")
		assert.Contains(t, result.Errors, "mileage_at_service")
	})

	t.Run("leaves mileage nil when empty", func(t *testing.T) {
		result := parseMaintenanceNewForm("2024-06-15", "", "")
		assert.Nil(t, result.MileageAtService)
	})

	t.Run("parses valid cost", func(t *testing.T) {
		result := parseMaintenanceNewForm("2024-06-15", "", "49.99")
		assert.Empty(t, result.Errors)
		require.NotNil(t, result.Cost)
		assert.Equal(t, 49.99, *result.Cost)
	})

	t.Run("returns error for invalid cost", func(t *testing.T) {
		result := parseMaintenanceNewForm("2024-06-15", "", "abc")
		assert.Contains(t, result.Errors, "cost")
	})

	t.Run("leaves cost nil when empty", func(t *testing.T) {
		result := parseMaintenanceNewForm("2024-06-15", "", "")
		assert.Nil(t, result.Cost)
	})
}

func TestPageHandler_MaintenanceNew(t *testing.T) {
	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}

	t.Run("renders form for authenticated user with vehicles", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/maintenance/new", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceNew(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Add Maintenance Record")
		assert.Contains(t, body, "Ford Focus")
	})

	t.Run("pre-selects vehicle from query param", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/maintenance/new?vehicle=v1", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceNew(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, `selected`)
	})

	t.Run("pre-selects vehicle from vehicle_id fallback param", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/maintenance/new?vehicle_id=v1", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceNew(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, `selected`)
	})

	t.Run("renders no-vehicles empty state", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{}}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/maintenance/new", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceNew(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "No Vehicles Found")
		assert.Contains(t, body, "Add a Vehicle")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestMaintenanceFormPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/maintenance/new", nil)
		rec := httptest.NewRecorder()

		handler.MaintenanceNew(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when vehicle service fails", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listErr: models.NewDatabaseError("get vehicles", assert.AnError)}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/maintenance/new", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceNew(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPageHandler_MaintenanceCreate(t *testing.T) {
	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}

	postForm := func(vals url.Values) *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/maintenance/new",
			strings.NewReader(vals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req
	}

	validForm := url.Values{
		"vehicle_id":   {"v1"},
		"service_type": {"Oil Change"},
		"service_date": {"2024-01-15"},
	}

	t.Run("redirects on successful creation", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := postForm(validForm)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceCreate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/maintenance?added=true", rec.Header().Get("Location"))
	})

	t.Run("returns 400 when vehicle not owned by user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		form := url.Values{
			"vehicle_id":   {"other-vehicle"},
			"service_type": {"Oil Change"},
			"service_date": {"2024-01-15"},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 when service date is invalid", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		form := url.Values{
			"vehicle_id":   {"v1"},
			"service_type": {"Oil Change"},
			"service_date": {"not-a-date"},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid date format")
	})

	t.Run("returns 400 when service type is missing", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		form := url.Values{
			"vehicle_id":   {"v1"},
			"service_type": {""},
			"service_date": {"2024-01-15"},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 500 when create fails", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		maintenanceStub := &stubMaintenanceSvc{createErr: models.NewDatabaseError("create", assert.AnError)}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, maintenanceStub)

		req := postForm(validForm)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceCreate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Failed to add record 1")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestMaintenanceFormPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := postForm(validForm)
		rec := httptest.NewRecorder()

		handler.MaintenanceCreate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when vehicle service fails", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listErr: models.NewDatabaseError("get vehicles", assert.AnError)}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := postForm(validForm)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceCreate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("accepts optional mileage and cost fields", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		form := url.Values{
			"vehicle_id":         {"v1"},
			"service_type":       {"Oil Change"},
			"service_date":       {"2024-01-15"},
			"mileage_at_service": {"45000"},
			"cost":               {"49.99"},
			"service_provider":   {"Quick Lube"},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceCreate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
	})

	t.Run("bulk creates multiple records for one vehicle", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		maintenanceStub := &stubMaintenanceSvc{}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, maintenanceStub)

		form := url.Values{
			"vehicle_id":         {"v1"},
			"service_type":       {"Oil Change", "Tire Rotation"},
			"service_date":       {"2024-01-15", "2024-02-20"},
			"mileage_at_service": {"45000", "46000"},
			"cost":               {"49.99", "80.00"},
			"service_provider":   {"Quick Lube", "Tire Shop"},
			"notes":              {"First record", "Second record"},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceCreate(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/maintenance?added=true", rec.Header().Get("Location"))
	})

	t.Run("bulk returns 400 when one record has validation error", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		form := url.Values{
			"vehicle_id":   {"v1"},
			"service_type": {"Oil Change", ""},
			"service_date": {"2024-01-15", "2024-02-20"},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "service type is required")
	})

	t.Run("bulk returns 400 when one record has invalid date", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		form := url.Values{
			"vehicle_id":   {"v1"},
			"service_type": {"Oil Change", "Tire Rotation"},
			"service_date": {"2024-01-15", "not-a-date"},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Invalid date format")
	})

	t.Run("bulk preserves valid record values on error re-render", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestMaintenanceFormPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		form := url.Values{
			"vehicle_id":   {"v1"},
			"service_type": {"Oil Change", ""},
			"service_date": {"2024-01-15", "2024-02-20"},
			"cost":         {"49.99", ""},
		}
		req := postForm(form)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.MaintenanceCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Oil Change")
		assert.Contains(t, body, "Record #1")
		assert.Contains(t, body, "Record #2")
	})
}
