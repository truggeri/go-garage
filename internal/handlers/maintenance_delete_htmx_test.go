package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

func TestMaintenanceDelete_HTMXCookieCSRF(t *testing.T) {
	for _, path := range []string{"/maintenance", "/maintenance/m1"} {
		t.Run(path, func(t *testing.T) {
			vehicle := &models.Vehicle{ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020}
			record := &models.MaintenanceRecord{ID: "m1", VehicleID: "v1", ServiceType: "oil_change", ServiceDate: time.Now()}
			vehicleSvc := &stubVehicleSvc{getResult: vehicle, listResult: []*models.Vehicle{vehicle}}
			maintenanceSvc := &stubMaintenanceSvc{getResult: record, listResult: []*models.MaintenanceRecord{record}}
			page := newTestMaintenanceDetailPageHandler(t, vehicleSvc, maintenanceSvc)
			api := MakeMaintenanceAPIHandler(maintenanceSvc, vehicleSvc)
			router := mux.NewRouter()
			router.Handle("/maintenance", middleware.CSRFProtection("test-secret")(http.HandlerFunc(page.MaintenanceList))).Methods(http.MethodGet)
			router.Handle("/maintenance/{id}", middleware.CSRFProtection("test-secret")(http.HandlerFunc(page.MaintenanceDetail))).Methods(http.MethodGet)
			router.Handle("/api/v1/maintenance/{id}", middleware.APICSRFProtection("test-secret")(http.HandlerFunc(api.RemoveOne))).Methods(http.MethodDelete)
			authedRouter := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r = addAuthContext(r, "u1", "testuser")
				r = r.WithContext(context.WithValue(r.Context(), middleware.AuthMethodContextKey, middleware.AuthMethodCookie))
				router.ServeHTTP(w, r)
			})

			get := httptest.NewRecorder()
			authedRouter.ServeHTTP(get, httptest.NewRequest(http.MethodGet, path, nil))
			require.Equal(t, http.StatusOK, get.Code)
			assert.Contains(t, get.Body.String(), `hx-delete="/api/v1/maintenance/m1"`)
			token := regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)">`).FindStringSubmatch(get.Body.String())
			require.Len(t, token, 2)

			for _, csrfToken := range []string{"", "invalid", token[1]} {
				req := httptest.NewRequest(http.MethodDelete, "/api/v1/maintenance/m1", nil)
				req.Header.Set("HX-Request", "true")
				req.Header.Set("X-CSRF-Token", csrfToken)
				rec := httptest.NewRecorder()
				authedRouter.ServeHTTP(rec, req)

				if csrfToken != token[1] {
					assert.Equal(t, http.StatusForbidden, rec.Code)
					assert.Contains(t, rec.Body.String(), "CSRF_ERROR")
					assert.Empty(t, rec.Header().Get("HX-Redirect"))
					assert.Empty(t, maintenanceSvc.deletedID)
					continue
				}
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, "/vehicles/v1", rec.Header().Get("HX-Redirect"))
				assert.Equal(t, "m1", maintenanceSvc.deletedID)
				var payload map[string]interface{}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
				assert.Equal(t, true, payload["success"])
				assert.Equal(t, "Maintenance record deleted successfully", payload["message"])
			}
		})
	}
}

func TestMaintenanceDelete_HTMXErrors(t *testing.T) {
	for _, tt := range []struct {
		name       string
		userID     string
		getErr     error
		vehicleErr error
		deleteErr  error
		status     int
	}{
		{name: "unauthenticated", status: http.StatusUnauthorized},
		{name: "not owner", userID: "other-user", status: http.StatusForbidden},
		{name: "record missing", userID: "u1", getErr: models.NewNotFoundError("MaintenanceRecord", "m1"), status: http.StatusNotFound},
		{name: "vehicle missing", userID: "u1", vehicleErr: models.NewNotFoundError("Vehicle", "v1"), status: http.StatusNotFound},
		{name: "delete failed", userID: "u1", deleteErr: models.NewDatabaseError("delete maintenance", assert.AnError), status: http.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			maintenanceSvc := &stubMaintenanceSvc{
				getResult: &models.MaintenanceRecord{ID: "m1", VehicleID: "v1"},
				getErr:    tt.getErr, deleteErr: tt.deleteErr,
			}
			api := MakeMaintenanceAPIHandler(maintenanceSvc, &stubVehicleSvc{
				getResult: &models.Vehicle{ID: "v1", UserID: "u1"}, getErr: tt.vehicleErr,
			})
			req := mux.SetURLVars(httptest.NewRequest(http.MethodDelete, "/api/v1/maintenance/m1", nil), map[string]string{"id": "m1"})
			req.Header.Set("HX-Request", "true")
			if tt.userID != "" {
				req = addAuthContext(req, tt.userID, "testuser")
			}
			rec := httptest.NewRecorder()
			api.RemoveOne(rec, req)

			assert.Equal(t, tt.status, rec.Code)
			assert.Empty(t, rec.Header().Get("HX-Redirect"))
			assert.Contains(t, rec.Body.String(), `"success":false`)
			if tt.deleteErr == nil {
				assert.Empty(t, maintenanceSvc.deletedID)
			}
		})
	}
}
