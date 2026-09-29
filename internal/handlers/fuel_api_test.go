package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/models"
)

func TestFuelHandler_ListAll(t *testing.T) {
	fillDate := time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		userID     string
		vehicle    *models.Vehicle
		vehicleErr error
		countErr   error
		listErr    error
		wantStatus int
	}{
		{
			name:       "returns fuel records for authenticated user",
			userID:     "u1",
			vehicle:    testVehicle("v1", "u1"),
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns forbidden for vehicle owned by another user",
			userID:     "u1",
			vehicle:    testVehicle("v1", "other-user"),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "returns not found for non-existent vehicle",
			userID:     "u1",
			vehicleErr: models.NewNotFoundError("Vehicle", "v999"),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns internal error when count fails",
			userID:     "u1",
			vehicle:    testVehicle("v1", "u1"),
			countErr:   assert.AnError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "returns internal error when list fails",
			userID:     "u1",
			vehicle:    testVehicle("v1", "u1"),
			listErr:    assert.AnError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "rejects unauthenticated request",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicleStub := &stubVehicleSvc{getResult: tt.vehicle, getErr: tt.vehicleErr}
			fuelStub := &stubFuelSvc{
				countResult: 1,
				listResult:  []*models.FuelRecord{testFuelRecord("f1", "v1", fillDate)},
				countErr:    tt.countErr,
				listErr:     tt.listErr,
			}
			h := MakeFuelAPIHandler(fuelStub, vehicleStub)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/v1/fuel", nil)
			if tt.userID != "" {
				req = addAuthContext(req, tt.userID, "testuser")
			}
			req = mux.SetURLVars(req, map[string]string{"vehicleId": "v1"})
			rec := httptest.NewRecorder()

			h.ListAll(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.True(t, resp["success"].(bool))
				data := resp["data"].([]interface{})
				assert.Len(t, data, 1)
			}
		})
	}
}

