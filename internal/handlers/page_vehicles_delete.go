package handlers

import (
	"net/http"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// VehicleDelete handles the vehicle deletion (POST /vehicles/{id}/delete).
func (h *PageHandler) VehicleDelete(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resource, ok := middleware.GetLoadedResourceFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	vehicle, ok := resource.(*models.Vehicle)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if vehicle.UserID != account.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.vehicleService.DeleteVehicle(r.Context(), vehicle.ID); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/vehicles", http.StatusSeeOther)
}
