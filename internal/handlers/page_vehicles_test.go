package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
}
