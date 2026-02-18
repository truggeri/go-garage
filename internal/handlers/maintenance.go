package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/repositories"
	"github.com/truggeri/go-garage/internal/services"
)

type maintenanceAPIHandler struct {
	svc        services.MaintenanceService
	vehicleSvc services.VehicleService
}

// MakeMaintenanceAPIHandler creates a new maintenance API handler.
func MakeMaintenanceAPIHandler(svc services.MaintenanceService, vehicleSvc services.VehicleService) *maintenanceAPIHandler {
	return &maintenanceAPIHandler{svc: svc, vehicleSvc: vehicleSvc}
}

// ListAll handles GET /api/v1/vehicles/{vehicleId}/maintenance
func (h *maintenanceAPIHandler) ListAll(w http.ResponseWriter, r *http.Request) {
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
	filterSpec := repositories.MaintenanceFilters{VehicleID: &vehicleID}

	totalCount, countErr := h.svc.CountMaintenance(ctx, filterSpec)
	if countErr != nil {
		respondWithProblem(w, 500, "INTERNAL_ERROR", "Failed counting")
		return
	}

	records, fetchErr := h.svc.ListMaintenance(ctx, filterSpec, repositories.PaginationParams{
		Limit: pageLen, Offset: offsetVal,
	})
	if fetchErr != nil {
		respondWithProblem(w, 500, "INTERNAL_ERROR", "Failed fetching")
		return
	}

	respondWithPayload(w, 200, buildMaintenanceListPayload(records, pageIdx, pageLen, totalCount))
}

// CreateOne handles POST /api/v1/vehicles/{vehicleId}/maintenance
func (h *maintenanceAPIHandler) CreateOne(w http.ResponseWriter, r *http.Request) {
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

	valErrs := validateRequiredKeys(inputData, "service_type", "service_date")
	if len(valErrs) > 0 {
		respondWithValidationProblems(w, "Missing fields", valErrs)
		return
	}

	newRec, buildErr := buildNewMaintenanceRecord(inputData, vehicleID)
	if buildErr != nil {
		handleDomainError(w, buildErr)
		return
	}

	if svcErr := h.svc.CreateMaintenance(ctx, newRec); svcErr != nil {
		handleDomainError(w, svcErr)
		return
	}

	respondWithPayload(w, 201, buildMaintenanceSinglePayload(newRec, "Maintenance record created successfully"))
}

// GetOne handles GET /api/v1/maintenance/{id}
func (h *maintenanceAPIHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	rec, lookupErr := h.svc.GetMaintenance(ctx, targetID)
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

	respondWithPayload(w, 200, buildMaintenanceSinglePayload(rec, ""))
}

// ReplaceOne handles PUT /api/v1/maintenance/{id}
func (h *maintenanceAPIHandler) ReplaceOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	existing, lookupErr := h.svc.GetMaintenance(ctx, targetID)
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

	changes, chErr := extractMaintenanceChanges(inputData)
	if chErr != nil {
		handleDomainError(w, chErr)
		return
	}

	updated, updateErr := h.svc.UpdateMaintenance(ctx, targetID, changes)
	if updateErr != nil {
		handleDomainError(w, updateErr)
		return
	}

	respondWithPayload(w, 200, buildMaintenanceSinglePayload(updated, "Maintenance record updated successfully"))
}

// RemoveOne handles DELETE /api/v1/maintenance/{id}
func (h *maintenanceAPIHandler) RemoveOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	existing, lookupErr := h.svc.GetMaintenance(ctx, targetID)
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

	if delErr := h.svc.DeleteMaintenance(ctx, targetID); delErr != nil {
		handleDomainError(w, delErr)
		return
	}

	respondWithPayload(w, 200, map[string]interface{}{
		"success": true, "message": "Maintenance record deleted successfully",
	})
}
