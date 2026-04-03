package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
	"github.com/truggeri/go-garage/internal/services"
	"github.com/truggeri/go-garage/internal/templateengine"
)

type stubFuelSvc struct {
	createErr   error
	getResult   *models.FuelRecord
	getErr      error
	listResult  []*models.FuelRecord
	listErr     error
	countResult int
	countErr    error
	updateRes   *models.FuelRecord
	updateErr   error
	deleteErr   error
}

func (s *stubFuelSvc) CreateFuel(_ context.Context, r *models.FuelRecord) error {
	if s.createErr != nil {
		return s.createErr
	}
	r.ID = "generated-fuel-id"
	return nil
}

func (s *stubFuelSvc) GetFuel(_ context.Context, _ string) (*models.FuelRecord, error) {
	return s.getResult, s.getErr
}

func (s *stubFuelSvc) GetVehicleFuel(_ context.Context, _ string) ([]*models.FuelRecord, error) {
	return s.listResult, s.listErr
}

func (s *stubFuelSvc) ListFuel(_ context.Context, _ repositories.FuelFilters, _ repositories.PaginationParams) ([]*models.FuelRecord, error) {
	return s.listResult, s.listErr
}

func (s *stubFuelSvc) CountFuel(_ context.Context, _ repositories.FuelFilters) (int, error) {
	return s.countResult, s.countErr
}

func (s *stubFuelSvc) UpdateFuel(_ context.Context, _ string, _ services.FuelUpdates) (*models.FuelRecord, error) {
	return s.updateRes, s.updateErr
}

func (s *stubFuelSvc) DeleteFuel(_ context.Context, _ string) error {
	return s.deleteErr
}

func newTestFuelPageHandler(
	t *testing.T,
	vehicleSvc *stubVehicleSvc,
	fuelSvc *stubFuelSvc,
) *PageHandler {
	t.Helper()
	engine := templateengine.NewEngine("../../web/templates", true)
	return NewPageHandler(engine, &mockAuthService{}, vehicleSvc, &stubMaintenanceSvc{}, fuelSvc, nil, nil)
}

