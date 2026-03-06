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

// newTestProfilePageHandler creates a PageHandler wired with the given services for profile tests.
func newTestProfilePageHandler(
	t *testing.T,
	userSvc *stubUserSvc,
	vehicleSvc *stubVehicleSvc,
	maintenanceSvc *stubMaintenanceSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, maintenanceSvc, userSvc)
}

func TestPageHandler_ViewProfile(t *testing.T) {
	now := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("renders profile for authenticated user", func(t *testing.T) {
		userStub := &stubUserSvc{
			getResult: &models.User{
				ID:        "u1",
				Username:  "johndoe",
				Email:     "john@example.com",
				FirstName: "John",
				LastName:  "Doe",
				CreatedAt: now,
				UpdatedAt: now,
			},
		}
		vehicleStub := &stubVehicleSvc{
			listResult: []*models.Vehicle{
				{ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
				{ID: "v2", UserID: "u1", Make: "Toyota", Model: "Camry", Year: 2022, Status: models.VehicleStatusActive},
			},
		}
		maintenanceStub := &stubMaintenanceSvc{
			listResult: []*models.MaintenanceRecord{
				{ID: "m1", VehicleID: "v1", ServiceType: models.ServiceTypeOilChange, ServiceDate: now},
			},
		}
		handler := newTestProfilePageHandler(t, userStub, vehicleStub, maintenanceStub)

		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req = addAuthContext(req, "u1", "johndoe")
		rec := httptest.NewRecorder()

		handler.ViewProfile(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "johndoe")
		assert.Contains(t, body, "john@example.com")
		assert.Contains(t, body, "John")
		assert.Contains(t, body, "Doe")
		assert.Contains(t, body, "Jan 15, 2024")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestProfilePageHandler(t, &stubUserSvc{}, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		rec := httptest.NewRecorder()

		handler.ViewProfile(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when user service fails", func(t *testing.T) {
		userStub := &stubUserSvc{getErr: models.NewDatabaseError("get user", assert.AnError)}
		handler := newTestProfilePageHandler(t, userStub, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.ViewProfile(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 500 when vehicle service fails", func(t *testing.T) {
		userStub := &stubUserSvc{
			getResult: &models.User{
				ID:        "u1",
				Username:  "testuser",
				Email:     "test@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
		}
		vehicleStub := &stubVehicleSvc{listErr: models.NewDatabaseError("list vehicles", assert.AnError)}
		handler := newTestProfilePageHandler(t, userStub, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.ViewProfile(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("renders profile with zero vehicles", func(t *testing.T) {
		userStub := &stubUserSvc{
			getResult: &models.User{
				ID:        "u1",
				Username:  "newuser",
				Email:     "new@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
		}
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{}}
		handler := newTestProfilePageHandler(t, userStub, vehicleStub, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req = addAuthContext(req, "u1", "newuser")
		rec := httptest.NewRecorder()

		handler.ViewProfile(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "newuser")
		assert.Contains(t, body, "new@example.com")
	})
}
