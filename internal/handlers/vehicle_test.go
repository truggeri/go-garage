package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
	"github.com/truggeri/go-garage/internal/services"
)

type stubVehicleSvc struct {
	createErr   error
	getResult   *models.Vehicle
	getErr      error
	listResult  []*models.Vehicle
	listErr     error
	countResult int
	countErr    error
	updateRes   *models.Vehicle
	updateErr   error
	deleteErr   error
	deletedID   string
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

func (s *stubVehicleSvc) SaveVehicle(_ context.Context, _ *models.Vehicle) error {
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