func TestPageHandler_FuelList(t *testing.T) {
	fillDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	pricePerUnit := 3.499

	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}
	baseRecord := &models.FuelRecord{
		ID: "f1", VehicleID: "v1", FuelType: "gasoline",
		FillDate: fillDate, Volume: 12.5, Mileage: 45000,
		PricePerUnit: &pricePerUnit,
	}

	t.Run("renders fuel list for authenticated user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		fuelStub := &stubFuelSvc{listResult: []*models.FuelRecord{baseRecord}}
		handler := newTestFuelPageHandler(t, vehicleStub, fuelStub)

		req := httptest.NewRequest(http.MethodGet, "/fuel", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Fuel Records")
		assert.Contains(t, body, "Gasoline")
		assert.Contains(t, body, "12.50 gal")
	})

	t.Run("renders empty state when no records", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		fuelStub := &stubFuelSvc{listResult: []*models.FuelRecord{}}
		handler := newTestFuelPageHandler(t, vehicleStub, fuelStub)

		req := httptest.NewRequest(http.MethodGet, "/fuel", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "No fuel records yet")
	})

	t.Run("renders empty state with no vehicles", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{}}
		fuelStub := &stubFuelSvc{}
		handler := newTestFuelPageHandler(t, vehicleStub, fuelStub)

		req := httptest.NewRequest(http.MethodGet, "/fuel", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Add a Vehicle")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestFuelPageHandler(t, &stubVehicleSvc{}, &stubFuelSvc{})

		req := httptest.NewRequest(http.MethodGet, "/fuel", nil)
		rec := httptest.NewRecorder()

		handler.FuelList(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("returns 403 when filtering by vehicle not owned by user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestFuelPageHandler(t, vehicleStub, &stubFuelSvc{})

		req := httptest.NewRequest(http.MethodGet, "/fuel?vehicle=other-vehicle-id", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelList(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("shows success flash when added=true", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		fuelStub := &stubFuelSvc{listResult: []*models.FuelRecord{baseRecord}}
		handler := newTestFuelPageHandler(t, vehicleStub, fuelStub)

		req := httptest.NewRequest(http.MethodGet, "/fuel?added=true", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelList(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Fuel record added successfully")
	})
}

func TestPageHandler_FuelDetail(t *testing.T) {
	fillDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	pricePerUnit := 3.499

	baseRecord := &models.FuelRecord{
		ID: "f1", VehicleID: "v1", FuelType: "gasoline",
		FillDate: fillDate, Volume: 12.5, Mileage: 45000,
		PricePerUnit: &pricePerUnit, Location: "Shell Station",
	}
	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}

	t.Run("renders fuel detail page for authenticated user", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestFuelPageHandler(t, vehicleStub, fuelStub)

		req := httptest.NewRequest(http.MethodGet, "/fuel/f1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelDetail(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Gasoline")
		assert.Contains(t, body, "Shell Station")
		assert.Contains(t, body, "Ford")
		assert.Contains(t, body, "Focus")
		assert.Contains(t, body, "2020")
	})

	t.Run("returns 404 when record not found", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getErr: models.NewNotFoundError("FuelRecord", "f99")}
		handler := newTestFuelPageHandler(t, &stubVehicleSvc{}, fuelStub)

		req := httptest.NewRequest(http.MethodGet, "/fuel/f99", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "f99"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelDetail(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 403 when vehicle belongs to another user", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		handler := newTestFuelPageHandler(t, vehicleStub, fuelStub)

		req := httptest.NewRequest(http.MethodGet, "/fuel/f1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelDetail(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestFuelPageHandler(t, &stubVehicleSvc{}, &stubFuelSvc{})

		req := httptest.NewRequest(http.MethodGet, "/fuel/f1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		rec := httptest.NewRecorder()

		handler.FuelDetail(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPageHandler_FuelDelete(t *testing.T) {
	fillDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	baseRecord := &models.FuelRecord{
		ID: "f1", VehicleID: "v1", FuelType: "gasoline",
		FillDate: fillDate, Volume: 12.5, Mileage: 45000,
	}
	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}

	t.Run("deletes fuel record and redirects", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestFuelPageHandler(t, vehicleStub, fuelStub)

		req := httptest.NewRequest(http.MethodPost, "/fuel/f1/delete", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelDelete(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Contains(t, rec.Header().Get("Location"), "/fuel?vehicle=v1")
	})

	t.Run("returns 404 when record not found", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getErr: models.NewNotFoundError("FuelRecord", "f99")}
		handler := newTestFuelPageHandler(t, &stubVehicleSvc{}, fuelStub)

		req := httptest.NewRequest(http.MethodPost, "/fuel/f99/delete", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "f99"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelDelete(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 403 when vehicle belongs to another user", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		handler := newTestFuelPageHandler(t, vehicleStub, fuelStub)

		req := httptest.NewRequest(http.MethodPost, "/fuel/f1/delete", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelDelete(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns 500 when delete fails", func(t *testing.T) {
		fuelStub := &stubFuelSvc{
			getResult: baseRecord,
			deleteErr: models.NewDatabaseError("delete fuel", assert.AnError),
		}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestFuelPageHandler(t, vehicleStub, fuelStub)

		req := httptest.NewRequest(http.MethodPost, "/fuel/f1/delete", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelDelete(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPageHandler_FuelNew(t *testing.T) {
	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}

	t.Run("renders fuel new page for authenticated user", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{baseVehicle}}
		handler := newTestFuelPageHandler(t, vehicleStub, &stubFuelSvc{})

		req := httptest.NewRequest(http.MethodGet, "/fuel/new", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelNew(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Add Fuel Record")
		assert.Contains(t, body, "Ford Focus")
	})

	t.Run("renders empty state with no vehicles", func(t *testing.T) {
		vehicleStub := &stubVehicleSvc{listResult: []*models.Vehicle{}}
		handler := newTestFuelPageHandler(t, vehicleStub, &stubFuelSvc{})

		req := httptest.NewRequest(http.MethodGet, "/fuel/new", nil)
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelNew(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "No Vehicles Found")
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestFuelPageHandler(t, &stubVehicleSvc{}, &stubFuelSvc{})

		req := httptest.NewRequest(http.MethodGet, "/fuel/new", nil)
		rec := httptest.NewRecorder()

		handler.FuelNew(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestPageHandler_FuelEdit(t *testing.T) {
	fillDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	baseRecord := &models.FuelRecord{
		ID: "f1", VehicleID: "v1", FuelType: "gasoline",
		FillDate: fillDate, Volume: 12.5, Mileage: 45000,
	}
	baseVehicle := &models.Vehicle{
		ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020,
		Status: models.VehicleStatusActive,
	}

	t.Run("renders fuel edit page for authenticated user", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{getResult: baseVehicle}
		handler := newTestFuelPageHandler(t, vehicleStub, fuelStub)

		req := httptest.NewRequest(http.MethodGet, "/fuel/f1/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelEdit(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "Edit Fuel Record")
		assert.Contains(t, body, "2020 Ford Focus")
	})

	t.Run("returns 404 when record not found", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getErr: models.NewNotFoundError("FuelRecord", "f99")}
		handler := newTestFuelPageHandler(t, &stubVehicleSvc{}, fuelStub)

		req := httptest.NewRequest(http.MethodGet, "/fuel/f99/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "f99"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelEdit(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 403 when vehicle belongs to another user", func(t *testing.T) {
		fuelStub := &stubFuelSvc{getResult: baseRecord}
		vehicleStub := &stubVehicleSvc{
			getResult: &models.Vehicle{ID: "v1", UserID: "other-user", Make: "Ford", Model: "Focus", Year: 2020, Status: models.VehicleStatusActive},
		}
		handler := newTestFuelPageHandler(t, vehicleStub, fuelStub)

		req := httptest.NewRequest(http.MethodGet, "/fuel/f1/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		req = addAuthContext(req, "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.FuelEdit(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns 500 when account missing from context", func(t *testing.T) {
		handler := newTestFuelPageHandler(t, &stubVehicleSvc{}, &stubFuelSvc{})

		req := httptest.NewRequest(http.MethodGet, "/fuel/f1/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "f1"})
		rec := httptest.NewRecorder()

		handler.FuelEdit(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
