package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
	"github.com/truggeri/go-garage/internal/services"
)

type stubVehicleSvc struct {
	createErr    error
	getResult    *models.Vehicle
	getErr       error
	listResult   []*models.Vehicle
	listErr      error
	countResult  int
	countErr     error
	updateRes    *models.Vehicle
	updateErr    error
	savedVehicle *models.Vehicle
	deleteErr    error
	deletedID    string
}

func (s *stubVehicleSvc) CreateVehicle(_ context.Context, v *models.Vehicle) error {
	if s.createErr != nil {
		return s.createErr
	}
	v.ID = "generated-id"
	return nil
}

func (s *stubVehicleSvc) GetVehicle(_ context.Context, _ string) (*models.Vehicle, error) {
	return s.getResult, s.getErr
}

func (s *stubVehicleSvc) GetUserVehicles(_ context.Context, _ string) ([]*models.Vehicle, error) {
	return s.listResult, s.listErr
}

func (s *stubVehicleSvc) UpdateVehicle(_ context.Context, _ string, _ services.VehicleUpdates) (*models.Vehicle, error) {
	return s.updateRes, s.updateErr
}

func (s *stubVehicleSvc) ArchiveVehicle(_ context.Context, _ string, _ models.VehicleStatus) (*models.Vehicle, error) {
	return nil, nil
}

func (s *stubVehicleSvc) DeleteVehicle(_ context.Context, id string) error {
	s.deletedID = id
	return s.deleteErr
}

func (s *stubVehicleSvc) SaveVehicle(_ context.Context, vehicle *models.Vehicle) error {
	s.savedVehicle = vehicle
	return s.updateErr
}

func (s *stubVehicleSvc) ListVehicles(_ context.Context, _ repositories.VehicleFilters, _ repositories.PaginationParams) ([]*models.Vehicle, error) {
	return s.listResult, s.listErr
}

func (s *stubVehicleSvc) CountVehicles(_ context.Context, _ repositories.VehicleFilters) (int, error) {
	return s.countResult, s.countErr
}

func (s *stubVehicleSvc) VerifyOwnership(_ context.Context, _, _ string) error {
	return nil
}

func addAuthContext(r *http.Request, userID, userName string) *http.Request {
	acct := &middleware.AccountInfo{ID: userID, Name: userName}
	ctx := context.WithValue(r.Context(), middleware.AccountContextKey, acct)
	return r.WithContext(ctx)
}

func addResourceContext(r *http.Request, resource interface{}) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.LoadedResourceContextKey, resource)
	return r.WithContext(ctx)
}

