package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/repositories"
	"github.com/truggeri/go-garage/internal/services"
)

type fuelAPIHandler struct {
	svc        services.FuelService
	vehicleSvc services.VehicleService
}

// MakeFuelAPIHandler creates a new fuel API handler.
func MakeFuelAPIHandler(svc services.FuelService, vehicleSvc services.VehicleService) *fuelAPIHandler {
	return &fuelAPIHandler{svc: svc, vehicleSvc: vehicleSvc}
}

// ListAll handles GET /api/v1/vehicles/{vehicleId}/fuel
func (h *fuelAPIHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, http.StatusUnauthorized, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	vehicleID := mux.Vars(r)["vehicleId"]
	if ok := h.ensureVehicleOwnership(w, r, vehicleID, caller.ID); !ok {
		return
	}

	pageIdx, pageLen := extractPaging(r)
	offsetVal := (pageIdx - 1) * pageLen
	filterSpec := repositories.FuelFilters{VehicleID: &vehicleID}

	totalCount, countErr := h.svc.CountFuel(ctx, filterSpec)
	if countErr != nil {
		respondWithProblem(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to count fuel records")
		return
	}

	records, fetchErr := h.svc.ListFuel(ctx, filterSpec, repositories.PaginationParams{
		Limit: pageLen, Offset: offsetVal,
	})
	if fetchErr != nil {
		respondWithProblem(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve fuel records")
		return
	}

	respondWithPayload(w, http.StatusOK, buildFuelListPayload(records, pageIdx, pageLen, totalCount))
}

// CreateOne handles POST /api/v1/vehicles/{vehicleId}/fuel
func (h *fuelAPIHandler) CreateOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, http.StatusUnauthorized, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	vehicleID := mux.Vars(r)["vehicleId"]
	if ok := h.ensureVehicleOwnership(w, r, vehicleID, caller.ID); !ok {
		return
	}

	inputData, parseErr := parseJSONBody(r)
	if parseErr != nil {
		respondWithProblem(w, http.StatusBadRequest, "INVALID_REQUEST", "Bad JSON")
		return
	}

	valErrs := validateRequiredKeys(inputData, "fill_date", "mileage")
	if len(valErrs) > 0 {
		respondWithValidationProblems(w, "Missing fields", valErrs)
		return
	}

	newRec, buildErr := buildNewFuelRecord(inputData, vehicleID)
	if buildErr != nil {
		handleDomainError(w, buildErr)
		return
	}

	if svcErr := h.svc.CreateFuel(ctx, newRec); svcErr != nil {
		handleDomainError(w, svcErr)
		return
	}

	respondWithPayload(w, http.StatusCreated, buildFuelSinglePayload(newRec, "Fuel record created successfully"))
}

// GetOne handles GET /api/v1/fuel/{id}
func (h *fuelAPIHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	rec, ok := h.getOwnedFuelRecord(w, r)
	if !ok {
		return
	}

	respondWithPayload(w, http.StatusOK, buildFuelSinglePayload(rec, ""))
}

// ReplaceOne handles PUT /api/v1/fuel/{id}
func (h *fuelAPIHandler) ReplaceOne(w http.ResponseWriter, r *http.Request) {
	rec, ok := h.getOwnedFuelRecord(w, r)
	if !ok {
		return
	}

	inputData, parseErr := parseJSONBody(r)
	if parseErr != nil {
		respondWithProblem(w, http.StatusBadRequest, "INVALID_REQUEST", "Bad JSON")
		return
	}

	changes, chErr := extractFuelChanges(inputData, rec)
	if chErr != nil {
		handleDomainError(w, chErr)
		return
	}

	updated, updateErr := h.svc.UpdateFuel(r.Context(), rec.ID, changes)
	if updateErr != nil {
		handleDomainError(w, updateErr)
		return
	}

	respondWithPayload(w, http.StatusOK, buildFuelSinglePayload(updated, "Fuel record updated successfully"))
}

// RemoveOne handles DELETE /api/v1/fuel/{id}
func (h *fuelAPIHandler) RemoveOne(w http.ResponseWriter, r *http.Request) {
	rec, ok := h.getOwnedFuelRecord(w, r)
	if !ok {
		return
	}

	if delErr := h.svc.DeleteFuel(r.Context(), rec.ID); delErr != nil {
		handleDomainError(w, delErr)
		return
	}

	respondWithPayload(w, http.StatusOK, map[string]interface{}{
		keySuccess: true, "message": "Fuel record deleted successfully",
	})
}
