package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

type capturingFuelSvc struct {
	stubFuelSvc
	created *models.FuelRecord
	updates *services.FuelUpdates
}

func (s *capturingFuelSvc) CreateFuel(ctx context.Context, r *models.FuelRecord) error {
	s.created = r
	return s.stubFuelSvc.CreateFuel(ctx, r)
}

func (s *capturingFuelSvc) UpdateFuel(_ context.Context, _ string, u services.FuelUpdates) (*models.FuelRecord, error) {
	s.updates = &u
	return s.updateRes, s.updateErr
}

func htmxFuelRequest(method, target, body string, vars map[string]string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("HX-Request", "true")
	req = addAuthContext(req, "u1", "testuser")
	return mux.SetURLVars(req, vars)
}

func fuelErrorFields(t *testing.T, rec *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var resp struct {
		Error struct {
			Details []FieldError `json:"details"`
		} `json:"error"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	fields := map[string]string{}
	for _, d := range resp.Error.Details {
		fields[d.Field] = d.Message
	}
	return fields
}

func TestFuelHandler_CreateOne_HTMXFormValues_CreatesAndRedirects(t *testing.T) {
	svc := &capturingFuelSvc{}
	h := MakeFuelAPIHandler(svc, &stubVehicleSvc{getResult: testVehicle("v1", "u1")})
	body := `{"vehicle_id":"v1","fill_date":"2024-02-05","fuel_type":"gasoline","mileage":"47250","volume":"12.5",` +
		`"price_per_unit":"3.5","octane_rating":"","partial_fill":"on","location":"Shell","brand":"","notes":"","city_driving_percentage":"60","vehicle_reported_mpg":""}`
	rec := httptest.NewRecorder()

	h.CreateOne(rec, htmxFuelRequest(http.MethodPost, "/api/v1/vehicles/v1/fuel", body, map[string]string{"vehicleId": "v1"}))

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "/fuel?added=true", rec.Header().Get("HX-Redirect"))
	require.NotNil(t, svc.created)
	assert.Equal(t, "v1", svc.created.VehicleID)
	assert.Equal(t, 47250, svc.created.Mileage)
	assert.InDelta(t, 12.5, svc.created.Volume, 0.001)
	assert.True(t, svc.created.PartialFill)
	assert.Nil(t, svc.created.OctaneRating)
	require.NotNil(t, svc.created.CityDrivingPercentage)
	assert.Equal(t, 60, *svc.created.CityDrivingPercentage)
}

func TestFuelHandler_CreateOne_HTMXInvalidValues_ReturnsInlineFieldErrors(t *testing.T) {
	h := MakeFuelAPIHandler(&capturingFuelSvc{}, &stubVehicleSvc{getResult: testVehicle("v1", "u1")})

	t.Run("non-numeric mileage", func(t *testing.T) {
		rec := httptest.NewRecorder()
		body := `{"fill_date":"2024-02-05","fuel_type":"gasoline","mileage":"abc","volume":"12"}`
		h.CreateOne(rec, htmxFuelRequest(http.MethodPost, "/api/v1/vehicles/v1/fuel", body, map[string]string{"vehicleId": "v1"}))
		require.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, fuelErrorFields(t, rec), "mileage")
	})

	t.Run("missing required fields", func(t *testing.T) {
		rec := httptest.NewRecorder()
		body := `{"fill_date":"","fuel_type":"","mileage":"","volume":""}`
		h.CreateOne(rec, htmxFuelRequest(http.MethodPost, "/api/v1/vehicles/v1/fuel", body, map[string]string{"vehicleId": "v1"}))
		require.Equal(t, http.StatusBadRequest, rec.Code)
		fields := fuelErrorFields(t, rec)
		for _, f := range []string{"fill_date", "fuel_type", "mileage", "volume"} {
			assert.Contains(t, fields, f)
		}
	})

	t.Run("vehicle owned by another user", func(t *testing.T) {
		other := MakeFuelAPIHandler(&capturingFuelSvc{}, &stubVehicleSvc{getResult: testVehicle("v1", "other")})
		rec := httptest.NewRecorder()
		other.CreateOne(rec, htmxFuelRequest(http.MethodPost, "/api/v1/vehicles/v1/fuel", `{}`, map[string]string{"vehicleId": "v1"}))
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Empty(t, rec.Header().Get("HX-Redirect"))
	})
}

func TestFuelHandler_UpdateOne_HTMXFormValues_UpdatesAndRedirects(t *testing.T) {
	fillDate := time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC)
	existing := testFuelRecord("f1", "v1", fillDate)
	existing.PartialFill = true
	svc := &capturingFuelSvc{stubFuelSvc: stubFuelSvc{getResult: existing, updateRes: existing}}
	h := MakeFuelAPIHandler(svc, &stubVehicleSvc{getResult: testVehicle("v1", "u1")})
	body := `{"fill_date":"2024-02-06","fuel_type":"gasoline","mileage":"47300","volume":"11.25","price_per_unit":"","octane_rating":"","location":"","brand":"","notes":"","city_driving_percentage":"","vehicle_reported_mpg":""}`
	rec := httptest.NewRecorder()

	h.UpdateOne(rec, htmxFuelRequest(http.MethodPut, "/api/v1/fuel/f1", body, map[string]string{"id": "f1"}))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "/fuel/f1?updated=true", rec.Header().Get("HX-Redirect"))
	require.NotNil(t, svc.updates)
	assert.Equal(t, 47300, *svc.updates.Mileage)
	assert.InDelta(t, 11.25, *svc.updates.Volume, 0.001)
	require.NotNil(t, svc.updates.PartialFill)
	assert.False(t, *svc.updates.PartialFill, "unchecked checkbox clears partial fill")
	require.NotNil(t, svc.updates.Location)
	assert.Empty(t, *svc.updates.Location)
}

func TestFuelHandler_UpdateOne_HTMXInvalidValues_ReturnsInlineFieldErrors(t *testing.T) {
	existing := testFuelRecord("f1", "v1", time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC))
	h := MakeFuelAPIHandler(&capturingFuelSvc{stubFuelSvc: stubFuelSvc{getResult: existing, updateRes: existing}},
		&stubVehicleSvc{getResult: testVehicle("v1", "u1")})

	rec := httptest.NewRecorder()
	body := `{"fill_date":"2024-02-06","fuel_type":"gasoline","mileage":"","volume":"-3"}`
	h.UpdateOne(rec, htmxFuelRequest(http.MethodPut, "/api/v1/fuel/f1", body, map[string]string{"id": "f1"}))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, fuelErrorFields(t, rec), "mileage")
}

func TestFuelHandler_RemoveOne_HTMX_RedirectsToVehicleFuelList(t *testing.T) {
	existing := testFuelRecord("f1", "v1", time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC))
	h := MakeFuelAPIHandler(&capturingFuelSvc{stubFuelSvc: stubFuelSvc{getResult: existing}},
		&stubVehicleSvc{getResult: testVehicle("v1", "u1")})
	rec := httptest.NewRecorder()

	h.RemoveOne(rec, htmxFuelRequest(http.MethodDelete, "/api/v1/fuel/f1", "", map[string]string{"id": "f1"}))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "/fuel?vehicle=v1", rec.Header().Get("HX-Redirect"))
}

func TestFuelHandler_RemoveOne_NonHTMX_OmitsRedirect(t *testing.T) {
	existing := testFuelRecord("f1", "v1", time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC))
	h := MakeFuelAPIHandler(&capturingFuelSvc{stubFuelSvc: stubFuelSvc{getResult: existing}},
		&stubVehicleSvc{getResult: testVehicle("v1", "u1")})
	req := htmxFuelRequest(http.MethodDelete, "/api/v1/fuel/f1", "", map[string]string{"id": "f1"})
	req.Header.Del("HX-Request")
	rec := httptest.NewRecorder()

	h.RemoveOne(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("HX-Redirect"))
}
