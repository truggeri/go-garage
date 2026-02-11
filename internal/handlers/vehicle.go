package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/repositories"
	"github.com/truggeri/go-garage/internal/services"
)

type vehicleAPIHandler struct {
	svc services.VehicleService
}

func MakeVehicleAPIHandler(svc services.VehicleService) *vehicleAPIHandler {
	return &vehicleAPIHandler{svc: svc}
}

func (h *vehicleAPIHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	pageIdx, pageLen := extractPaging(r)
	filterSpec := buildVehicleFilterSpec(r, caller.ID)
	offsetVal := (pageIdx - 1) * pageLen

	totalCount, countErr := h.svc.CountVehicles(ctx, filterSpec)
	if countErr != nil {
		respondWithProblem(w, 500, "INTERNAL_ERROR", "Failed counting")
		return
	}

	records, fetchErr := h.svc.ListVehicles(ctx, filterSpec, repositories.PaginationParams{
		Limit: pageLen, Offset: offsetVal,
	})
	if fetchErr != nil {
		respondWithProblem(w, 500, "INTERNAL_ERROR", "Failed fetching")
		return
	}

	respondWithPayload(w, 200, buildListPayload(records, pageIdx, pageLen, totalCount))
}

func (h *vehicleAPIHandler) CreateOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	inputData, parseErr := parseJSONBody(r)
	if parseErr != nil {
		respondWithProblem(w, 400, "INVALID_REQUEST", "Bad JSON")
		return
	}

	valErrs := validateRequiredKeys(inputData, "vin", "make", "model", "year")
	if len(valErrs) > 0 {
		respondWithValidationProblems(w, "Missing fields", valErrs)
		return
	}

	newRec, buildErr := buildNewVehicleRecord(inputData, caller.ID)
	if buildErr != nil {
		respondWithProblem(w, 400, "VALIDATION_ERROR", buildErr.Error())
		return
	}

	if svcErr := h.svc.CreateVehicle(ctx, newRec); svcErr != nil {
		handleDomainError(w, svcErr)
		return
	}

	respondWithPayload(w, 201, buildSinglePayload(newRec, "Vehicle created successfully"))
}

func (h *vehicleAPIHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	rec, lookupErr := h.svc.GetVehicle(ctx, targetID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	if rec.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	respondWithPayload(w, 200, buildSinglePayload(rec, ""))
}

func (h *vehicleAPIHandler) ReplaceOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	existing, lookupErr := h.svc.GetVehicle(ctx, targetID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	if existing.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	inputData, parseErr := parseJSONBody(r)
	if parseErr != nil {
		respondWithProblem(w, 400, "INVALID_REQUEST", "Bad JSON")
		return
	}

	changes := extractVehicleChanges(inputData)
	updated, updateErr := h.svc.UpdateVehicle(ctx, targetID, changes)
	if updateErr != nil {
		handleDomainError(w, updateErr)
		return
	}

	respondWithPayload(w, 200, buildSinglePayload(updated, "Vehicle updated successfully"))
}

func (h *vehicleAPIHandler) RemoveOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	existing, lookupErr := h.svc.GetVehicle(ctx, targetID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	if existing.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	if delErr := h.svc.DeleteVehicle(ctx, targetID); delErr != nil {
		handleDomainError(w, delErr)
		return
	}

	respondWithPayload(w, 200, map[string]interface{}{
		"success": true, "message": "Vehicle deleted successfully",
	})
}

func (h *vehicleAPIHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	existing, lookupErr := h.svc.GetVehicle(ctx, targetID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	if existing.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	respondWithPayload(w, 200, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"vehicle_id": targetID, "total_maintenance_cost": 0.0,
			"maintenance_count": 0, "total_fuel_cost": 0.0,
			"fuel_fill_count": 0, "average_fuel_price": 0.0,
			"last_maintenance_date": nil, "last_fuel_date": nil,
		},
	})
}