func TestFuelHandler_CreateOne(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		vehicle    *models.Vehicle
		body       map[string]interface{}
		createErr  error
		wantStatus int
	}{
		{
			name:    "creates fuel record with valid input",
			userID:  "u1",
			vehicle: testVehicle("v1", "u1"),
			body: map[string]interface{}{
				"fill_date": "2024-02-05T14:30:00Z",
				"mileage":   47250,
				"volume":    12.0,
				"fuel_type": "gasoline",
				"price":     60.0,
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:    "rejects missing required fields",
			userID:  "u1",
			vehicle: testVehicle("v1", "u1"),
			body: map[string]interface{}{
				"mileage":   47250,
				"volume":    12.0,
				"fuel_type": "gasoline",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "rejects invalid fuel type",
			userID:  "u1",
			vehicle: testVehicle("v1", "u1"),
			body: map[string]interface{}{
				"fill_date": "2024-02-05",
				"mileage":   47250,
				"volume":    12.0,
				"fuel_type": "invalid",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "rejects price when volume is zero",
			userID:  "u1",
			vehicle: testVehicle("v1", "u1"),
			body: map[string]interface{}{
				"fill_date": "2024-02-05",
				"mileage":   47250,
				"volume":    0,
				"fuel_type": "gasoline",
				"price":     60.0,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "rejects fractional mileage",
			userID:  "u1",
			vehicle: testVehicle("v1", "u1"),
			body: map[string]interface{}{
				"fill_date": "2024-02-05",
				"mileage":   47250.5,
				"volume":    12.0,
				"fuel_type": "gasoline",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns internal error when create fails",
			userID:     "u1",
			vehicle:    testVehicle("v1", "u1"),
			body:       validFuelBody(),
			createErr:  assert.AnError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "returns forbidden for vehicle owned by another user",
			userID:     "u1",
			vehicle:    testVehicle("v1", "other-user"),
			body:       validFuelBody(),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "rejects unauthenticated request",
			body:       validFuelBody(),
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := MakeFuelAPIHandler(&stubFuelSvc{createErr: tt.createErr}, &stubVehicleSvc{getResult: tt.vehicle})
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/v1/fuel", bytes.NewReader(jsonBody))
			if tt.userID != "" {
				req = addAuthContext(req, tt.userID, "testuser")
			}
			req = mux.SetURLVars(req, map[string]string{"vehicleId": "v1"})
			rec := httptest.NewRecorder()

			h.CreateOne(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantStatus == http.StatusCreated {
				var resp map[string]interface{}
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.Equal(t, "Fuel record created successfully", resp["message"])
				data := resp["data"].(map[string]interface{})
				assert.Equal(t, 5.0, data["price_per_unit"])
				assert.Equal(t, 60.0, data["price"])
			}
		})
	}
}

func TestFuelHandler_GetOne(t *testing.T) {
	fillDate := time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		userID     string
		vehicle    *models.Vehicle
		fuel       *models.FuelRecord
		fuelErr    error
		wantStatus int
	}{
		{
			name:       "returns fuel record when owned by user",
			userID:     "u1",
			vehicle:    testVehicle("v1", "u1"),
			fuel:       testFuelRecord("f1", "v1", fillDate),
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns forbidden for record on vehicle owned by another user",
			userID:     "u1",
			vehicle:    testVehicle("v1", "other-user"),
			fuel:       testFuelRecord("f1", "v1", fillDate),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "returns not found for non-existent record",
			userID:     "u1",
			fuelErr:    models.NewNotFoundError("FuelRecord", "f999"),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "rejects unauthenticated request",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := MakeFuelAPIHandler(
				&stubFuelSvc{getResult: tt.fuel, getErr: tt.fuelErr},
				&stubVehicleSvc{getResult: tt.vehicle},
			)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/fuel/f1", nil)
			if tt.userID != "" {
				req = addAuthContext(req, tt.userID, "testuser")
			}
			req = mux.SetURLVars(req, map[string]string{"id": "f1"})
			rec := httptest.NewRecorder()

			h.GetOne(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.True(t, resp["success"].(bool))
			}
		})
	}
}

func TestFuelHandler_UpdateOne(t *testing.T) {
	fillDate := time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC)
	updated := testFuelRecord("f1", "v1", fillDate)
	updated.Mileage = 47300

	tests := []struct {
		name       string
		userID     string
		vehicle    *models.Vehicle
		fuelErr    error
		body       map[string]interface{}
		wantStatus int
	}{
		{
			name:       "updates fuel record successfully",
			userID:     "u1",
			vehicle:    testVehicle("v1", "u1"),
			body:       map[string]interface{}{"mileage": 47300},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns forbidden for record on vehicle owned by another user",
			userID:     "u1",
			vehicle:    testVehicle("v1", "other-user"),
			body:       map[string]interface{}{"mileage": 47300},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "returns not found for non-existent record",
			userID:     "u1",
			fuelErr:    models.NewNotFoundError("FuelRecord", "f999"),
			body:       map[string]interface{}{"mileage": 47300},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "rejects invalid update",
			userID:     "u1",
			vehicle:    testVehicle("v1", "u1"),
			body:       map[string]interface{}{"volume": 0},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects fractional mileage update",
			userID:     "u1",
			vehicle:    testVehicle("v1", "u1"),
			body:       map[string]interface{}{"mileage": 47300.5},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects unauthenticated request",
			body:       map[string]interface{}{"mileage": 47300},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := MakeFuelAPIHandler(
				&stubFuelSvc{getResult: testFuelRecord("f1", "v1", fillDate), getErr: tt.fuelErr, updateRes: updated},
				&stubVehicleSvc{getResult: tt.vehicle},
			)
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/fuel/f1", bytes.NewReader(jsonBody))
			if tt.userID != "" {
				req = addAuthContext(req, tt.userID, "testuser")
			}
			req = mux.SetURLVars(req, map[string]string{"id": "f1"})
			rec := httptest.NewRecorder()

			h.UpdateOne(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.Equal(t, "Fuel record updated successfully", resp["message"])
			}
		})
	}
}

