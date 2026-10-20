package handlers

import (
	"bytes"
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

func TestMaintenanceEdit_HTMXCookieCSRF(t *testing.T) {
	cost, mileage := 99.99, 30000
	record := &models.MaintenanceRecord{
		ID: "m1", VehicleID: "v1", ServiceType: "other", CustomServiceType: "Inspection",
		ServiceDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		Cost:        &cost, MileageAtService: &mileage, ServiceProvider: "Shop", Notes: "Keep notes",
	}
	maintenance := &stubMaintenanceSvc{getResult: record, updateRes: record}
	vehicle := &stubVehicleSvc{getResult: &models.Vehicle{ID: "v1", UserID: "u1", Year: 2020, Make: "Ford", Model: "Focus"}}
	page := newTestMaintenanceEditPageHandler(t, vehicle, maintenance)
	api := MakeMaintenanceAPIHandler(maintenance, vehicle)
	router := mux.NewRouter()
	router.Handle("/maintenance/{id}/edit", middleware.CSRFProtection("secret")(http.HandlerFunc(page.MaintenanceEdit))).Methods(http.MethodGet)
	router.Handle("/maintenance/{id}", http.HandlerFunc(page.MaintenanceDetail)).Methods(http.MethodGet)
	router.Handle("/api/v1/maintenance/{id}", middleware.APICSRFProtection("secret")(http.HandlerFunc(api.ReplaceOne))).Methods(http.MethodPut)
	authed := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = addAuthContext(r, "u1", "testuser")
		r = r.WithContext(context.WithValue(r.Context(), middleware.AuthMethodContextKey, middleware.AuthMethodCookie))
		router.ServeHTTP(w, r)
	})
	get := httptest.NewRecorder()
	authed.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/maintenance/m1/edit", nil))
	require.Equal(t, http.StatusOK, get.Code)
	token := regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)">`).FindStringSubmatch(get.Body.String())
	require.Len(t, token, 2)

	form := map[string]interface{}{
		"service_type": "other", "custom_service_type": "Inspection", "service_date": "2024-06-15",
		"mileage_at_service": "35000", "cost": "120.50", "service_provider": "Shop", "notes": "Keep notes",
	}
	put := func(csrf string) *httptest.ResponseRecorder {
		body, err := json.Marshal(form)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/maintenance/m1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("HX-Request", "true")
		req.Header.Set("X-CSRF-Token", csrf)
		response := httptest.NewRecorder()
		authed.ServeHTTP(response, req)
		return response
	}
	for _, csrf := range []string{"", "invalid"} {
		assert.Equal(t, http.StatusForbidden, put(csrf).Code)
		assert.Nil(t, maintenance.updates)
	}
	success := put(token[1])
	require.Equal(t, http.StatusOK, success.Code)
	assert.Equal(t, "/maintenance/m1?updated=true", success.Header().Get("HX-Redirect"))
	require.NotNil(t, maintenance.updates)
	assert.Equal(t, 35000, *maintenance.updates.MileageAtService)
	assert.Equal(t, 120.50, *maintenance.updates.Cost)
	assert.Equal(t, "Inspection", *maintenance.updates.CustomServiceType)
	assert.Equal(t, "Shop", *maintenance.updates.ServiceProvider)
	assert.Equal(t, "Keep notes", *maintenance.updates.Notes)
	assert.Equal(t, "2024-06-15", maintenance.updates.ServiceDate.Format("2006-01-02"))

	detail := httptest.NewRecorder()
	authed.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, success.Header().Get("HX-Redirect"), nil))
	assert.Contains(t, detail.Body.String(), "Maintenance record updated successfully.")

	for _, field := range []string{"cost", "mileage_at_service", "service_provider", "notes"} {
		form[field] = ""
	}
	form["service_type"] = "oil_change"
	require.Equal(t, http.StatusOK, put(token[1]).Code)
	assert.True(t, maintenance.updates.ClearCost)
	assert.True(t, maintenance.updates.ClearMileageAtService)
	assert.Empty(t, *maintenance.updates.CustomServiceType)
	assert.Empty(t, *maintenance.updates.ServiceProvider)
	assert.Empty(t, *maintenance.updates.Notes)

	for _, test := range []struct{ field, value string }{
		{"service_type", ""}, {"service_type", "unknown"}, {"service_date", ""},
		{"service_date", "bad-date"}, {"cost", "-1"}, {"cost", "NaN"}, {"cost", "Inf"},
		{"service_date", "2099-06-15"},
		{"mileage_at_service", "1.5"}, {"mileage_at_service", "-1"}, {"mileage_at_service", "99999999999999999999"},
		{"custom_service_type", ""},
	} {
		t.Run(test.field+"="+test.value, func(t *testing.T) {
			old := form[test.field]
			form[test.field] = test.value
			if test.field == "custom_service_type" {
				form["service_type"] = "other"
			}
			maintenance.updates = nil
			response := put(token[1])
			assert.Equal(t, http.StatusBadRequest, response.Code)
			assert.Contains(t, response.Body.String(), `"field":"`+test.field+`"`)
			assert.Empty(t, response.Header().Get("HX-Redirect"))
			assert.Nil(t, maintenance.updates)
			form[test.field] = old
			form["service_type"] = "oil_change"
		})
	}
	for _, value := range []interface{}{nil, 12.5, []string{"1", "2"}} {
		form["cost"] = value
		maintenance.updates = nil
		response := put(token[1])
		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.Contains(t, response.Body.String(), `"field":"cost"`)
		assert.Nil(t, maintenance.updates)
	}
	form["cost"] = ""
	maintenance.updateErr = models.NewDatabaseError("update", assert.AnError)
	failure := put(token[1])
	assert.Equal(t, http.StatusInternalServerError, failure.Code)
	assert.Empty(t, failure.Header().Get("HX-Redirect"))
	vehicle.getResult.UserID = "other"
	assert.Equal(t, http.StatusForbidden, put(token[1]).Code)
}
