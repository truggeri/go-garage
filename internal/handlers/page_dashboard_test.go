package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/auth"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/templateengine"
)

// newTestDashboardPageHandler creates a PageHandler wired with the given services and a real TokenManager.
func newTestDashboardPageHandler(
	t *testing.T,
	tokenMgr *auth.TokenManager,
	vehicleSvc *stubVehicleSvc,
	maintenanceSvc *stubMaintenanceSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, tokenMgr, vehicleSvc, maintenanceSvc)
}

// buildTestTokenManager creates a TokenManager suitable for unit tests.
func buildTestTokenManager(t *testing.T) *auth.TokenManager {
	t.Helper()
	mgr, err := auth.BuildTokenManager("test-secret-key-for-unit-tests", auth.StandardTokenDurations())
	require.NoError(t, err)
	return mgr
}

// generateTestAccessToken creates a signed access token for a test user.
func generateTestAccessToken(t *testing.T, mgr *auth.TokenManager, userID, userName string) string {
	t.Helper()
	bundle, err := mgr.GenerateTokenBundle(auth.TokenPayload{AccountID: userID, AccountName: userName})
	require.NoError(t, err)
	return bundle.AccessToken
}

func TestPageHandler_Dashboard(t *testing.T) {
	serviceDate := time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC)
	cost := 89.99

	t.Run("renders dashboard for authenticated user", func(t *testing.T) {
		tokenMgr := buildTestTokenManager(t)
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
		handler := newTestDashboardPageHandler(t, tokenMgr, vehicleStub, maintenanceStub)

		token := generateTestAccessToken(t, tokenMgr, "u1", "testuser")
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
		rec := httptest.NewRecorder()

		handler.Dashboard(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Welcome back, testuser")
		assert.Contains(t, body, "Ford")
		assert.Contains(t, body, "Oil Change")
	})

	t.Run("redirects unauthenticated user to login", func(t *testing.T) {
		tokenMgr := buildTestTokenManager(t)
		handler := newTestDashboardPageHandler(t, tokenMgr, &stubVehicleSvc{}, &stubMaintenanceSvc{})

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		rec := httptest.NewRecorder()

		handler.Dashboard(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("Location"))
	})

	t.Run("renders empty state when no vehicles", func(t *testing.T) {
		tokenMgr := buildTestTokenManager(t)
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{}}
		maintenanceStub := &stubMaintenanceSvc{}
		handler := newTestDashboardPageHandler(t, tokenMgr, vehicleStub, maintenanceStub)

		token := generateTestAccessToken(t, tokenMgr, "u1", "newuser")
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
		rec := httptest.NewRecorder()

		handler.Dashboard(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Welcome back, newuser")
		assert.Contains(t, body, "No maintenance records yet")
	})

	t.Run("returns 500 when vehicle service fails", func(t *testing.T) {
		tokenMgr := buildTestTokenManager(t)
		vehicleStub := &stubVehicleSvc{listErr: models.NewDatabaseError("list vehicles", assert.AnError)}
		handler := newTestDashboardPageHandler(t, tokenMgr, vehicleStub, &stubMaintenanceSvc{})

		token := generateTestAccessToken(t, tokenMgr, "u1", "testuser")
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
		rec := httptest.NewRecorder()

		handler.Dashboard(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
