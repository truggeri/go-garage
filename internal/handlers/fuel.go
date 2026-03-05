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
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	vehicleID := mux.Vars(r)["vehicleId"]
	vehicle, lookupErr := h.vehicleSvc.GetVehicle(ctx, vehicleID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	if vehicle.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	pageIdx, pageLen := extractPaging(r)
	offsetVal := (pageIdx - 1) * pageLen
	filterSpec := repositories.FuelFilters{VehicleID: &vehicleID}

	totalCount, countErr := h.svc.CountFuelRecords(ctx, filterSpec)
	if countErr != nil {
		respondWithProblem(w, 500, "INTERNAL_ERROR", "Failed counting")
		return
	}

	records, fetchErr := h.svc.ListFuelRecords(ctx, filterSpec, repositories.PaginationParams{
		Limit: pageLen, Offset: offsetVal,
	})
	if fetchErr != nil {
		respondWithProblem(w, 500, "INTERNAL_ERROR", "Failed fetching")
		return
	}

	respondWithPayload(w, 200, buildFuelListPayload(records, pageIdx, pageLen, totalCount))
}

// CreateOne handles POST /api/v1/vehicles/{vehicleId}/fuel
func (h *fuelAPIHandler) CreateOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	vehicleID := mux.Vars(r)["vehicleId"]
	vehicle, lookupErr := h.vehicleSvc.GetVehicle(ctx, vehicleID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	if vehicle.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	inputData, parseErr := parseJSONBody(r)
	if parseErr != nil {
		respondWithProblem(w, 400, "INVALID_REQUEST", "Bad JSON")
		return
	}

	valErrs := validateRequiredKeys(inputData, "fill_date", "odometer", "cost_per_unit", "volume")
	if len(valErrs) > 0 {
		respondWithValidationProblems(w, "Missing fields", valErrs)
		return
	}

	newRec, buildErr := buildNewFuelRecord(inputData, vehicleID)
	if buildErr != nil {
		handleDomainError(w, buildErr)
		return
	}

	if svcErr := h.svc.CreateFuelRecord(ctx, newRec); svcErr != nil {
		handleDomainError(w, svcErr)
		return
	}

	respondWithPayload(w, 201, buildFuelSinglePayload(newRec, "Fuel record created successfully"))
}

// GetOne handles GET /api/v1/fuel/{id}
func (h *fuelAPIHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	rec, lookupErr := h.svc.GetFuelRecord(ctx, targetID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	vehicle, vErr := h.vehicleSvc.GetVehicle(ctx, rec.VehicleID)
	if vErr != nil {
		handleDomainError(w, vErr)
		return
	}

	if vehicle.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	respondWithPayload(w, 200, buildFuelSinglePayload(rec, ""))
}

// ReplaceOne handles PUT /api/v1/fuel/{id}
func (h *fuelAPIHandler) ReplaceOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	existing, lookupErr := h.svc.GetFuelRecord(ctx, targetID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	vehicle, vErr := h.vehicleSvc.GetVehicle(ctx, existing.VehicleID)
	if vErr != nil {
		handleDomainError(w, vErr)
		return
	}

	if vehicle.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	inputData, parseErr := parseJSONBody(r)
	if parseErr != nil {
		respondWithProblem(w, 400, "INVALID_REQUEST", "Bad JSON")
		return
	}

	changes := extractFuelChanges(inputData)

	updated, updateErr := h.svc.UpdateFuelRecord(ctx, targetID, changes)
	if updateErr != nil {
		handleDomainError(w, updateErr)
		return
	}

	respondWithPayload(w, 200, buildFuelSinglePayload(updated, "Fuel record updated successfully"))
}

// RemoveOne handles DELETE /api/v1/fuel/{id}
func (h *fuelAPIHandler) RemoveOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	existing, lookupErr := h.svc.GetFuelRecord(ctx, targetID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	vehicle, vErr := h.vehicleSvc.GetVehicle(ctx, existing.VehicleID)
	if vErr != nil {
		handleDomainError(w, vErr)
		return
	}

	if vehicle.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	if delErr := h.svc.DeleteFuelRecord(ctx, targetID); delErr != nil {
		handleDomainError(w, delErr)
		return
	}

	respondWithPayload(w, 200, map[string]interface{}{
		"success": true, "message": "Fuel record deleted successfully",
	})
}