func TestVehicleHandler_ListAll(t *testing.T) {
	t.Run("returns vehicles for authenticated user", func(t *testing.T) {
		stub := &stubVehicleSvc{
			countResult: 1,
			listResult: []*models.Vehicle{
				{ID: "v1", UserID: "u1", VIN: "ABC12345678901234", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
			},
		}
		h := MakeVehicleAPIHandler(stub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		req = addAuthContext(req, "u1", "john_doe")
		rec := httptest.NewRecorder()

		h.ListAll(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.True(t, resp["success"].(bool))
		data := resp["data"].([]interface{})
		assert.Len(t, data, 1)
	})

	t.Run("rejects unauthenticated request", func(t *testing.T) {
		stub := &stubVehicleSvc{}
		h := MakeVehicleAPIHandler(stub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles", nil)
		rec := httptest.NewRecorder()

		h.ListAll(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestVehicleHandler_CreateOne(t *testing.T) {
	t.Run("creates vehicle with valid input", func(t *testing.T) {
		stub := &stubVehicleSvc{}
		h := MakeVehicleAPIHandler(stub)

		body := map[string]interface{}{
			"vin": "1HGBH41JXMN109186", "make": "Honda", "model": "Civic", "year": 2021,
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		h.CreateOne(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.True(t, resp["success"].(bool))
		assert.Equal(t, "Vehicle created successfully", resp["message"])
	})

	t.Run("rejects missing required fields", func(t *testing.T) {
		stub := &stubVehicleSvc{}
		h := MakeVehicleAPIHandler(stub)

		body := map[string]interface{}{"make": "Honda"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		h.CreateOne(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects duplicate VIN", func(t *testing.T) {
		stub := &stubVehicleSvc{
			createErr: models.NewDuplicateError("Vehicle", "vin", "1HGBH41JXMN109186"),
		}
		h := MakeVehicleAPIHandler(stub)

		body := map[string]interface{}{
			"vin": "1HGBH41JXMN109186", "make": "Honda", "model": "Civic", "year": 2021,
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		h.CreateOne(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("htmx form creates vehicle with numeric inputs and redirects", func(t *testing.T) {
		h := MakeVehicleAPIHandler(&stubVehicleSvc{})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewBufferString(
			`{"vin":"1HGBH41JXMN109186","make":" Honda ","model":" Civic ","year":"2021","purchase_price":"25000.50","purchase_mileage":"12000","current_mileage":"15000","purchase_date":"","display_name":" My Civic "}`))
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		h.CreateOne(rec, addAuthContext(req, "u1", "testuser"))

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "/vehicles/generated-id?added=true", rec.Header().Get("HX-Redirect"))
		var resp struct {
			Data map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, float64(2021), resp.Data["year"])
		assert.Equal(t, float64(25000.50), resp.Data["purchase_price"])
		assert.Equal(t, float64(12000), resp.Data["purchase_mileage"])
		assert.Equal(t, float64(15000), resp.Data["current_mileage"])
		assert.Equal(t, "Honda", resp.Data["make"])
		assert.Equal(t, "Civic", resp.Data["model"])
		assert.Equal(t, "My Civic", resp.Data["display_name"])
	})

	t.Run("htmx form returns field errors for missing and invalid inputs", func(t *testing.T) {
		h := MakeVehicleAPIHandler(&stubVehicleSvc{})
		for _, tc := range []struct {
			body  string
			field string
		}{
			{`{"vin":"1HGBH41JXMN109186","make":"","model":"Civic","year":"2021"}`, "make"},
			{`{"vin":"1HGBH41JXMN109186","make":"  ","model":"Civic","year":"2021"}`, "make"},
			{`{"vin":"1HGBH41JXMN109186","make":"Honda","model":"Civic","year":"abc"}`, "year"},
			{`{"vin":"1HGBH41JXMN109186","make":"Honda","model":"Civic","year":"2021","purchase_date":"bad"}`, "purchase_date"},
			{`{"vin":"1HGBH41JXMN109186","make":"Honda","model":"Civic","year":"2021","current_mileage":"-1"}`, "current_mileage"},
		} {
			t.Run(tc.field+" "+tc.body, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewBufferString(tc.body))
				req.Header.Set("HX-Request", "true")
				rec := httptest.NewRecorder()
				h.CreateOne(rec, addAuthContext(req, "u1", "testuser"))

				assert.Equal(t, http.StatusBadRequest, rec.Code)
				assert.Empty(t, rec.Header().Get("HX-Redirect"))
				var resp struct {
					Error struct {
						Details []FieldError `json:"details"`
					} `json:"error"`
				}
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				require.NotEmpty(t, resp.Error.Details)
				assert.Equal(t, tc.field, resp.Error.Details[0].Field)
			})
		}
	})

	t.Run("htmx duplicate VIN returns inline error", func(t *testing.T) {
		h := MakeVehicleAPIHandler(&stubVehicleSvc{
			createErr: models.NewDuplicateError("Vehicle", "vin", "1HGBH41JXMN109186"),
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", bytes.NewBufferString(
			`{"vin":"1HGBH41JXMN109186","make":"Honda","model":"Civic","year":"2021"}`))
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()
		h.CreateOne(rec, addAuthContext(req, "u1", "testuser"))

		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Contains(t, rec.Body.String(), `"field":"vin"`)
		assert.Empty(t, rec.Header().Get("HX-Redirect"))
	})
}

func TestVehicleHandler_GetOne(t *testing.T) {
	t.Run("returns vehicle when owned by user", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getResult: &models.Vehicle{
				ID: "v1", UserID: "u1", VIN: "ABC12345678901234",
				Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
			},
		}
		h := MakeVehicleAPIHandler(stub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/v1", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("returns forbidden for vehicle owned by another user", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getResult: &models.Vehicle{
				ID: "v1", UserID: "other-user", VIN: "ABC12345678901234",
				Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
			},
		}
		h := MakeVehicleAPIHandler(stub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/v1", nil)
		req = addAuthContext(req, "different-user", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns not found for non-existent vehicle", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getErr: models.NewNotFoundError("Vehicle", "v999"),
		}
		h := MakeVehicleAPIHandler(stub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/v999", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v999"})
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestVehicleHandler_ReplaceOne(t *testing.T) {
	t.Run("updates vehicle successfully", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getResult: &models.Vehicle{
				ID: "v1", UserID: "u1", VIN: "ABC12345678901234",
				Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
			},
			updateRes: &models.Vehicle{
				ID: "v1", UserID: "u1", VIN: "ABC12345678901234",
				Make: "Ford", Model: "Focus", Year: 2020, Color: "Blue", Status: models.VehicleStatusActive,
			},
		}
		h := MakeVehicleAPIHandler(stub)

		body := map[string]interface{}{"color": "Blue"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/vehicles/v1", bytes.NewReader(jsonBody))
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.ReplaceOne(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "Vehicle updated successfully", resp["message"])
	})

	t.Run("updates the complete vehicle from an htmx form and redirects with a flash", func(t *testing.T) {
		purchaseDate := time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC)
		purchasePrice := 25000.0
		purchaseMileage := 12000
		currentMileage := 45000
		existing := &models.Vehicle{
			ID: "v1", UserID: "u1", VIN: "1HGBH41JXMN109186",
			Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
			DisplayName: "Daily driver", Color: "Blue", LicensePlate: "ABC-1234",
			PurchaseDate: &purchaseDate, PurchasePrice: &purchasePrice,
			PurchaseMileage: &purchaseMileage, CurrentMileage: &currentMileage, Notes: "Good car",
		}
		stub := &stubVehicleSvc{getResult: existing}
		h := MakeVehicleAPIHandler(stub)
		body := map[string]interface{}{
			"vin": "1HGCM82633A004352", "make": "Toyota", "model": "Camry", "year": "2021",
			"display_name": "Daily driver", "color": "Blue", "license_plate": "ABC-1234",
			"purchase_date": "2021-06-15", "purchase_price": "25000",
			"purchase_mileage": "12000", "current_mileage": "45000", "notes": "Good car",
		}
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/vehicles/v1", bytes.NewReader(jsonBody))
		req.Header.Set("HX-Request", "true")
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.ReplaceOne(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "/vehicles/v1?updated=true", rec.Header().Get("HX-Redirect"))
		require.NotNil(t, stub.savedVehicle)
		assert.Equal(t, "u1", stub.savedVehicle.UserID)
		assert.Equal(t, models.VehicleStatusActive, stub.savedVehicle.Status)
		assert.Equal(t, "1HGCM82633A004352", stub.savedVehicle.VIN)
		assert.Equal(t, "Toyota", stub.savedVehicle.Make)
		assert.Equal(t, "Camry", stub.savedVehicle.Model)
		assert.Equal(t, 2021, stub.savedVehicle.Year)
		assert.Equal(t, "Daily driver", stub.savedVehicle.DisplayName)
		assert.Equal(t, "Blue", stub.savedVehicle.Color)
		assert.Equal(t, "ABC-1234", stub.savedVehicle.LicensePlate)
		assert.Equal(t, &purchaseDate, stub.savedVehicle.PurchaseDate)
		assert.Equal(t, &purchasePrice, stub.savedVehicle.PurchasePrice)
		assert.Equal(t, &purchaseMileage, stub.savedVehicle.PurchaseMileage)
		assert.Equal(t, &currentMileage, stub.savedVehicle.CurrentMileage)
		assert.Equal(t, "Good car", stub.savedVehicle.Notes)
	})

	t.Run("clears optional fields when the htmx form submits them empty", func(t *testing.T) {
		purchaseDate := time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC)
		purchasePrice := 25000.0
		purchaseMileage := 12000
		currentMileage := 45000
		stub := &stubVehicleSvc{getResult: &models.Vehicle{
			ID: "v1", UserID: "u1", VIN: "1HGBH41JXMN109186", Make: "Ford",
			Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
			PurchaseDate: &purchaseDate, PurchasePrice: &purchasePrice,
			PurchaseMileage: &purchaseMileage, CurrentMileage: &currentMileage,
		}}
		h := MakeVehicleAPIHandler(stub)
		body := map[string]interface{}{
			"vin": "1HGBH41JXMN109186", "make": "Ford", "model": "Focus", "year": "2020",
			"purchase_date": "", "purchase_price": "", "purchase_mileage": "", "current_mileage": "",
		}
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/vehicles/v1", bytes.NewReader(jsonBody))
		req.Header.Set("HX-Request", "true")
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.ReplaceOne(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, stub.savedVehicle)
		assert.Nil(t, stub.savedVehicle.PurchaseDate)
		assert.Nil(t, stub.savedVehicle.PurchasePrice)
		assert.Nil(t, stub.savedVehicle.PurchaseMileage)
		assert.Nil(t, stub.savedVehicle.CurrentMileage)
	})

	t.Run("returns numeric input errors inline for htmx forms", func(t *testing.T) {
		stub := &stubVehicleSvc{getResult: &models.Vehicle{
			ID: "v1", UserID: "u1", VIN: "1HGBH41JXMN109186", Make: "Ford",
			Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
		}}
		h := MakeVehicleAPIHandler(stub)
		body := map[string]interface{}{
			"vin": "1HGBH41JXMN109186", "make": "Ford", "model": "Focus", "year": "2020",
			"current_mileage": "not-a-number",
		}
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/vehicles/v1", bytes.NewReader(jsonBody))
		req.Header.Set("HX-Request", "true")
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.ReplaceOne(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), `"field":"current_mileage"`)
		assert.Nil(t, stub.savedVehicle)
	})

	t.Run("returns required field errors inline for htmx forms", func(t *testing.T) {
		stub := &stubVehicleSvc{getResult: &models.Vehicle{
			ID: "v1", UserID: "u1", VIN: "1HGBH41JXMN109186", Make: "Ford",
			Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
		}}
		h := MakeVehicleAPIHandler(stub)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/vehicles/v1", bytes.NewBufferString(
			`{"vin":"1HGBH41JXMN109186","make":"","model":"Focus","year":"2020"}`))
		req.Header.Set("HX-Request", "true")
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.ReplaceOne(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), `"field":"make"`)
		assert.Nil(t, stub.savedVehicle)
	})

	t.Run("returns string field errors inline for htmx forms", func(t *testing.T) {
		stub := &stubVehicleSvc{getResult: &models.Vehicle{
			ID: "v1", UserID: "u1", VIN: "1HGBH41JXMN109186", Make: "Ford",
			Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
		}}
		h := MakeVehicleAPIHandler(stub)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/vehicles/v1", bytes.NewBufferString(
			`{"vin":"1HGBH41JXMN109186","make":123,"model":"Focus","year":"2020"}`))
		req.Header.Set("HX-Request", "true")
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.ReplaceOne(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), `"field":"make"`)
		assert.Nil(t, stub.savedVehicle)
	})

	t.Run("returns model validation errors inline for htmx forms", func(t *testing.T) {
		stub := &stubVehicleSvc{getResult: &models.Vehicle{
			ID: "v1", UserID: "u1", VIN: "1HGBH41JXMN109186", Make: "Ford",
			Model: "Focus", Year: 2020, Status: "invalid",
		}}
		h := MakeVehicleAPIHandler(stub)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/vehicles/v1", bytes.NewBufferString(
			`{"vin":"1HGBH41JXMN109186","make":"Ford","model":"Focus","year":"2020","purchase_price":"-1"}`))
		req.Header.Set("HX-Request", "true")
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.ReplaceOne(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), `"field":"purchase_price"`)
		assert.Nil(t, stub.savedVehicle)
	})

	t.Run("returns save errors for htmx forms", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getResult: &models.Vehicle{
				ID: "v1", UserID: "u1", VIN: "1HGBH41JXMN109186",
				Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
			},
			updateErr: assert.AnError,
		}
		h := MakeVehicleAPIHandler(stub)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/vehicles/v1", bytes.NewBufferString(
			`{"vin":"1HGBH41JXMN109186","make":"Ford","model":"Focus","year":"2020"}`))
		req.Header.Set("HX-Request", "true")
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.ReplaceOne(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.NotEmpty(t, stub.savedVehicle)
		assert.Empty(t, rec.Header().Get("HX-Redirect"))
	})

	t.Run("returns duplicate VIN errors inline for htmx forms", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getResult: &models.Vehicle{
				ID: "v1", UserID: "u1", VIN: "1HGBH41JXMN109186",
				Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
			},
			updateErr: models.NewDuplicateError("Vehicle", "vin", "1HGCM82633A004352"),
		}
		h := MakeVehicleAPIHandler(stub)
		body := map[string]interface{}{
			"vin": "1HGCM82633A004352", "make": "Ford", "model": "Focus", "year": "2020",
		}
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/vehicles/v1", bytes.NewReader(jsonBody))
		req.Header.Set("HX-Request", "true")
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.ReplaceOne(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), `"field":"vin"`)
		assert.Contains(t, rec.Body.String(), "VIN already exists")
	})
}