func TestFuelHandler_RemoveOne(t *testing.T) {
	fillDate := time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		userID     string
		vehicle    *models.Vehicle
		fuelErr    error
		deleteErr  error
		wantStatus int
	}{
		{
			name:       "deletes fuel record successfully",
			userID:     "u1",
			vehicle:    testVehicle("v1", "u1"),
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns forbidden for record on vehicle owned by another user",
			userID:     "u1",
			vehicle:    testVehicle("v1", "other-user"),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "returns not found for non-existent record",
			userID:     "u1",
			fuelErr:    models.NewNotFoundError("FuelRecord", "f999"),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns internal error when delete fails",
			userID:     "u1",
			vehicle:    testVehicle("v1", "u1"),
			deleteErr:  assert.AnError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "rejects unauthenticated request",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := MakeFuelAPIHandler(
				&stubFuelSvc{getResult: testFuelRecord("f1", "v1", fillDate), getErr: tt.fuelErr, deleteErr: tt.deleteErr},
				&stubVehicleSvc{getResult: tt.vehicle},
			)
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/fuel/f1", nil)
			if tt.userID != "" {
				req = addAuthContext(req, tt.userID, "testuser")
			}
			req = mux.SetURLVars(req, map[string]string{"id": "f1"})
			rec := httptest.NewRecorder()

			h.RemoveOne(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.Equal(t, "Fuel record deleted successfully", resp["message"])
			}
		})
	}
}

func TestFuelHandler_MalformedJSON_ReturnsBadRequest(t *testing.T) {
	fillDate := time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC)

	t.Run("create rejects malformed body", func(t *testing.T) {
		h := MakeFuelAPIHandler(&stubFuelSvc{}, &stubVehicleSvc{getResult: testVehicle("v1", "u1")})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/v1/fuel", bytes.NewReader([]byte("{invalid")))
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"vehicleId": "v1"})
		rec := httptest.NewRecorder()

		h.CreateOne(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("update rejects malformed body", func(t *testing.T) {
		h := MakeFuelAPIHandler(
			&stubFuelSvc{getResult: testFuelRecord("f1", "v1", fillDate)},
			&stubVehicleSvc{getResult: testVehicle("v1", "u1")},
		)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/fuel/f1", bytes.NewReader([]byte("{invalid")))
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		rec := httptest.NewRecorder()

		h.UpdateOne(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestFuelHandler_UpdateOne_ServiceError_ReturnsInternalError(t *testing.T) {
	fillDate := time.Date(2024, 2, 5, 14, 30, 0, 0, time.UTC)
	h := MakeFuelAPIHandler(
		&stubFuelSvc{getResult: testFuelRecord("f1", "v1", fillDate), updateErr: assert.AnError},
		&stubVehicleSvc{getResult: testVehicle("v1", "u1")},
	)
	body, err := json.Marshal(map[string]interface{}{"mileage": 47300})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/fuel/f1", bytes.NewReader(body))
	req = addAuthContext(req, "u1", "testuser")
	req = mux.SetURLVars(req, map[string]string{"id": "f1"})
	rec := httptest.NewRecorder()

	h.UpdateOne(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func testVehicle(id, userID string) *models.Vehicle {
	return &models.Vehicle{
		ID:     id,
		UserID: userID,
		VIN:    "ABC12345678901234",
		Make:   "Ford",
		Model:  "Focus",
		Year:   2020,
		Status: models.VehicleStatusActive,
	}
}

func testFuelRecord(id, vehicleID string, fillDate time.Time) *models.FuelRecord {
	pricePerUnit := 5.0
	return &models.FuelRecord{
		ID:           id,
		VehicleID:    vehicleID,
		FillDate:     fillDate,
		Mileage:      47250,
		Volume:       12.0,
		FuelType:     "gasoline",
		PricePerUnit: &pricePerUnit,
	}
}

func validFuelBody() map[string]interface{} {
	return map[string]interface{}{
		"fill_date": "2024-02-05T14:30:00Z",
		"mileage":   47250,
		"volume":    12.0,
		"fuel_type": "gasoline",
	}
}
