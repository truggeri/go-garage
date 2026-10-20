package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

func TestMaintenanceCreate_HTMX(t *testing.T) {
	for _, tt := range []struct {
		name  string
		body  string
		field string
	}{
		{"numeric inputs", `{"service_type":"other","custom_service_type":"Alignment","service_date":"2024-01-15","mileage_at_service":"45000","cost":"49.99","service_provider":"Garage","notes":"Checked"}`, ""},
		{"blank optional numbers", `{"service_type":"oil_change","service_date":"2024-01-15","mileage_at_service":"","cost":""}`, ""},
		{"missing service type", `{"service_date":"2024-01-15"}`, "service_type"},
		{"invalid date", `{"service_type":"oil_change","service_date":"invalid"}`, "service_date"},
		{"custom service required", `{"service_type":"other","service_date":"2024-01-15"}`, "custom_service_type"},
		{"invalid mileage", `{"service_type":"oil_change","service_date":"2024-01-15","mileage_at_service":"abc"}`, "mileage_at_service"},
		{"fractional mileage", `{"service_type":"oil_change","service_date":"2024-01-15","mileage_at_service":"1.5"}`, "mileage_at_service"},
		{"overflow mileage", `{"service_type":"oil_change","service_date":"2024-01-15","mileage_at_service":"9999999999999999999"}`, "mileage_at_service"},
		{"negative mileage", `{"service_type":"oil_change","service_date":"2024-01-15","mileage_at_service":"-1"}`, "mileage_at_service"},
		{"invalid cost", `{"service_type":"oil_change","service_date":"2024-01-15","cost":"abc"}`, "cost"},
		{"nonfinite cost", `{"service_type":"oil_change","service_date":"2024-01-15","cost":"NaN"}`, "cost"},
		{"infinite cost", `{"service_type":"oil_change","service_date":"2024-01-15","cost":"Infinity"}`, "cost"},
		{"negative cost", `{"service_type":"oil_change","service_date":"2024-01-15","cost":"-1"}`, "cost"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := MakeMaintenanceAPIHandler(&stubMaintenanceSvc{}, &stubVehicleSvc{
				getResult: &models.Vehicle{ID: "v2", UserID: "u1"},
			})
			req := mux.SetURLVars(httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/v2/maintenance", strings.NewReader(tt.body)), map[string]string{"vehicleId": "v2"})
			req = addAuthContext(req, "u1", "testuser")
			req.Header.Set("HX-Request", "true")
			rec := httptest.NewRecorder()
			h.CreateOne(rec, req)

			if tt.field != "" {
				require.Equal(t, http.StatusBadRequest, rec.Code)
				assert.Empty(t, rec.Header().Get("HX-Redirect"))
				assert.Contains(t, rec.Body.String(), `"field":"`+tt.field+`"`)
				return
			}
			require.Equal(t, http.StatusCreated, rec.Code)
			assert.Equal(t, "/maintenance/generated-id?added=true", rec.Header().Get("HX-Redirect"))
			var payload struct {
				Success bool `json:"success"`
				Data    struct {
					VehicleID         string   `json:"vehicle_id"`
					Cost              *float64 `json:"cost"`
					MileageAtService  *int     `json:"mileage_at_service"`
					CustomServiceType string   `json:"custom_service_type"`
					ServiceProvider   string   `json:"service_provider"`
					Notes             string   `json:"notes"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
			assert.True(t, payload.Success)
			assert.Equal(t, "v2", payload.Data.VehicleID)
			if tt.name == "numeric inputs" {
				require.NotNil(t, payload.Data.Cost)
				require.NotNil(t, payload.Data.MileageAtService)
				assert.Equal(t, 49.99, *payload.Data.Cost)
				assert.Equal(t, 45000, *payload.Data.MileageAtService)
				assert.Equal(t, "Alignment", payload.Data.CustomServiceType)
				assert.Equal(t, "Garage", payload.Data.ServiceProvider)
				assert.Equal(t, "Checked", payload.Data.Notes)
			} else {
				assert.Nil(t, payload.Data.Cost)
				assert.Nil(t, payload.Data.MileageAtService)
			}
		})
	}
}

func TestMaintenanceCreate_HTMXCookieCSRF(t *testing.T) {
	vehicle := &models.Vehicle{ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020}
	vehicleSvc := &stubVehicleSvc{getResult: vehicle, listResult: []*models.Vehicle{vehicle}}
	page := newTestMaintenanceFormPageHandler(t, vehicleSvc, &stubMaintenanceSvc{})
	api := MakeMaintenanceAPIHandler(&stubMaintenanceSvc{}, vehicleSvc)
	router := mux.NewRouter()
	router.Handle("/maintenance/new", middleware.CSRFProtection("test-secret")(http.HandlerFunc(page.MaintenanceNew))).Methods(http.MethodGet)
	router.Handle("/api/v1/vehicles/{vehicleId}/maintenance", middleware.APICSRFProtection("test-secret")(http.HandlerFunc(api.CreateOne))).Methods(http.MethodPost)
	authed := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = addAuthContext(r, "u1", "testuser")
		r = r.WithContext(context.WithValue(r.Context(), middleware.AuthMethodContextKey, middleware.AuthMethodCookie))
		router.ServeHTTP(w, r)
	})
	get := httptest.NewRecorder()
	authed.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/maintenance/new?vehicle=v1", nil))
	require.Equal(t, http.StatusOK, get.Code)
	assert.Contains(t, get.Body.String(), `hx-post="/api/v1/vehicles/v1/maintenance"`)
	assert.Contains(t, get.Body.String(), `hx-ext="json-enc"`)
	assert.Contains(t, get.Body.String(), `hx-swap="none"`)
	assert.Contains(t, get.Body.String(), `data-field="cost"`)
	assert.NotContains(t, get.Body.String(), `action="/maintenance/new"`)
	token := regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)">`).FindStringSubmatch(get.Body.String())
	require.Len(t, token, 2)

	for _, csrf := range []string{"", "invalid", token[1]} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/v1/maintenance",
			strings.NewReader(`{"service_type":"oil_change","service_date":"2024-01-15"}`))
		req.Header.Set("HX-Request", "true")
		req.Header.Set("X-CSRF-Token", csrf)
		rec := httptest.NewRecorder()
		authed.ServeHTTP(rec, req)
		if csrf != token[1] {
			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.Contains(t, rec.Body.String(), "CSRF_ERROR")
			assert.Empty(t, rec.Header().Get("HX-Redirect"))
		} else {
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Equal(t, "/maintenance/generated-id?added=true", rec.Header().Get("HX-Redirect"))
		}
	}
}

func TestMaintenanceCreate_PlainAPIHasNoRedirect(t *testing.T) {
	h := MakeMaintenanceAPIHandler(&stubMaintenanceSvc{}, &stubVehicleSvc{getResult: &models.Vehicle{ID: "v1", UserID: "u1"}})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/v1/maintenance",
		strings.NewReader(`{"service_type":"oil_change","service_date":"2024-01-15","cost":49.99,"mileage_at_service":45000}`)),
		map[string]string{"vehicleId": "v1"})
	req = addAuthContext(req, "u1", "testuser")
	rec := httptest.NewRecorder()
	h.CreateOne(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Empty(t, rec.Header().Get("HX-Redirect"))
	assert.Contains(t, rec.Body.String(), `"cost":49.99`)
	assert.Contains(t, rec.Body.String(), `"mileage_at_service":45000`)
}

func TestMaintenanceDetail_CreatedFlash(t *testing.T) {
	page := newTestMaintenanceDetailPageHandler(t,
		&stubVehicleSvc{getResult: &models.Vehicle{ID: "v1", UserID: "u1"}},
		&stubMaintenanceSvc{getResult: &models.MaintenanceRecord{ID: "m1", VehicleID: "v1", ServiceDate: time.Now()}})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/maintenance/m1?added=true", nil), map[string]string{"id": "m1"})
	req = addAuthContext(req, "u1", "testuser")
	rec := httptest.NewRecorder()
	page.MaintenanceDetail(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Maintenance record created successfully.")
}
