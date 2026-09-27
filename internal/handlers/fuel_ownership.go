package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

func (h *FuelAPIHandler) getOwnedFuelRecord(w http.ResponseWriter, r *http.Request) (*models.FuelRecord, bool) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, http.StatusUnauthorized, "AUTHENTICATION_ERROR", "Not authenticated")
		return nil, false
	}

	targetID := mux.Vars(r)["id"]
	rec, lookupErr := h.svc.GetFuel(ctx, targetID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return nil, false
	}

	if ok := h.ensureVehicleOwnership(w, r, rec.VehicleID, caller.ID); !ok {
		return nil, false
	}
	return rec, true
}

func (h *FuelAPIHandler) ensureVehicleOwnership(w http.ResponseWriter, r *http.Request, vehicleID, userID string) bool {
	vehicle, lookupErr := h.vehicleSvc.GetVehicle(r.Context(), vehicleID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return false
	}
	if vehicle.UserID != userID {
		respondWithProblem(w, http.StatusForbidden, "FORBIDDEN", "Not your vehicle")
		return false
	}
	return true
}