func TestVehicleHandler_RemoveOne(t *testing.T) {
	t.Run("deletes vehicle successfully", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getResult: &models.Vehicle{
				ID: "v1", UserID: "u1", VIN: "ABC12345678901234",
				Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
			},
		}
		h := MakeVehicleAPIHandler(stub)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/vehicles/v1", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.RemoveOne(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "Vehicle deleted successfully", resp["message"])
		assert.Empty(t, rec.Header().Get("HX-Redirect"))
		assert.Equal(t, "v1", stub.deletedID)
	})

	t.Run("redirects htmx delete to vehicle list", func(t *testing.T) {
		stub := &stubVehicleSvc{getResult: &models.Vehicle{ID: "v1", UserID: "u1"}}
		h := MakeVehicleAPIHandler(stub)
		req := mux.SetURLVars(addAuthContext(httptest.NewRequest(http.MethodDelete, "/api/v1/vehicles/v1", nil), "u1", "testuser"), map[string]string{"id": "v1"})
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		h.RemoveOne(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "/vehicles", rec.Header().Get("HX-Redirect"))
		assert.Equal(t, "v1", stub.deletedID)
	})

	t.Run("preserves the API error without redirect", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "u1"},
			deleteErr: models.NewValidationError("", "Vehicle cannot be deleted"),
		}
		h := MakeVehicleAPIHandler(stub)
		req := mux.SetURLVars(addAuthContext(httptest.NewRequest(http.MethodDelete, "/api/v1/vehicles/v1", nil), "u1", "testuser"), map[string]string{"id": "v1"})
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		h.RemoveOne(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Empty(t, rec.Header().Get("HX-Redirect"))
		assert.Contains(t, rec.Body.String(), "Vehicle cannot be deleted")
	})
}

func TestVehicleHandler_GetStats(t *testing.T) {
	t.Run("returns stats for owned vehicle", func(t *testing.T) {
		stub := &stubVehicleSvc{
			getResult: &models.Vehicle{
				ID: "v1", UserID: "u1", VIN: "ABC12345678901234",
				Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive,
			},
		}
		h := MakeVehicleAPIHandler(stub)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/v1/stats", nil)
		req = addAuthContext(req, "u1", "testuser")
		req = mux.SetURLVars(req, map[string]string{"id": "v1"})
		rec := httptest.NewRecorder()

		h.GetStats(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.True(t, resp["success"].(bool))
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "v1", data["vehicle_id"])
	})
}
