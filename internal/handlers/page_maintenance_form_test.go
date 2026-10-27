package handlers

import (
	"net/http"
	"net/http/httptest"
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
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, maintenanceSvc, nil, nil, nil)
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
