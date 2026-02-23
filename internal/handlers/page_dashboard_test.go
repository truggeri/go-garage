package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/templateengine"
)

// newTestDashboardPageHandler creates a PageHandler wired with the given services.
func newTestDashboardPageHandler(
	t *testing.T,
	vehicleSvc *stubVehicleSvc,
	maintenanceSvc *stubMaintenanceSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, maintenanceSvc)
}

func TestPageHandler_Dashboard(t *testing.T) {
	serviceDate := time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC)
	cost := 89.99

	t.Run("renders dashboard for authenticated user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{
			listResult: []*models.Vehicle{
				{ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
			},
		}
		maintenanceStub := &stubMaintenanceSvc{
			listResult: []*models.MaintenanceRecord{
				{ID: "m1", VehicleID: "v1", ServiceType: "Oil Change", ServiceDate: serviceDate, Cost: &cost},
			},
		}
		handler := newTestDashboardPageHandler(t, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.Dashboard(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Welcome back, testuser")
		assert.Contains(t, body, "Ford")
		assert.Contains(t, body, "Oil Change")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestDashboardPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		rec := httptest.NewRecorder()

		handler.Dashboard(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("renders empty state when no vehicles", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{}}
		handler := newTestDashboardPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req = addAuthContext(req, "u1", "newuser")
		rec := httptest.NewRecorder()

		handler.Dashboard(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Welcome back, newuser")
		assert.Contains(t, body, "No maintenance records yet")
	})

	t.Run("returns 500 when vehicle service fails", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listErr: models.NewDatabaseError("list vehicles", assert.AnError)}
		handler := newTestDashboardPageHandler(t, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.Dashboard(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
